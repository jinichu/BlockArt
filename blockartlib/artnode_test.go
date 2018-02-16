package blockartlib
import (
"testing"
"math/rand"
"math"
"fmt"
)

// Test shape overlap
func TestTwoLinesDontOverlap(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 0 10 h 20",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 30 10 h 20",
		Stroke: "red",
		Fill:   "transparent",
	}

	expectedResult := false
	actualResult := doesShapeOverlap(sh0, sh1)


	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoLinesColinear(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 0 10 H 20",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 0 10 H 20",
		Stroke: "red",
		Fill:   "transparent",
	}

	expectedResult := true
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}


func TestTwoLinesTouch(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 0 10 H 20",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 0 10 H -20",
		Stroke: "red",
		Fill:   "transparent",
	}

	expectedResult := true
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoLinesCross(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 400 170 L 400 310",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 340 230 L 460 230 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	expectedResult := true
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoShapeOutlinesOverlap(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 400 300 L 400 200 L 300 150 L 300 250 L 400 300 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 350 250 L 500 250 L 500 350 L 400 300 L 350 350 L 350 250 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	expectedResult := true
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoShapeOutlinesDontOverlap(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 400 300 L 400 200 L 300 150 L 300 250 L 400 300 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 800 600 L 800 400 L 600 300 L 600 500 L 800 600 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	expectedResult := false
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoShapeOutlinesDontOverlapOneCompletelyEnclosingTheOther(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 350 250 L 500 250 L 600 350 L 450 450 L 350 350 L 350 250 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 500 300 L 500 350 L 500 350 L 450 350 L 450 300 L 500 300",
		Stroke: "red",
		Fill:   "transparent",
	}
	expectedResult := false
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoShapesDontOverlapOuterShapeOutlineInnerShapeFilled(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 350 250 L 500 250 L 600 350 L 450 450 L 350 350 L 350 250 ",
		Stroke: "red",
		Fill:   "transparent",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 500 300 L 500 350 L 500 350 L 450 350 L 450 300 L 500 300",
		Stroke: "red",
		Fill:   "filled",
	}
	expectedResult := false
	actualResult := doesShapeOverlap(sh0, sh1)


	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTwoShapesDoOverlapOuterShapeFilledInnerShapeOutline(t * testing.T) {
	sh0 := Shape{
		Type:   PATH,
		Svg:    "M 450 150 L 200 300 L 300 500 L 600 300 L 450 150 ",
		Stroke: "red",
		Fill:   "red",
	}

	sh1 := Shape{
		Type:   PATH,
		Svg:    "M 400 300 L 350 300 L 350 350 L 400 350 L 400 300",
		Stroke: "red",
		Fill:   "transparent",
	}


	expectedResult := true
	actualResult := doesShapeOverlap(sh0, sh1)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}

	actualResult = doesShapeOverlap(sh1, sh0)

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

//Test ink cost
func TestInkCostTransparentStrokeTransparentFill(t * testing.T) {
	sh := Shape{
		Type:   PATH,
		Svg:    "M 0 0 H 20 V 20 h -20 Z",
		Stroke: "transparent",
		Fill:   "transparent",
	}
	expectedResult := InvalidShapeSvgStringError(sh.Svg)
	_, actualResult := sh.InkCost()

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestInkCostRedStrokeTransparentFill(t * testing.T) {
	sh := Shape{
		Type:   PATH,
		Svg:    "M 0 0 H 20 V 20 h -20 Z",
		Stroke: "red",
		Fill:   "transparent",
	}
	expectedResult := uint32(80)
	actualResult, _ := sh.InkCost()

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestInkCostTransparentStrokeRedFill(t * testing.T) {
	sh := Shape{
		Type:   PATH,
		Svg:    "M 0 0 H 20 V 20 h -20 Z",
		Stroke: "transparent",
		Fill:   "red",
	}
	expectedResult := uint32(400)
	actualResult, _ := sh.InkCost()

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestInkCostRedStrokeRedFill(t * testing.T) {
	sh := Shape{
		Type:   PATH,
		Svg:    "M 0 0 H 20 V 20 h -20 Z",
		Stroke: "red",
		Fill:   "red",
	}
	expectedResult := uint32(480)
	actualResult, _ := sh.InkCost()

	if expectedResult != actualResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}


// Testing Fill Cost
func TestCalculateSimpleFillCost(t * testing.T) {
	testPath := "M 0 0 H 20 V 20 h -20 Z"
	expectedResult := 400.0
	actualResult, _ := calculateFillCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestCalculateComplexPolygon(t * testing.T) {
	testPath := "M 400 300 L 350 250 L 300 250 L 350 200 L 300 150 L 350 100 L 400 150 L 400 200 L 450 200 L 400 250 L 400 300"
	expectedResult := 12500.0
	actualResult, _ := calculateFillCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestCalculateComplexPolygon2(t * testing.T) {
	testPath := "M 400 250 L 450 200 L 400 150 L 400 200 L 350 200 L 400 250"
	expectedResult := 3750.0
	actualResult, _ := calculateFillCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestCalculateComplexPolygon3(t * testing.T) {
	testPath := "M 390 240 L 450 210 L 390 210 L 360 150 L 330 210 L 300 240 L 300 330 L 390 300 L 390 240"
	expectedResult := 11700.0
	actualResult, _ := calculateFillCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestSelfIntersectionFails(t * testing.T) {
	testPath := "M 400 300 L 500 450 L 400 450 L 500 350 L 400 350 L 400 300"
	expectedResult := InvalidShapeSvgStringError(testPath)
	_, err := calculateFillCost(testPath)
	if err != expectedResult {
		t.Fatalf("Expected %s but got %s", expectedResult.Error(), err)
	}
}

func TestInersectionFails2(t *testing.T) {
	testPath := "M 400 300 L 500 250 L 650 300 L 300 350 L 500 350 L 500 300 L 400 300 "
	expectedResult := InvalidShapeSvgStringError(testPath)
	actualResult, err := calculateFillCost(testPath)
	fmt.Printf("%+v, actual result", actualResult)
	if err != expectedResult {
		t.Fatalf("Expected %s but got %s", expectedResult.Error(), err)
	}
}


//Testing LineCost
func TestCalculateSimpleLineCost(t * testing.T) {
	testPath := "M 0 10 H 20"
	expectedResult := 20.0
	actualResult := calculateLineCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestCalculateBentLineCost(t * testing.T) {
	testPath := "M 50 50 L 100 100 l 25 0"
	expectedResult := math.Sqrt(5000) + 25
	actualResult := calculateLineCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestTrianglesLineCost(t * testing.T) {
	testPath := "M 50 50 L 100 100 l 25 0 Z"
	expectedResult := math.Sqrt(5000) + 25 + math.Sqrt(8125)
	actualResult := calculateLineCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestIntersectingShapeOutlineCost(t * testing.T) {
	testPath := "M 550 200 L 450 300 L 350 200 L 450 200 L 500 250 L 550 300"
	expectedResult := (6 * math.Sqrt(5000)) + 100
	actualResult := calculateLineCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

func TestIntersectingLinesCost(t * testing.T) {
	testPath := "M 250 200 L 400 200 M 300 100 L 300 250"
	expectedResult := 300.0
	actualResult := calculateLineCost(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %f but got %f", expectedResult, actualResult)
	}
}

// Testing svgStringValidityCheck
func TestValidShapeSvgString(t *testing.T) {
	testPaths := []string{
		"M 0 10 H 20",
		"M 0 0 H 20 V 20 h -20 Z",
	}


	for _, testPath := range testPaths {
		expectedResult := nilFunc()
		actualResult := svgStringValidityCheck(testPath)

		if actualResult != expectedResult {
			if actualResult != nil {
				t.Fatalf("Expected %s but got %s", expectedResult.Error(), actualResult.Error())
			}
		}
	}
}
func TestTooLongShapeSvgString(t *testing.T) {
	testPath := generateRandomString(129)
	actualResult := ShapeSvgStringTooLongError(testPath)
	expectedResult := svgStringValidityCheck(testPath)

	if actualResult != expectedResult {
		t.Fatalf("Expected %s but got %s", expectedResult.Error(), actualResult.Error())
	}
}

func TestInvalidPathSvgString(t *testing.T) {
	testPaths := []string{
		"",
		"M 0 10 H 20 Z 30",
		"M 0 10 H 20 z 30",
		"M 0 10 H 20 10 z",
		"M 0 10 h 20 10 z",
		"M 0 10 V 20 10 z",
		"M 0 10 v 20 10 z",
		"M 0 H 20 Z 30 5",
		"m 0 H 20 z 30 5",
		"L 0 H 20 10 z 30",
		"l 0 h 20 10 z 30",
	}


	for _, testPath := range testPaths {
		expectedResult := InvalidShapeSvgStringError(testPath)
		actualResult := svgStringValidityCheck(testPath)

		if actualResult != expectedResult {
			t.Fatalf("Expected %s but got %s", expectedResult.Error(), actualResult.Error())
		}
	}
}


// Testing isValidPath Helper Function
func TestValidPathBasic(t *testing.T) {
	var actualResult bool
	var expectedResult bool
	actualResult = isValidPath("M 0 10 H 20")
	expectedResult = true

	if actualResult != expectedResult {
		t.Fatalf("Expected %t but got %t", expectedResult, actualResult)
	}
}

func TestValidPathNegativeNumbers(t *testing.T) {
	var actualResult bool
	var expectedResult bool
	actualResult = isValidPath("M 0 0 H 20 V 20 h -20 Z")
	expectedResult = true

	if actualResult != expectedResult {
		t.Fatalf("Expected %t but got %t", expectedResult, actualResult)
	}
}

func TestInvalidPathUnknownCommand(t *testing.T) {
	var actualResult bool
	var expectedResult bool

	actualResult = isValidPath("M 0 10 H 20 P")
	expectedResult = false

	if actualResult != expectedResult {
		t.Fatalf("Expected %t but got %t", expectedResult, actualResult)
	}
}

func TestInvalidPathExtraLocationInput(t *testing.T) {
	var actualResult bool

	expectedResult := false

	testPaths := []string{
		"M 0 10 H 20 Z 30",
		"M 0 10 H 20 z 30",
		"M 0 10 H 20 10 z",
		"M 0 10 h 20 10 z",
		"M 0 10 V 20 10 z",
		"M 0 10 v 20 10 z",
	}

	testCharsTested := []string{"Z", "z", "H", "h", "V", "v"}

	for index, path := range testPaths {
		actualResult = isValidPath(path)

		if actualResult != expectedResult {
			t.Fatalf("Expected %t for command %s but got %t", expectedResult, testCharsTested[index], actualResult)
		}
	}
}

func TestInvalidPathLackingLocationInput(t *testing.T) {
	var actualResult bool

	expectedResult := false

	testPaths := []string{
		"M 0 H 20 Z 30 5",
		"m 0 H 20 z 30 5",
		"L 0 H 20 10 z 30",
		"l 0 h 20 10 z 30",
	}

	testCharsTested := []string{"M", "m", "L", "l"}

	for index, path := range testPaths {
		actualResult = isValidPath(path)

		if actualResult != expectedResult {
			t.Fatalf("Expected %t for command %s but got %t", expectedResult, testCharsTested[index], actualResult)
		}
	}
}

// HELPERS Generate random string 
func generateRandomString(n int) (res string) {
	var alphabet = []rune("abcdefghijklmnopqrstuvwxyz")

	x := make([]rune, n)
	for i := range x {
		x[i] = alphabet[rand.Intn(len(alphabet))]
	}
	res = string(x)
	return
}

func nilFunc() error {
	return nil
}
