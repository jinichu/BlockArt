package blockartlib

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	"net/rpc"
	"strconv"
	"strings"
	"time"
)

type ArtNode struct {
	client    *rpc.Client      // RPC client to connect to the InkMiner
	privKey   ecdsa.PrivateKey // Pub/priv key pair of this ArtNode
	minerAddr string
}

type Point struct {
	x float64
	y float64
}

type Vector struct {
	point0 Point
	point1 Point
}

// Adds a new shape to the canvas.
// Can return the following errors:
// - DisconnectedError
// - InsufficientInkError
// - InvalidShapeSvgStringError
// - ShapeSvgStringTooLongError
func (a *ArtNode) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	shape := Shape{
		Type:   shapeType,
		Svg:    shapeSvgString,
		Fill:   fill,
		Stroke: stroke,
	}

	err = svgStringValidityCheck(shapeSvgString)
	if err != nil {
		return "", "", 0, err
	}

	err = svgShapeValidityCheck(shapeSvgString, fill, stroke)
	if err != nil {
		return "", "", 0, err
	}

	args := Operation{
		OpType:      ADD,
		OpSig:       OpSig{},
		PubKey:      a.privKey.PublicKey,
		ValidateNum: validateNum,
		Id:          time.Now().Unix(),
	}

	args.ADD.Shape = shape

	args, err = args.Sign(a.privKey)
	if err != nil {
		return "", "", 0, fmt.Errorf("signing error: %+v", err)
	}

	shapeHash, err = args.Hash()
	if err != nil {
		return "", "", 0, err
	}

	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return "", "", 0, DisconnectedError(a.minerAddr)
	}

	var resp AddShapeResponse
	if err = a.client.Call("InkMinerRPC.AddShape", args, &resp); err != nil {
		return "", "", 0, err
	}

	return shapeHash, resp.BlockHash, resp.InkRemaining, nil
}

// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (a *ArtNode) GetSvgString(shapeHash string) (svgString string, err error) {
	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return "", DisconnectedError(a.minerAddr)
	}

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
	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return 0, DisconnectedError(a.minerAddr)
	}

	var resp uint32

	err = a.client.Call("InkMinerRPC.GetInk", "", &resp)
	if err != nil {
		return 0, err
	}

	return resp, nil
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
// - OutOfBoundsError
// - ShapeOverlapError
func (a *ArtNode) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	args := Operation{
		OpType:      DELETE,
		OpSig:       OpSig{},
		PubKey:      a.privKey.PublicKey,
		ValidateNum: validateNum,
		Id:          time.Now().Unix(),
	}
	args.DELETE.ShapeHash = shapeHash

	args, err = args.Sign(a.privKey)
	if err != nil {
		return 0, err
	}

	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return 0, DisconnectedError(a.minerAddr)
	}

	var resp uint32

	err = a.client.Call("InkMinerRPC.DeleteShape", args, &resp)
	if err != nil {
		return 0, err
	}

	return resp, nil
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetShapes(blockHash string) (shapeHashes []string, err error) {
	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return nil, DisconnectedError(a.minerAddr)
	}

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
	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return "", DisconnectedError(a.minerAddr)
	}

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
	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return nil, DisconnectedError(a.minerAddr)
	}

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
	// Simple RPC call to check if we can reach the InkMiner
	var req string
	var success bool
	err = a.client.Call("InkMinerRPC.TestConnection", req, &success)
	if err != nil {
		return 0, DisconnectedError(a.minerAddr)
	}

	var resp uint32

	err = a.client.Call("InkMinerRPC.GetInk", "", &resp)
	if err != nil {
		return 0, err
	}

	if err := a.client.Close(); err != nil {
		return 0, err
	}

	return resp, nil
}


// HELPERS

// Check if the svg string is a closed-form shape
func isClosed(operation string, original_pos Point, current_pos Point) bool {
	isClosed := false

	if operation == "Z" || operation == "z" {
		isClosed = true
	} else if operation == "L" && current_pos.x == original_pos.x && current_pos.y == original_pos.y {
		isClosed = true
	} else if operation == "l" && original_pos.x == current_pos.x && original_pos.y == current_pos.y {
		isClosed = true
	}

	return isClosed
}

func isEqual(point0 Point, point1 Point) bool {
	if point0.x != point1.x {
		return false
	}

	if point0.y != point1.y {
		return false
	}

	return true
}

// Check if the svg string is self intersecting
func isSelfIntersecting(vectors []Vector) bool {
	for index, vector := range(vectors) {
		if (index > 2) {
			other_vectors := vectors[:index]

			for _, other_vector := range(other_vectors) {
				v_point0 := vector.point0
				v_point1 := vector.point1
				ov_point0 := other_vector.point0
				ov_point1 := other_vector.point1

				if (!isEqual(v_point0, ov_point1) && !isEqual(ov_point0, v_point1)) {
					if (isIntersecting(vector.point0,vector.point1,other_vector.point0,other_vector.point1) == true) {
						return true
					}
				}
			}
		}
	}
	return false
}

// Calculate the ink cost to fill a shape
func calculateFillCost(shapeSvgString string) (cost float64, err error) {
	i := 3
	cost = 0.0
	arr := strings.Fields(shapeSvgString)

	var operation string
	var vertices []Point
	var vectors []Vector
	original_pos := Point{parseFloat(arr[1]), parseFloat(arr[2])}
	current_pos := original_pos
	vertices = append(vertices, original_pos)
	new_pos := Point{}
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

		switch operation {
		case "M":
			new_pos = Point{x, y}
			i += 3
		case "m":
			new_pos = Point{x + current_pos.x, y + current_pos.y}
			i += 3
		case "L":
			new_pos = Point{x, y}
			i += 3
		case "l":
			new_pos = Point{x + current_pos.x, y + current_pos.y}
			i += 3
		case "H":
			new_pos = Point{x, current_pos.y}
			i += 2
		case "h":
			new_pos = Point{x + current_pos.x, current_pos.y}
			i += 2
		case "V":
			new_pos = Point{current_pos.x, y}
			i += 2
		case "v":
			new_pos = Point{current_pos.x, y + current_pos.y}
			i += 2
		case "Z", "z":
			new_pos = original_pos
			i += 1
		default:
		}
		current_pos = new_pos
		vertices = append(vertices, new_pos)
	}

	// check to see if Shape is a closed shape
	if isClosed(operation, original_pos, current_pos) == false {
		return 0, InvalidShapeSvgStringError(shapeSvgString)
	}

  	vectors = computeVectors(vertices)
	if isSelfIntersecting(vectors) == true {
		return 0, InvalidShapeSvgStringError(shapeSvgString)
	}

	cost = calculateArea(vertices)
	err = nil
	return
}

func computeVectors(vertices []Point) []Vector {
  var vectors []Vector
  j := 0

  for {
    if j+1 >= len(vertices) {
	break
    } else {
	point0 := vertices[j]
	point1 := vertices[j+1]
	vector := Vector{point0: point0, point1: point1}
	vectors = append(vectors, vector)
	j += 1
    }
  }
  return vectors
}

// Calculate the cost to draw a line
func calculateLineCost(shapeSvgString string) (cost float64) {
	i := 0
	cost = 0.0
	arr := strings.Fields(shapeSvgString)

	var operation string
	current_pos := Point{}
	original_pos := Point{}
	new_pos := Point{}
	var x float64
	var y float64

	originalPosIsInitialized := false

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
			new_pos = Point{x, y}
			if !originalPosIsInitialized {
				original_pos = new_pos
				originalPosIsInitialized = true
			}
			i += 3
		case "m":
			new_pos = Point{x + current_pos.x, y + current_pos.y}
			i += 3
		case "L":
			new_pos = Point{x, y}
			cost += calculateDistance(current_pos, new_pos)
			i += 3
		case "l":
			new_pos = Point{x + current_pos.x, y + current_pos.y}
			cost += calculateDistance(current_pos, new_pos)
			i += 3
		case "H":
			new_pos = Point{x, current_pos.y}
			cost += calculateDistance(current_pos, new_pos)
			i += 2
		case "h":
			new_pos = Point{x + current_pos.x, current_pos.y}
			cost += calculateDistance(current_pos, new_pos)
			i += 2
		case "V":
			new_pos = Point{current_pos.x, y}
			cost += calculateDistance(current_pos, new_pos)
			i += 2
		case "v":
			new_pos = Point{current_pos.x, y + current_pos.y}
			cost += calculateDistance(current_pos, new_pos)
			i += 2
		case "Z", "z":
			new_pos = original_pos
			cost += calculateDistance(current_pos, new_pos)
			i += 1
		default:
		}
		current_pos = new_pos
	}
	return cost
}

// Calculate the diwtance between two points
func calculateDistance(point0 Point, point1 Point) (distance float64) {
	x0 := point0.x
	x1 := point1.x
	y0 := point0.y
	y1 := point1.y
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

func svgShapeValidityCheck(svgString string, fill string, stroke string) (err error) {
	if fill == "transparent" && stroke == "transparent" {
		return InvalidShapeSvgStringError(svgString)
	} else if fill != "transparent" {
		_, err = calculateFillCost(svgString)
		if err != nil {
			return err
		}
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
func calculateArea(vertices []Point) (area float64) {
	n := len(vertices)
	area = 0.0

	for i, _ := range vertices {
		j := (i + 1) % n
		area += vertices[i].x * vertices[j].y
		area -= vertices[j].x * vertices[i].y
	}

	return math.Abs(area) / 2
}

// Compute all the vertices of an SVG string
func ComputeVertices(shapeSvgString string) []Point {
	i := 3
	arr := strings.Fields(shapeSvgString)

	var operation string
	var vertices []Point
	original_pos := Point{parseFloat(arr[1]), parseFloat(arr[2])}
	current_pos := original_pos
	vertices = append(vertices, original_pos)
	new_pos := Point{}
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

		switch operation {
		case "M":
			new_pos = Point{x, y}
			i += 3
		case "m":
			new_pos = Point{x + current_pos.x, y + current_pos.y}
			i += 3
		case "L":
			new_pos = Point{x, y}
			i += 3
		case "l":
			new_pos = Point{x + current_pos.x, y + current_pos.y}
			i += 3
		case "H":
			new_pos = Point{x, current_pos.y}
			i += 2
		case "h":
			new_pos = Point{x + current_pos.x, current_pos.y}
			i += 2
		case "V":
			new_pos = Point{current_pos.x, y}
			i += 2
		case "v":
			new_pos = Point{current_pos.x, y + current_pos.y}
			i += 2
		case "Z", "z":
			new_pos = original_pos
			i += 1
		default:
		}
		current_pos = new_pos
		vertices = append(vertices, new_pos)
	}
	return vertices
}

// Gets the ink cost of a particular operation
// Can return the following errors:
// -InvalidShapeSvgStringError
func (sh Shape) InkCost() (cost uint32, err error) {
	lineCost := calculateLineCost(sh.Svg)

	if sh.Fill == "transparent" && sh.Stroke == "transparent" {
		return 0, InvalidShapeSvgStringError(sh.Svg)
	}

	if sh.Fill == "transparent" && sh.Stroke != "transparent" {
		return uint32(lineCost), nil
	}

	fillCost, err := calculateFillCost(sh.Svg)
	if err != nil {
		return 0, err
	}

	if sh.Stroke != "transparent" {
		return uint32(fillCost + lineCost), nil
	}

	if sh.Stroke == "transparent" {
		return uint32(fillCost), nil
	}
	return
}

func (p *Point) GetX() float64 {
	return p.x
}

func (p *Point) GetY() float64 {
	return p.y
}

func isIntersecting(point0 Point, point1 Point, point2 Point, point3 Point) bool {
	if (isCrossing(point0,point1,point2,point0,point1,point3) == true && isCrossing(point2,point3,point0,point2,point3,point1) == true) {
		return true;
	}

	if isPointOnLine(point0, point2, point1) || isPointOnLine(point0, point3, point1) || isPointOnLine(point2, point0, point3) || isPointOnLine(point2, point1, point3)  {
		return true
	}

	return false;
}

func isPointOnLine(point0 Point, point1 Point, point2 Point) bool {
	if calculateDir(point0,point1,point2) == 0 {
		if  point1.x >= math.Min(point0.x, point2.x) && point1.x <= math.Max(point0.x, point2.x) && point1.y <= math.Max(point0.y, point2.y) && math.Min(point0.y, point2.y) <= point1.y {
		  return true
		}
		
	}
	return false
}

func isCrossing(aa Point, ab Point, ac Point, ba Point, bb Point, bc Point) bool {
	calc0 := calculateDir(aa,ab,ac)
	calc1 := calculateDir(ba,bb,bc)

	if calc0 > 0 {
		calc0 = 1
	} else if calc0 < 0 {
		calc0 = -1
	}

	if calc1 > 0 {
		calc1 = 1
	} else if calc1 < 0 {
		calc1 = -1
	}

	if calc0 != calc1 {
		return true
	}
	return false
}

func calculateDir(point0 Point, point1 Point, point2 Point) float64 {
	return (point1.y - point0.y) * (point2.x - point1.x) - (point1.x - point0.x) * (point2.y - point1.y)
}


func pointInPolygon(vectors []Vector, point Point) bool {
	numTouches := 0
	i := len(vectors) - 1

	for {
		if i <= 0 {
			break
		}
		currentVector := vectors[i]
		if (isIntersecting(currentVector.point0, currentVector.point1, point, Point{point.x, 100000})) {
			numTouches += 1
			if (calculateDir(currentVector.point0, point, currentVector.point1) == 0) {
				if isPointOnLine(currentVector.point0, point, currentVector.point1) {
					return true
				} else {
					return false 
				}
			}
		}
		i -= 1
	}

	if numTouches % 2 == 1 {
		return true	
	}
	return false
}

func doesShapeOverlap(sh0 Shape, sh1 Shape) bool {
	vectors0 := computeVectors(ComputeVertices(sh0.Svg))
	vectors1 := computeVectors(ComputeVertices(sh1.Svg))

	isFilled0 := (sh0.Fill != "transparent")
	isFilled1 := (sh1.Fill != "transparent")

	for _, vector := range(vectors0) {
		for _, other_vector := range(vectors1) {
			if (isIntersecting(vector.point0, vector.point1, other_vector.point0, other_vector.point1) == true) {
				return true
			} 
		}
	}


	//if isFilled == true, check if one is perfectly inside another
	if isFilled0 == true && isFilled1 == true {
		for _, vector0 := range(vectors0) {
			if (pointInPolygon(vectors1, vector0.point0) || pointInPolygon(vectors1, vector0.point1)) {
			  return true
			}
		}

		for _, vector1 := range(vectors1) {
			if (pointInPolygon(vectors0, vector1.point0) || pointInPolygon(vectors0, vector1.point1)) {
				return true
			}
		}

	} else if isFilled0 == true && isFilled1 == false {
		for _, vector1 := range(vectors1) {
			if (pointInPolygon(vectors0, vector1.point0) == true || pointInPolygon(vectors0, vector1.point1) == true) {
				return true
			}
		}

	} else if isFilled1 == true && isFilled0 == false {
		for _, vector0 := range(vectors0) {
			if (pointInPolygon(vectors1, vector0.point0) || pointInPolygon(vectors1, vector0.point1)) {
				return true
			}
		}
	}
	return false
}