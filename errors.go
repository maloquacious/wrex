package wrex

import (
	"fmt"

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

// ImpassableSeamError describes a move blocked by an inaccessible seam.
// It unwraps to ErrImpassableSeam so callers can use both errors.Is and
// errors.As.
type ImpassableSeamError struct {
	Seam      SeamID
	Cell      Cell
	Direction LocalDirection
	Exit      LocalDirection
}

func (e *ImpassableSeamError) Error() string {
	return fmt.Sprintf("%s: seam=%d face=%d direction=%d edge=%d", ErrImpassableSeam, e.Seam, e.Cell.Face, e.Direction, e.Exit)
}

// Unwrap returns ErrImpassableSeam.
func (e *ImpassableSeamError) Unwrap() error { return ErrImpassableSeam }
