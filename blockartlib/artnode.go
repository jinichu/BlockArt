package blockartlib

import (
	"crypto/ecdsa"
	"net/rpc"
	"time"
	"strings"
	"strconv"
	"math"
	"fmt"
)

type ArtNode struct {
	client  *rpc.Client      // RPC client to connect to the InkMiner
	privKey ecdsa.PrivateKey // Pub/priv key pair of this ArtNode
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

// Gets the ink cost of a particular operation
// Can return the following errors:
// -InvalidShapeSvgStringError
func calculateInkCost(shapeSvgString string, fill string, stroke string) (cost uint32, err error) {
	var floatCost float64
	if fill == "transparent" {
		floatCost = calculateLineCost(shapeSvgString)
	} else {
		floatCost, err = calculateFillCost(shapeSvgString)
		if err != nil {
			return 0, err
		}
	}

	return uint32(floatCost), nil
}

// Check if the svg string is a closed-form shape
func isClosed(operation string, original_pos Point, current_pos Point) bool {
	isClosed := false

	if operation == "Z" || operation == "z" {
		isClosed = true
	} else if operation == "L" && current_pos.x == original_pos.x && current_pos.y == original_pos.y {
		isClosed = true
	} else if operation == "l" && original_pos.x == current_pos.x && original_pos.y == current_pos.y   {
		isClosed = true
	}

	return isClosed
}

func onSegment(p Point, q Point, r Point) bool {
    if (q.x <= math.Max(p.x, r.x) && q.x >= math.Min(p.x, r.x) &&
	  q.y <= math.Max(p.y, r.y) && q.y >= math.Min(p.y, r.y)) {

	 return true
    }
 
    return false
}

func orientation(p Point, q Point, r Point) int {
	val := (q.y + 1 - p.y) * (r.x + 1 - q.x) - (q.x + 1 - p.x) * (r.y + 1 - q.y)
	if (val == 0 ){
		return 0
	}

	if (val > 0) {
		return 1
	} else {
		return 2
	}
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
      var other_vectors []Vector
      other_vectors = vectors[:index]

      for _, other_vector := range(other_vectors) {
        p1 := vector.point0
        q1 := vector.point1
        p2 := other_vector.point0
        q2 := other_vector.point1
        o1 := orientation(p1, q1, p2);
        o2 := orientation(p1, q1, q2);
        o3 := orientation(p2, q2, p1);
        o4 := orientation(p2, q2, q1);

        if (!isEqual(p1, q2) && !isEqual(p2, q1)) {


          if (o1 != o2 && o3 != o4) {
              return true
          }
          
          if (o1 == 0 && onSegment(p1, p2, q1)) {
            return true
         }
          
          if (o2 == 0 && onSegment(p1, q2, q1)) {
            return true
          }
          
          if (o3 == 0 && onSegment(p2, p1, q2)) {
            return true
          }
          
          if (o4 == 0 && onSegment(p2, q1, q2)) {
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
		if (i >= len(arr)) {
			break
		}
		operation = arr[i]

		if operation == "M" || operation == "m" || operation == "L" || operation == "l" {
			x = parseFloat(arr[i + 1])
			y = parseFloat(arr[i + 2])
		} else if operation == "V" || operation == "v"  {
			y = parseFloat(arr[i + 1])
		} else if operation == "H" || operation == "h" {
			x = parseFloat(arr[i + 1])
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

	// check to see if shape is self intersecting
	j := 0

  fmt.Printf("VERTICES: %+v\n", vertices)

	for {
		if j + 1 >= len(vertices) {
			break
		} else {
			point0 := vertices[j]
			point1 := vertices[j + 1]
			vector := Vector{point0: point0, point1: point1}
			vectors = append(vectors, vector)
			j += 1
		}
	}

	if isSelfIntersecting(vectors) == true {
		return 0, InvalidShapeSvgStringError(shapeSvgString)
	}


	cost = calculateArea(vertices)
	err = nil
	return
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
		if (i >= len(arr)) {
			return cost
		}

		operation = arr[i]

		if operation == "M" || operation == "m" || operation == "L" || operation == "l" {
			x = parseFloat(arr[i + 1])
			y = parseFloat(arr[i + 2])
		} else if operation == "V" || operation == "v"  {
			y = parseFloat(arr[i + 1])
		} else if operation == "H" || operation == "h" {
			x = parseFloat(arr[i + 1])
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
	return math.Sqrt(math.Pow((x1 - x0), 2) + math.Pow((y1 - y0), 2))
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
		if (i >= len(arr)) {
			return true
		}
		operation = arr[i]

		switch operation {
			case "M", "m", "L", "l":
				if i + 1 >= len(arr) || i + 2 >= len(arr) {
					return false
				}

				if !isNumeric(arr[i + 1]) || !isNumeric(arr[i + 2]) {
					return false
				}

				i += 3
			case "H", "h", "V", "v":
				if i + 1 >= len(arr) {
					return false
				}

				if !isNumeric(arr[i + 1]) {
					return false
				}

				i += 2
			case "Z", "z":
				if i + 1 >= len(arr) {
					return true
				}

				if isNumeric(arr[i + 1]) {
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


	for i, _ := range(vertices) {
		j := (i + 1 ) % n
		area += vertices[i].x * vertices[j].y
		area -= vertices[j].x * vertices[i].y
	}

	return math.Abs(area) / 2
}


// Compute all the vertices of an SVG string
func ComputeVertices(shapeSvgString string) [][]float64 {
  i := 0
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
  return vertices
}

// Gets the ink cost of a particular operation
// Can return the following errors:
// -InvalidShapeSvgStringError
func (sh Shape) InkCost() (cost uint32, err error) {
  var fcost float64
  if sh.Fill == "transparent" {
    fcost = calculateLineCost(sh.Svg)
  } else {
    fcost, err = calculateFillCost(sh.Svg)
    if err != nil {
      return 0, err
    }
  }

  return uint32(fcost), nil
}

