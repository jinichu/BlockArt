package blockartlib

import (
	"crypto/ecdsa"
	"encoding/json"
	"math"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	crypto "../crypto"
)

type ArtNode struct {
	client  *rpc.Client      // RPC client to connect to the InkMiner
	privKey ecdsa.PrivateKey // Pub/priv key pair of this ArtNode
}

// Adds a new shape to the canvas.
// Can return the following errors:
// - DisconnectedError
// - InsufficientInkError
// - InvalidShapeSvgStringError
// - ShapeSvgStringTooLongError
func (a *ArtNode) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	inkCost, err := calculateInkCost(shapeSvgString, fill, stroke)
	if err != nil {
		return "", "", 0, err
	}

	publicKey, err := crypto.MarshalPublic(&a.privKey.PublicKey)
	if err != nil {
		return "", "", 0, err
	}

	shape := Shape{
		Svg:    shapeSvgString,
		Fill:   fill,
		Stroke: stroke,
	}

	err = svgStringValidityCheck(shapeSvgString)
	if err != nil {
		return "", "", 0, err
	}

	id := time.Now().String()

	args := Operation{
		OpType:      ADD,
		Shape:       shape,
		OpSig:       OpSig{},
		PubKey:      publicKey,
		InkCost:     inkCost,
		ValidateNum: validateNum,
		Id:          id,
	}

	shapeHash, err = crypto.Hash(args)
	if err != nil {
		return "", "", 0, err
	}

	args.ShapeHash = shapeHash

	bytes, err := json.Marshal(args)
	if err != nil {
		return "", "", 0, err
	}

	r, s, err := crypto.Sign(bytes, a.privKey)
	if err != nil {
		return "", "", 0, err
	}

	args.OpSig = OpSig{r, s}

	var resp AddShapeResponse
	err = a.client.Call("InkMinerRPC.AddShape", args, &resp)
	//TODO: retrieve blockHash, inkRemaining from call to ink miner to add shape

	if err != nil {
		return "", "", 0, err
	}

	return shapeHash, resp.BlockHash, resp.InkRemaining, nil
}

// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (a *ArtNode) GetSvgString(shapeHash string) (svgString string, err error) {
	var resp string

	err = a.client.Call("InkMinerRPC.GetSvgString", shapeHash, &resp)
	if err != nil {
		return "", err
	}

	svgString = resp
	return
}

// Returns the amount of ink currently available.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetInk() (inkRemaining uint32, err error) {
	var resp uint32

	err = a.client.Call("InkMinerRPC.GetInk", "", &resp)
	if err != nil {
		return 0, err
	}

	inkRemaining = resp
	return
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
// - OutOfBoundsError
// - ShapeOverlapError
func (a *ArtNode) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	publicKey, err := crypto.MarshalPublic(&a.privKey.PublicKey)
	if err != nil {
		return 0, err
	}

	id := time.Now().String()

	args := Operation{
		OpType:      DELETE,
		OpSig:       OpSig{},
		PubKey:      publicKey,
		ShapeHash:   shapeHash,
		ValidateNum: validateNum,
		Id:          id,
	}

	bytes, err := json.Marshal(args)
	if err != nil {
		return 0, err
	}

	r, s, err := crypto.Sign(bytes, a.privKey)
	if err != nil {
		return 0, err
	}

	args.OpSig = OpSig{r, s}

	var resp uint32

	err = a.client.Call("InkMinerRPC.DeleteShape", args, &resp)
	if err != nil {
		return 0, err
	}

	inkRemaining = resp
	return
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetShapes(blockHash string) (shapeHashes []string, err error) {
	var resp GetShapesResponse

	err = a.client.Call("InkMinerRPC.GetShapes", blockHash, &resp)
	if err != nil {
		return nil, err
	}

	shapeHashes = resp.ShapeHashes
	return
}

// Returns the block hash of the genesis block.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetGenesisBlock() (blockHash string, err error) {
	var resp string

	err = a.client.Call("InkMinerRPC.GetGenesisBlock", "", &resp)
	if err != nil {
		return "", err
	}

	blockHash = resp
	return
}

// Retrieves the children blocks of the block identified by blockHash.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetChildren(blockHash string) (blockHashes []string, err error) {
	var resp GetChildrenResponse

	err = a.client.Call("InkMinerRPC.GetChildrenBlocks", blockHash, &resp)
	if err != nil {
		return nil, err
	}

	blockHashes = resp.BlockHashes
	return
}

// Closes the canvas/connection to the BlockArt network.
// - DisconnectedError
func (a *ArtNode) CloseCanvas() (inkRemaining uint32, err error) {
	var resp uint32

	err = a.client.Call("InkMinerRPC.GetInk", "", &resp)
	if err != nil {
		return 0, err
	}

	inkRemaining = resp
	return
}

// HELPERS

// Gets the ink cost of a particular operation
// Can return the following errors:
// -InvalidShapeSvgStringError
func calculateInkCost(shapeSvgString string, fill string, stroke string) (cost uint32, err error) {
	if fill == "transparent" {
		cost = uint32(calculateLineCost(shapeSvgString))
	} else {
		if !isClosed(shapeSvgString) {
			return 0, InvalidShapeSvgStringError(shapeSvgString)
		} else if isSelfIntersecting(shapeSvgString) {
			return 0, InvalidShapeSvgStringError(shapeSvgString)
		}
		cost = uint32(calculateFillCost(shapeSvgString))
	}

	return cost, nil
}

// Check if the svg string is a closed-form shape
func isClosed(shapeSvgString string) bool {
	arr := strings.Fields(shapeSvgString)

	res := false

	if arr[len(arr)-1] == "Z" || arr[len(arr)-1] == "Z" {
		res = true
	} else if arr[len(arr)-3] == "L" && arr[len(arr)-2] == arr[1] && arr[len(arr)-1] == arr[2] {
		res = true
	} else if arr[len(arr)-3] == "l" {
		//TODO
		print(res)
	}
	return res
}

// Check if the svg string is self intersecting
func isSelfIntersecting(shapeSvgString string) bool {
	//TODO
	return false
}

// Calculate the ink cost to fill a shape
func calculateFillCost(shapeSvgString string) (cost float64) {
	i := 0
	cost = 0.0
	arr := strings.Fields(shapeSvgString)

	var operation string
	var current_pos []float64
	var original_pos []float64
	var new_pos []float64
	var vertices [][]float64
	var x float64
	var y float64

	for {
		if i >= len(arr) {
			break
		}
		operation = arr[i]

		if operation == "M" || operation == "m" || operation == "L" || operation == "l" {
			x = parseFloat(arr[i+1])
			y = parseFloat(arr[i+2])
		} else if operation == "V" || operation == "v" {
			y = parseFloat(arr[i+1])
		} else if operation == "H" || operation == "h" {
			x = parseFloat(arr[i+1])
		}

		new_pos = []float64{}

		switch operation {
		case "M":
			new_pos = []float64{x, y}
			if original_pos == nil {
				original_pos = new_pos
			}
			i += 3
		case "m":
			new_pos = []float64{x + current_pos[0], y + current_pos[1]}
			i += 3
		case "L":
			new_pos = []float64{x, y}
			i += 3
		case "l":
			new_pos = []float64{x + current_pos[0], y + current_pos[1]}
			i += 3
		case "H":
			new_pos = []float64{x, current_pos[1]}
			i += 2
		case "h":
			new_pos = []float64{x + current_pos[0], current_pos[1]}
			i += 2
		case "V":
			new_pos = []float64{current_pos[0], y}
			i += 2
		case "v":
			new_pos = []float64{current_pos[0], y + current_pos[1]}
			i += 2
		case "Z", "z":
			new_pos = original_pos
			i += 1
		default:
		}
		current_pos = new_pos
		vertices = append(vertices, new_pos)
	}
	cost = calculateArea(vertices)
	return
}

// Calculate the cost to draw a line
func calculateLineCost(shapeSvgString string) (cost float64) {
	i := 0
	cost = 0.0
	arr := strings.Fields(shapeSvgString)

	var operation string
	var current_pos []float64
	var original_pos []float64
	var new_pos []float64
	var x float64
	var y float64

	for {
		if i >= len(arr) {
			return cost
		}

		operation = arr[i]

		if operation == "M" || operation == "m" || operation == "L" || operation == "l" {
			x = parseFloat(arr[i+1])
			y = parseFloat(arr[i+2])
		} else if operation == "V" || operation == "v" {
			y = parseFloat(arr[i+1])
		} else if operation == "H" || operation == "h" {
			x = parseFloat(arr[i+1])
		}

		switch operation {
		case "M":
			new_pos = []float64{x, y}
			if original_pos == nil {
				original_pos = new_pos
			}
			i += 3
		case "m":
			new_pos = []float64{x + current_pos[0], y + current_pos[1]}
			i += 3
		case "L":
			new_pos = []float64{x, y}
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 3
		case "l":
			new_pos = []float64{x + current_pos[0], y + current_pos[1]}
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 3
		case "H":
			new_pos = []float64{x, current_pos[1]}
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 2
		case "h":
			new_pos = []float64{x + current_pos[0], current_pos[1]}
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 2
		case "V":
			new_pos = []float64{current_pos[0], y}
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 2
		case "v":
			new_pos = []float64{current_pos[0], y + current_pos[1]}
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 2
		case "Z", "z":
			new_pos = original_pos
			cost += calculateDistance(current_pos[0], new_pos[0], current_pos[1], new_pos[1])
			i += 1
		default:
		}
		current_pos = new_pos
	}
	return cost
}

// Calculate the diwtance between two points
func calculateDistance(x0 float64, x1 float64, y0 float64, y1 float64) (distance float64) {
	return math.Sqrt(math.Pow((x1-x0), 2) + math.Pow((y1-y0), 2))
}

// Parse a string number to a float
func parseFloat(s string) (f float64) {
	f, _ = strconv.ParseFloat(s, 64)
	return
}

// Checks if valid svg string
// - InvalidShapeSvgString Error
// - ShapeSvgStringTooLong Error
func svgStringValidityCheck(svgString string) (err error) {
	if len(svgString) > 128 {
		return ShapeSvgStringTooLongError(svgString)
	}

	if !isValidPath(svgString) {
		return InvalidShapeSvgStringError(svgString)
	}

	return nil
}

// Checks if a path is valid
func isValidPath(svgString string) bool {
	i := 0
	var operation string
	arr := strings.Fields(svgString)

	if len(arr) == 0 {
		return false
	}

	if arr[0] != "M" {
		return false
	}

	for {
		if i >= len(arr) {
			return true
		}
		operation = arr[i]

		switch operation {
		case "M", "m", "L", "l":
			if i+1 >= len(arr) || i+2 >= len(arr) {
				return false
			}

			if !isNumeric(arr[i+1]) || !isNumeric(arr[i+2]) {
				return false
			}

			i += 3
		case "H", "h", "V", "v":
			if i+1 >= len(arr) {
				return false
			}

			if !isNumeric(arr[i+1]) {
				return false
			}

			i += 2
		case "Z", "z":
			if i+1 >= len(arr) {
				return true
			}

			if isNumeric(arr[i+1]) {
				return false
			}

			i += 1
		default:
			return false
		}
	}
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// Calculate the area of a shape given a list of vertices
func calculateArea(vertices [][]float64) (area float64) {
	n := len(vertices)
	area = 0.0

	for i, _ := range vertices {
		j := (i + 1) % n
		area += vertices[i][0] * vertices[j][1]
		area -= vertices[j][0] * vertices[i][1]
	}

	return math.Abs(area) / 2
}
