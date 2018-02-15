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
	    if (isIntersecting(vector, other_vector) == true) {
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
	current_pos := Point{}
	original_pos := Point{parseFloat(arr[1]), parseFloat(arr[2])}
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
	current_pos := Point{}
	original_pos := Point{parseFloat(arr[1]), parseFloat(arr[2])}
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

func isIntersecting(vector0 Vector, vector1 Vector) bool {
	v_point0 := vector0.point0
	v_point1 := vector0.point1
	ov_point0 := vector1.point0
	ov_point1 := vector1.point1
	o1 := direction(vector0, ov_point0);
	o2 := direction(vector0, ov_point1);
	o3 := direction(vector1, v_point0);
	o4 := direction(vector1, v_point1);

	if (o1 != o2 && o3 != o4) {
		return true
	}

	if (o1 == 0 && onSegment(vector0, ov_point0)) {
		return true
	}

	if (o2 == 0 && onSegment(vector0, ov_point1)) {
		return true
	}

	if (o3 == 0 && onSegment(vector1, v_point0)) {
		return true
	}

	if (o4 == 0 && onSegment(vector1, v_point1)) {
		return true
	}
	return false
}

func onSegment(vector Vector, q Point) bool {
	p := vector.point0
	r := vector.point1
	if q.x <= math.Max(p.x, r.x) && q.x >= math.Min(p.x, r.x) && q.y <= math.Max(p.y, r.y) && q.y >= math.Min(p.y, r.y) {
		return true
	}
	return false
}

func direction(vector Vector, r Point) int {
	p := vector.point0
	q := vector.point1
	val := (q.y + 1 - p.y) * (r.x + 1 - q.x) - (q.x + 1 - p.x) * (r.y + 1 - q.y)

	if (val == 0 ){
		return 0
	} else if val > 0 {
		return 1
	} else {
		return 2
	}
}


func pointInPolygon(vectors []Vector, point Point) bool {
	touches := 0
	i := 0

	for {
		if i >= len(vectors) {
			break
		}
		vectorB := Vector{point, Point{point.x, math.Inf(1)}}
		if (isIntersecting(vectors[i], vectorB)) {
			vectorC := Vector{vectors[i].point0, point}
			if (direction(vectorC, vectors[i].point1) == 0) {
				return onSegment(vectors[i], point);
			} else {
				touches += 1
			}
		}
		i += 1
	}
	return touches % 2 == 1;
}

//Check to see if two shapes overlap
func shapesOverlap(svgString0 string, svgString1 string, isFilled bool) bool {
	vectors0 := computeVectors(ComputeVertices(svgString0))
	vectors1 := computeVectors(ComputeVertices(svgString1))

	for _, vector0 := range(vectors0) {
		for _, vector1 := range(vectors1) {
			if (isIntersecting(vector0, vector1) == true) {
			  return true
			}
		}
	}

	//if isFilled == true, check if one is perfectly inside another
	if isFilled == true {
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
	}
	return false
}
