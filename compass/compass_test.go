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

func TestPolarDirectionsConvergeFromEveryCell(t *testing.T) {
	world, _ := wrex.NewWorld(3)
	for _, test := range []struct {
		name    string
		bearing wrex.Bearing
		seam    wrex.SeamID
	}{
		{name: "north", bearing: North, seam: NorthPoleSeam},
		{name: "south", bearing: South, seam: SouthPoleSeam},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, face := range world.Faces() {
				for q := -world.Radius(); q <= world.Radius(); q++ {
					for r := -world.Radius(); r <= world.Radius(); r++ {
						start := wrex.Cell{Face: face.ID, Hex: wrex.Coord{Q: q, R: r}}
						if !world.Contains(start) {
							continue
						}
						assertConvergesOnSeam(t, world, start, test.bearing, test.seam)
					}
				}
			}
		})
	}
}

func assertConvergesOnSeam(t *testing.T, world *wrex.World, start wrex.Cell, bearing wrex.Bearing, wantSeam wrex.SeamID) {
	t.Helper()
	cell := start
	visited := make(map[wrex.Cell]struct{})
	for steps := int64(0); steps <= world.CellCount(); steps++ {
		if _, exists := visited[cell]; exists {
			t.Fatalf("bearing %d from %#v cycles at %#v", bearing, start, cell)
		}
		visited[cell] = struct{}{}

		direction, err := DirectionTowardPole(world, cell, bearing)
		if err != nil {
			t.Fatalf("DirectionTowardPole(%#v, %d): %v", cell, bearing, err)
		}
		next, err := world.Move(cell, direction)
		if err == nil {
			cell = next
			continue
		}
		var blocked *wrex.ImpassableSeamError
		if !errors.As(err, &blocked) {
			t.Fatalf("move from %#v: %v", cell, err)
		}
		if blocked.Seam != wantSeam {
			t.Fatalf("bearing %d from %#v reached seam %d, want %d", bearing, start, blocked.Seam, wantSeam)
		}
		if next != cell {
			t.Fatalf("blocked move returned %#v, want %#v", next, cell)
		}
		return
	}
	t.Fatalf("bearing %d from %#v did not converge", bearing, start)
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
