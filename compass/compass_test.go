package compass

import (
	"errors"
	"testing"

	"github.com/maloquacious/wrex"
)

func TestBearingRoundTrip(t *testing.T) {
	world, _ := wrex.NewWorld(3)
	for _, face := range world.Faces() {
		for d := wrex.Dir0; d <= wrex.Dir5; d++ {
			bearing, err := Bearing(world, face.ID, d)
			if err != nil {
				t.Fatalf("Bearing(%d, %d): %v", face.ID, d, err)
			}
			got, err := LocalDirection(world, face.ID, bearing)
			if err != nil {
				t.Fatalf("LocalDirection(%d, %d): %v", face.ID, bearing, err)
			}
			if got != d {
				t.Fatalf("face %d direction round trip = %d, want %d", face.ID, got, d)
			}
		}
	}
}

func TestNorthPoleIsInaccessibleAndNorthConverges(t *testing.T) {
	world, _ := wrex.NewWorld(3)
	pole, ok := Pole(world, North)
	if !ok || pole.ID != NorthPoleSeam {
		t.Fatalf("north pole = %#v, %v", pole, ok)
	}

	for _, startingFace := range world.Faces() {
		cell := wrex.Cell{Face: startingFace.ID}
		blocked := false
		for steps := 0; steps < 1_000; steps++ {
			d, err := LocalDirection(world, cell.Face, North)
			if err != nil {
				t.Fatal(err)
			}
			next, err := world.Move(cell, d)
			if errors.Is(err, wrex.ErrImpassableSeam) {
				blocked = true
				break
			}
			if err != nil {
				t.Fatalf("northward move from face %d: %v", startingFace.ID, err)
			}
			cell = next
		}
		if !blocked {
			t.Fatalf("northward travel from face %d did not converge", startingFace.ID)
		}
	}
}

func TestPolarSeamsAreDistinct(t *testing.T) {
	world, _ := wrex.NewWorld(3)
	north, northOK := Pole(world, North)
	south, southOK := Pole(world, South)
	if !northOK || !southOK {
		t.Fatalf("missing pole: north=%v south=%v", northOK, southOK)
	}
	if north.ID == south.ID {
		t.Fatalf("polar seams share ID %d", north.ID)
	}
}

func TestZeroValueWorldHasNoPole(t *testing.T) {
	var world wrex.World
	if pole, ok := Pole(&world, North); ok {
		t.Fatalf("Pole = %#v, true, want no pole", pole)
	}
	if _, err := LocalDirection(&world, 0, North); !errors.Is(err, wrex.ErrInvalidWorld) {
		t.Fatalf("LocalDirection error = %v, want ErrInvalidWorld", err)
	}
	if _, err := Bearing(&world, 0, wrex.Dir0); !errors.Is(err, wrex.ErrInvalidWorld) {
		t.Fatalf("Bearing error = %v, want ErrInvalidWorld", err)
	}
}
