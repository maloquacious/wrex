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

// LocalDirection returns the face-local movement direction for a compass
// bearing. Recompute it after every face transition.
func LocalDirection(world *wrex.World, face wrex.FaceID, bearing wrex.Bearing) (wrex.LocalDirection, error) {
	if bearing > Northwest {
		return 0, fmt.Errorf("compass: invalid bearing %d", bearing)
	}
	return world.LocalDirectionFor(face, bearing)
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
