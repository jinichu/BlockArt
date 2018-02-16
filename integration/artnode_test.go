package integration

import (
	"fmt"
	"testing"
	"time"

	"../blockartlib"
)

func TestArtNodeClusterErrors(t *testing.T) {
	defer SetBlockDelay(5 * time.Millisecond)()

	ts := NewTestCluster(t, 1)
	defer ts.Close()

	canvas := ts.ArtNodes[0]

	SucceedsSoon(t, func() error {
		ink, err := canvas.GetInk()
		if err != nil {
			return err
		}
		want := uint32(300)
		if ink < want {
			return fmt.Errorf("GetInk() = %d; want %d", ink, want)
		}
		return nil
	})

	svgStringBad := "M 400 300 L 500 250 L 650 300 L 300 350 L 500 350 L 500 300 L 400 300"

	_, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgStringBad, "red", "red")
	want := blockartlib.InvalidShapeSvgStringError(svgStringBad)
	if err.Error() != want.Error() {
		t.Fatalf("got %s; want %s", err, want)
	}

	svgStringOk := "M 40 30 L 50 25 L 20 30 L 30 35 L 50 35 L 50 30 L 40 30"

	if _, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgStringOk, "red", "red"); err != nil {
		t.Fatal(err)
	}
}
