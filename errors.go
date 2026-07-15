package wrex

import (
	"github.com/maloquacious/wrex/internal/cerrs"
)

const (
	// ErrInvalidRadius is returned when NewWorld receives a radius outside the
	// supported inclusive range [MinRadius, MaxRadius].
	ErrInvalidRadius = cerrs.Error("wrex: invalid radius")

	// ErrInvalidWorld is returned when an operation requires a World created by
	// NewWorld but receives an uninitialized or otherwise invalid World.
	ErrInvalidWorld = cerrs.Error("wrex: invalid world")

	// ErrInvalidCell is returned when a Cell does not identify a playable
	// coordinate in a World.
	ErrInvalidCell = cerrs.Error("wrex: invalid cell")

	// ErrInvalidCellID is returned when a CellID has nonzero reserved bits or
	// decodes to a face or coordinate that is invalid for a World.
	ErrInvalidCellID = cerrs.Error("wrex: invalid cell id")

	// ErrImpassableSeam is returned when a move would enter one of the six
	// inaccessible square faces.
	ErrImpassableSeam = cerrs.Error("wrex: impassable seam")
)
