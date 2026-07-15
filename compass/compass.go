// Package compass assigns geographic compass names and polar meaning to the
// neutral world-relative bearings defined by package wrex.
//
// The core package models only Bearing0 through Bearing5. This package adopts
// the convention that Bearing0 is north and that the remaining values proceed
// clockwise in 60-degree sectors.
package compass

import (
	"fmt"

	"github.com/maloquacious/wrex"
)

const (
	North     wrex.Bearing = wrex.Bearing0
	Northeast wrex.Bearing = wrex.Bearing1
	Southeast wrex.Bearing = wrex.Bearing2
	South     wrex.Bearing = wrex.Bearing3
	Southwest wrex.Bearing = wrex.Bearing4
	Northwest wrex.Bearing = wrex.Bearing5

	// NorthPoleSeam is the inaccessible seam toward which Bearing0 converges.
	NorthPoleSeam wrex.SeamID = 0
	// SouthPoleSeam is the inaccessible seam opposite NorthPoleSeam.
	SouthPoleSeam wrex.SeamID = 3
)

// LocalDirection converts a compass bearing through a face's reference frame.
// It is not a convergent polar-routing policy; use DirectionTowardPole when
// navigating north or south.
func LocalDirection(world *wrex.World, face wrex.FaceID, bearing wrex.Bearing) (wrex.LocalDirection, error) {
	if bearing > Northwest {
		return 0, fmt.Errorf("compass: invalid bearing %d", bearing)
	}
	return world.LocalDirectionFor(face, bearing)
}

// DirectionTowardPole returns the next local direction on a shortest route
// toward the pole identified by bearing. The direction is cell-dependent and
// must be recomputed after every move. Only North and South identify poles.
func DirectionTowardPole(world *wrex.World, cell wrex.Cell, bearing wrex.Bearing) (wrex.LocalDirection, error) {
	var seam wrex.SeamID
	switch bearing {
	case North:
		seam = NorthPoleSeam
	case South:
		seam = SouthPoleSeam
	default:
		return 0, fmt.Errorf("compass: bearing %d does not identify a pole", bearing)
	}
	return world.DirectionTowardSeam(cell, seam)
}

// Bearing returns the compass bearing represented by a local direction on a
// particular face.
func Bearing(world *wrex.World, face wrex.FaceID, direction wrex.LocalDirection) (wrex.Bearing, error) {
	return world.BearingFor(face, direction)
}

// Pole returns the inaccessible seam used for the requested polar bearing.
// Only North and South identify poles.
func Pole(world *wrex.World, bearing wrex.Bearing) (wrex.Seam, bool) {
	var id wrex.SeamID
	switch bearing {
	case North:
		id = NorthPoleSeam
	case South:
		id = SouthPoleSeam
	default:
		return wrex.Seam{}, false
	}
	for _, seam := range world.Seams() {
		if seam.ID == id {
			return seam, true
		}
	}
	return wrex.Seam{}, false
}
