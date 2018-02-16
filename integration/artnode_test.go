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
    want := uint32(80)
    if ink < want {
      return fmt.Errorf("GetInk() = %d; want %d", ink, want)
    }
    return nil
  })

  svgString := "M 400 300 L 500 250 L 650 300 L 300 350 L 500 350 L 500 300 L 400 300"

  _, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgString, "red", "red")
  if err != blockartlib.InvalidShapeSvgStringError(svgString) {
    t.Fatalf("%s", err)
  }
}
