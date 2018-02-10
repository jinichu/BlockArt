package blockartlib
import (
	"testing"
	"math/rand"
)

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

