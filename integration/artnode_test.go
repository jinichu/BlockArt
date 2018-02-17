package integration

import (
	"fmt"
	"sync"
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

	{
		svgStringBad := "M 400 300 L 500 250 L 650 300 L 300 350 L 500 350 L 500 300 L 400 300"
		_, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgStringBad, "red", "red")
		want := blockartlib.InvalidShapeSvgStringError(svgStringBad)
		if err.Error() != want.Error() {
			t.Fatalf("got %s; want %s", err, want)
		}
	}

	{
		svgStringOk := "M 40 30 L 50 25 L 20 30 L 30 35 L 50 35 L 50 30 L 40 30"

		if _, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgStringOk, "red", "red"); err != nil {
			t.Fatal(err)
		}
	}

	{
		svgStringOk := "M 0 10 V 20 z"

		if _, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgStringOk, "red", "red"); err != nil {
			t.Fatal(err)
		}
	}

	{
		svgStringBad := "M 0 0 H 20 V 20 -20 Z"
		_, _, _, err := canvas.AddShape(6, blockartlib.PATH, svgStringBad, "red", "red")
		want := blockartlib.InvalidShapeSvgStringError(svgStringBad)
		if err.Error() != want.Error() {
			t.Fatalf("got %s; want %s", err, want)
		}
	}
}

func TestArtNodeClusterOnlyOneCommits(t *testing.T) {
	defer SetBlockDelay(50 * time.Millisecond)()

	ts := NewTestCluster(t, 2)
	defer ts.Close()

	SucceedsSoon(t, func() error {
		ink, err := ts.ArtNodes[0].GetInk()
		if err != nil {
			return err
		}
		want := uint32(300)
		if ink < want {
			return fmt.Errorf("GetInk() = %d; want %d", ink, want)
		}
		return nil
	})

	svgString := "M 40 30 L 50 25 L 20 30 L 30 35 L 50 35 L 50 30 L 40 30"

	results := struct {
		sync.Mutex

		errors int
	}{}

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			if _, _, _, err := ts.ArtNodes[i].AddShape(6, blockartlib.PATH, svgString, "red", "red"); err != nil {
				results.Lock()
				results.errors += 1
				results.Unlock()
				return
			}
		}()
	}

	wg.Wait()
	results.Lock()
	defer results.Unlock()

	want := 1
	if results.errors != want {
		t.Fatalf("expected want error, got %d", results.errors)
	}
}
