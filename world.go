// Package wrex models a finite hex-grid world on 24 playable hexagonal faces
// separated by six inaccessible square-face seams.
package wrex

import (
	"fmt"
)

const (
	// MinRadius is the smallest supported radius for a playable face.
	MinRadius = 3

	// MaxRadius is the largest supported radius for a playable face.
	MaxRadius = 300

	hexFaceCount = 24
	seamCount    = 6

	cellCoordBits = 10
	cellFaceBits  = 5
	cellCoordMask = (1 << cellCoordBits) - 1
	cellFaceMask  = (1 << cellFaceBits) - 1
	cellQShift    = cellCoordBits
	cellFaceShift = cellCoordBits * 2
	cellUsedBits  = cellCoordBits*2 + cellFaceBits
	cellUsedMask  = (1 << cellUsedBits) - 1
	cellCoordBias = MaxRadius
)

// FaceID identifies one of the world's 24 playable hexagonal faces.
type FaceID uint8

// SeamID identifies one of the six inaccessible square faces. Seams are
// topology metadata, not playable maps or valid Cell locations.
type SeamID uint8

// Coord is an axial hex coordinate. Its implicit cube coordinate is
// S = -Q - R.
type Coord struct {
	Q int
	R int
}

// Cell identifies a playable hex by face and local axial coordinate.
type Cell struct {
	Face FaceID
	Hex  Coord
}

// CellID is the stable, packed 32-bit identifier for a playable Cell.
//
// Layout, least-significant bit first:
//
//	bits  0..9   R + MaxRadius
//	bits 10..19  Q + MaxRadius
//	bits 20..24  FaceID
//	bits 25..31  reserved; must be zero
//
// Q and R each occupy ten bits because the supported coordinate range is
// [-300, 300]. FaceID occupies five bits because 24 faces require five bits.
// The seven high bits are deliberately reserved for future format expansion.
type CellID uint32

// LocalDirection is one of the six directions in a face-local axial frame.
//
// The names are intentionally neutral: a face-local direction is not a global
// compass bearing. Dir0 begins at axial (+1, 0), and the values proceed
// counterclockwise around the face.
type LocalDirection uint8

const (
	Dir0 LocalDirection = iota // (+1,  0)
	Dir1                       // (+1, -1)
	Dir2                       // ( 0, -1)
	Dir3                       // (-1,  0)
	Dir4                       // (-1, +1)
	Dir5                       // ( 0, +1)
)

// Direction is retained as a source-compatible alias for LocalDirection.
// Deprecated: use LocalDirection.
type Direction = LocalDirection

// Delta returns the axial-coordinate movement vector for a local direction.
func (d LocalDirection) Delta() Coord {
	switch d {
	case Dir0:
		return Coord{Q: 1}
	case Dir1:
		return Coord{Q: 1, R: -1}
	case Dir2:
		return Coord{R: -1}
	case Dir3:
		return Coord{Q: -1}
	case Dir4:
		return Coord{Q: -1, R: 1}
	case Dir5:
		return Coord{R: 1}
	default:
		panic(fmt.Sprintf("wrex: invalid direction %d", d))
	}
}

// Bearing is one of six world-relative orientation sectors.
//
// The core package deliberately assigns no compass meaning to these values.
// Bearing0 through Bearing5 are interpreted by child packages such as compass.
// Values proceed clockwise around the world-relative frame.
type Bearing uint8

const (
	Bearing0 Bearing = iota
	Bearing1
	Bearing2
	Bearing3
	Bearing4
	Bearing5
)

// EdgeKind describes what lies beyond one side of a playable face.
type EdgeKind uint8

const (
	HexEdge EdgeKind = iota
	SeamEdge
)

// Edge describes one side of a playable hex face.
type Edge struct {
	Kind    EdgeKind
	Face    FaceID
	Entry   LocalDirection
	Reverse bool
	Seam    SeamID
}

// Face is one playable hexagonal region.
type Face struct {
	ID FaceID

	// Bearing0 is the local direction corresponding to world-relative Bearing0.
	// Child packages may assign a meaning such as geographic north to Bearing0.
	Bearing0 LocalDirection

	Edges [6]Edge
}

// Seam describes one inaccessible square face. It is not a map and cannot be
// occupied by a player. Faces lists the four playable faces bordering it.
type Seam struct {
	ID    SeamID
	Faces [4]FaceID
}

// World contains 24 playable hexagonal faces and six impassable seams.
type World struct {
	radius int
	faces  [hexFaceCount]Face
	seams  [seamCount]Seam
}

// topologyEdge is one entry in the authoritative edge table below. Neighbor
// values 0..23 identify playable faces; values 24..29 identify seams 0..5.
type topologyEdge struct {
	neighbor uint8
	entry    LocalDirection
	reverse  bool
}

// faceTopology is the dual of the class-III (2,1) subdivision of an
// octahedron. It is a spherical rotation system for 24 hexagons and six
// squares. Each row is indexed by the face's local direction and records the
// face (or seam), destination edge, and boundary-position reversal.
var faceTopology = [hexFaceCount][6]topologyEdge{
	{{neighbor: 24}, {neighbor: 21, entry: Dir5}, {neighbor: 22, entry: Dir2, reverse: true}, {neighbor: 2, entry: Dir4, reverse: true}, {neighbor: 1, entry: Dir3, reverse: true}, {neighbor: 3, entry: Dir1}},
	{{neighbor: 26}, {neighbor: 5, entry: Dir5}, {neighbor: 3, entry: Dir2, reverse: true}, {neighbor: 0, entry: Dir4, reverse: true}, {neighbor: 2, entry: Dir3, reverse: true}, {neighbor: 20, entry: Dir1}},
	{{neighbor: 29}, {neighbor: 19, entry: Dir5}, {neighbor: 20, entry: Dir2, reverse: true}, {neighbor: 1, entry: Dir4, reverse: true}, {neighbor: 0, entry: Dir3, reverse: true}, {neighbor: 22, entry: Dir1}},
	{{neighbor: 24}, {neighbor: 0, entry: Dir5}, {neighbor: 1, entry: Dir2, reverse: true}, {neighbor: 5, entry: Dir4, reverse: true}, {neighbor: 4, entry: Dir3, reverse: true}, {neighbor: 15, entry: Dir1}},
	{{neighbor: 28}, {neighbor: 17, entry: Dir5}, {neighbor: 15, entry: Dir2, reverse: true}, {neighbor: 3, entry: Dir4, reverse: true}, {neighbor: 5, entry: Dir3, reverse: true}, {neighbor: 14, entry: Dir1}},
	{{neighbor: 26}, {neighbor: 13, entry: Dir5}, {neighbor: 14, entry: Dir2, reverse: true}, {neighbor: 4, entry: Dir4, reverse: true}, {neighbor: 3, entry: Dir3, reverse: true}, {neighbor: 1, entry: Dir1}},
	{{neighbor: 25}, {neighbor: 18, entry: Dir5}, {neighbor: 19, entry: Dir2, reverse: true}, {neighbor: 8, entry: Dir4, reverse: true}, {neighbor: 7, entry: Dir3, reverse: true}, {neighbor: 9, entry: Dir1}},
	{{neighbor: 27}, {neighbor: 11, entry: Dir5}, {neighbor: 9, entry: Dir2, reverse: true}, {neighbor: 6, entry: Dir4, reverse: true}, {neighbor: 8, entry: Dir3, reverse: true}, {neighbor: 23, entry: Dir1}},
	{{neighbor: 29}, {neighbor: 22, entry: Dir5}, {neighbor: 23, entry: Dir2, reverse: true}, {neighbor: 7, entry: Dir4, reverse: true}, {neighbor: 6, entry: Dir3, reverse: true}, {neighbor: 19, entry: Dir1}},
	{{neighbor: 25}, {neighbor: 6, entry: Dir5}, {neighbor: 7, entry: Dir2, reverse: true}, {neighbor: 11, entry: Dir4, reverse: true}, {neighbor: 10, entry: Dir3, reverse: true}, {neighbor: 12, entry: Dir1}},
	{{neighbor: 28}, {neighbor: 14, entry: Dir5}, {neighbor: 12, entry: Dir2, reverse: true}, {neighbor: 9, entry: Dir4, reverse: true}, {neighbor: 11, entry: Dir3, reverse: true}, {neighbor: 17, entry: Dir1}},
	{{neighbor: 27}, {neighbor: 16, entry: Dir5}, {neighbor: 17, entry: Dir2, reverse: true}, {neighbor: 10, entry: Dir4, reverse: true}, {neighbor: 9, entry: Dir3, reverse: true}, {neighbor: 7, entry: Dir1}},
	{{neighbor: 25}, {neighbor: 9, entry: Dir5}, {neighbor: 10, entry: Dir2, reverse: true}, {neighbor: 14, entry: Dir4, reverse: true}, {neighbor: 13, entry: Dir3, reverse: true}, {neighbor: 18, entry: Dir1}},
	{{neighbor: 26}, {neighbor: 20, entry: Dir5}, {neighbor: 18, entry: Dir2, reverse: true}, {neighbor: 12, entry: Dir4, reverse: true}, {neighbor: 14, entry: Dir3, reverse: true}, {neighbor: 5, entry: Dir1}},
	{{neighbor: 28}, {neighbor: 4, entry: Dir5}, {neighbor: 5, entry: Dir2, reverse: true}, {neighbor: 13, entry: Dir4, reverse: true}, {neighbor: 12, entry: Dir3, reverse: true}, {neighbor: 10, entry: Dir1}},
	{{neighbor: 24}, {neighbor: 3, entry: Dir5}, {neighbor: 4, entry: Dir2, reverse: true}, {neighbor: 17, entry: Dir4, reverse: true}, {neighbor: 16, entry: Dir3, reverse: true}, {neighbor: 21, entry: Dir1}},
	{{neighbor: 27}, {neighbor: 23, entry: Dir5}, {neighbor: 21, entry: Dir2, reverse: true}, {neighbor: 15, entry: Dir4, reverse: true}, {neighbor: 17, entry: Dir3, reverse: true}, {neighbor: 11, entry: Dir1}},
	{{neighbor: 28}, {neighbor: 10, entry: Dir5}, {neighbor: 11, entry: Dir2, reverse: true}, {neighbor: 16, entry: Dir4, reverse: true}, {neighbor: 15, entry: Dir3, reverse: true}, {neighbor: 4, entry: Dir1}},
	{{neighbor: 25}, {neighbor: 12, entry: Dir5}, {neighbor: 13, entry: Dir2, reverse: true}, {neighbor: 20, entry: Dir4, reverse: true}, {neighbor: 19, entry: Dir3, reverse: true}, {neighbor: 6, entry: Dir1}},
	{{neighbor: 29}, {neighbor: 8, entry: Dir5}, {neighbor: 6, entry: Dir2, reverse: true}, {neighbor: 18, entry: Dir4, reverse: true}, {neighbor: 20, entry: Dir3, reverse: true}, {neighbor: 2, entry: Dir1}},
	{{neighbor: 26}, {neighbor: 1, entry: Dir5}, {neighbor: 2, entry: Dir2, reverse: true}, {neighbor: 19, entry: Dir4, reverse: true}, {neighbor: 18, entry: Dir3, reverse: true}, {neighbor: 13, entry: Dir1}},
	{{neighbor: 24}, {neighbor: 15, entry: Dir5}, {neighbor: 16, entry: Dir2, reverse: true}, {neighbor: 23, entry: Dir4, reverse: true}, {neighbor: 22, entry: Dir3, reverse: true}, {neighbor: 0, entry: Dir1}},
	{{neighbor: 29}, {neighbor: 2, entry: Dir5}, {neighbor: 0, entry: Dir2, reverse: true}, {neighbor: 21, entry: Dir4, reverse: true}, {neighbor: 23, entry: Dir3, reverse: true}, {neighbor: 8, entry: Dir1}},
	{{neighbor: 27}, {neighbor: 7, entry: Dir5}, {neighbor: 8, entry: Dir2, reverse: true}, {neighbor: 22, entry: Dir4, reverse: true}, {neighbor: 21, entry: Dir3, reverse: true}, {neighbor: 16, entry: Dir1}},
}

// seamTopology lists the bordering playable faces in cyclic order around each
// square. Together with faceTopology it defines the complete rotation system.
var seamTopology = [seamCount][4]FaceID{
	{0, 21, 15, 3},
	{6, 18, 12, 9},
	{1, 5, 13, 20},
	{7, 11, 16, 23},
	{4, 17, 10, 14},
	{2, 19, 8, 22},
}

// NewWorld constructs a world whose playable faces are regular hex maps of
// the supplied radius.
func NewWorld(radius int) (*World, error) {
	if radius < MinRadius || radius > MaxRadius {
		return nil, fmt.Errorf("%w: got %d, want %d <= radius <= %d", ErrInvalidRadius, radius, MinRadius, MaxRadius)
	}

	w := &World{radius: radius}
	w.initTopology()
	return w, nil
}

// Radius returns the radius of every playable hex face.
func (w *World) Radius() int { return w.radius }

// Faces returns a copy of the 24 playable faces.
func (w *World) Faces() []Face {
	if !w.valid() {
		return nil
	}
	faces := make([]Face, len(w.faces))
	copy(faces, w.faces[:])
	return faces
}

// Seams returns a copy of the six impassable square-face descriptors.
func (w *World) Seams() []Seam {
	if !w.valid() {
		return nil
	}
	seams := make([]Seam, len(w.seams))
	copy(seams, w.seams[:])
	return seams
}

// Contains reports whether a cell names a playable coordinate in the world.
func (w *World) Contains(cell Cell) bool {
	if !w.valid() || int(cell.Face) >= len(w.faces) {
		return false
	}

	q, r := cell.Hex.Q, cell.Hex.R
	if q < -w.radius || q > w.radius || r < -w.radius || r > w.radius {
		return false
	}
	s := q + r
	return s >= -w.radius && s <= w.radius
}

// EncodeCell validates cell and returns its stable packed identifier.
func (w *World) EncodeCell(cell Cell) (CellID, error) {
	if !w.valid() {
		return 0, ErrInvalidWorld
	}
	if !w.Contains(cell) {
		return 0, fmt.Errorf("%w: face=%d q=%d r=%d", ErrInvalidCell, cell.Face, cell.Hex.Q, cell.Hex.R)
	}

	q := uint32(cell.Hex.Q + cellCoordBias)
	r := uint32(cell.Hex.R + cellCoordBias)
	id := uint32(cell.Face)<<cellFaceShift | q<<cellQShift | r
	return CellID(id), nil
}

// DecodeCell validates id and returns the Cell it identifies in this World.
// Reserved bits must be zero, making accidental use of a future encoding fail
// explicitly rather than decode silently.
func (w *World) DecodeCell(id CellID) (Cell, error) {
	if !w.valid() {
		return Cell{}, ErrInvalidWorld
	}
	raw := uint32(id)
	if raw & ^uint32(cellUsedMask) != 0 {
		return Cell{}, fmt.Errorf("%w: reserved bits are nonzero: %#08x", ErrInvalidCellID, raw)
	}

	cell := Cell{
		Face: FaceID((raw >> cellFaceShift) & cellFaceMask),
		Hex: Coord{
			Q: int((raw>>cellQShift)&cellCoordMask) - cellCoordBias,
			R: int(raw&cellCoordMask) - cellCoordBias,
		},
	}
	if !w.Contains(cell) {
		return Cell{}, fmt.Errorf("%w: decoded face=%d q=%d r=%d", ErrInvalidCellID, cell.Face, cell.Hex.Q, cell.Hex.R)
	}
	return cell, nil
}

// Move attempts to move one step in direction d.
func (w *World) Move(cell Cell, d LocalDirection) (Cell, error) {
	if !w.valid() {
		return cell, ErrInvalidWorld
	}
	if !w.Contains(cell) {
		return cell, fmt.Errorf("%w: face=%d q=%d r=%d", ErrInvalidCell, cell.Face, cell.Hex.Q, cell.Hex.R)
	}
	if d > Dir5 {
		return cell, fmt.Errorf("wrex: invalid direction: %d", d)
	}

	delta := d.Delta()
	next := Coord{Q: cell.Hex.Q + delta.Q, R: cell.Hex.R + delta.R}
	if Distance(Coord{}, next) <= w.radius {
		return Cell{Face: cell.Face, Hex: next}, nil
	}

	exit := d
	if _, onExit := edgePosition(cell.Hex, exit, w.radius); !onExit {
		// A move from a corner can leave through an edge other than d. Derive
		// that edge from the violated bound; when two bounds are violated and
		// the cell lies on edge d, the explicit choice above removes ambiguity.
		var ok bool
		exit, ok = exitEdge(next, w.radius)
		if !ok {
			return cell, fmt.Errorf("wrex: could not determine exit edge: face=%d q=%d r=%d direction=%d", cell.Face, cell.Hex.Q, cell.Hex.R, d)
		}
	}
	edge := w.faces[cell.Face].Edges[exit]
	if edge.Kind == SeamEdge {
		return cell, fmt.Errorf("%w: seam=%d face=%d direction=%d edge=%d", ErrImpassableSeam, edge.Seam, cell.Face, d, exit)
	}

	position, ok := edgePosition(cell.Hex, exit, w.radius)
	if !ok {
		return cell, fmt.Errorf("wrex: coordinate is not on exit edge: face=%d q=%d r=%d direction=%d edge=%d", cell.Face, cell.Hex.Q, cell.Hex.R, d, exit)
	}
	if edge.Reverse {
		position = w.radius - position
	}

	return Cell{Face: edge.Face, Hex: boundaryCoord(edge.Entry, position, w.radius)}, nil
}

// BearingFor converts a face-local direction to a world-relative bearing.
func (w *World) BearingFor(face FaceID, d LocalDirection) (Bearing, error) {
	if !w.valid() {
		return 0, ErrInvalidWorld
	}
	if int(face) >= len(w.faces) || d > Dir5 {
		return 0, fmt.Errorf("wrex: invalid face or local direction: face=%d direction=%d", face, d)
	}
	zero := w.faces[face].Bearing0
	return Bearing((int(zero) - int(d) + 6) % 6), nil
}

// LocalDirectionFor converts a world-relative bearing to the corresponding
// local movement direction on face.
func (w *World) LocalDirectionFor(face FaceID, b Bearing) (LocalDirection, error) {
	if !w.valid() {
		return 0, ErrInvalidWorld
	}
	if int(face) >= len(w.faces) || b > Bearing5 {
		return 0, fmt.Errorf("wrex: invalid face or bearing: face=%d bearing=%d", face, b)
	}
	zero := w.faces[face].Bearing0
	return LocalDirection((int(zero) - int(b) + 6) % 6), nil
}

// Distance returns the number of hex steps between two axial coordinates on
// the same face. If the mathematical distance cannot be represented by an
// int, Distance returns the largest representable int.
func Distance(a, b Coord) int {
	const maxInt = int(^uint(0) >> 1)

	dq, qNegative := differenceMagnitude(a.Q, b.Q)
	dr, rNegative := differenceMagnitude(a.R, b.R)
	if dq > uint(maxInt) || dr > uint(maxInt) {
		return maxInt
	}

	var ds uint
	if qNegative == rNegative {
		ds = dq + dr
	} else if dq >= dr {
		ds = dq - dr
	} else {
		ds = dr - dq
	}
	if ds > uint(maxInt) {
		return maxInt
	}

	distance := max(dq, dr, ds)
	return int(distance)
}

// CellCount returns the number of playable cells in the world.
func (w *World) CellCount() int64 {
	if !w.valid() {
		return 0
	}
	r := int64(w.radius)
	return hexFaceCount * (1 + 3*r*(r+1))
}

func (w *World) valid() bool {
	return w != nil && w.radius >= MinRadius && w.radius <= MaxRadius
}

// initTopology instantiates the authoritative spherical rotation system.
func (w *World) initTopology() {
	for i := range w.faces {
		w.faces[i].ID = FaceID(i)
		for d, topology := range faceTopology[i] {
			if topology.neighbor >= hexFaceCount {
				w.faces[i].Edges[d] = Edge{Kind: SeamEdge, Seam: SeamID(topology.neighbor - hexFaceCount)}
				continue
			}
			w.faces[i].Edges[d] = Edge{
				Kind:    HexEdge,
				Face:    FaceID(topology.neighbor),
				Entry:   topology.entry,
				Reverse: topology.reverse,
			}
		}
	}

	for s := range w.seams {
		w.seams[s].ID = SeamID(s)
		w.seams[s].Faces = seamTopology[s]
	}

	w.assignBearing0Directions()
}

func (w *World) assignBearing0Directions() {
	const infinity = int(^uint(0) >> 1)
	distance := [hexFaceCount]int{}
	for i := range distance {
		distance[i] = infinity
	}

	queue := make([]FaceID, 0, hexFaceCount)
	for _, face := range w.faces {
		for d, edge := range face.Edges {
			if edge.Kind == SeamEdge && edge.Seam == 0 {
				distance[face.ID] = 0
				w.faces[face.ID].Bearing0 = LocalDirection(d)
				queue = append(queue, face.ID)
				break
			}
		}
	}

	for head := 0; head < len(queue); head++ {
		current := queue[head]
		for _, edge := range w.faces[current].Edges {
			if edge.Kind != HexEdge || distance[edge.Face] != infinity {
				continue
			}
			distance[edge.Face] = distance[current] + 1
			queue = append(queue, edge.Face)
		}
	}

	for i := range w.faces {
		if distance[i] == 0 {
			continue
		}
		best := Dir0
		bestDistance := distance[i]
		for d, edge := range w.faces[i].Edges {
			if edge.Kind == HexEdge && distance[edge.Face] < bestDistance {
				best = LocalDirection(d)
				bestDistance = distance[edge.Face]
			}
		}
		w.faces[i].Bearing0 = best
	}
}

func exitEdge(c Coord, radius int) (LocalDirection, bool) {
	switch {
	case c.Q > radius:
		return Dir0, true
	case c.R < -radius:
		return Dir1, true
	case c.Q+c.R < -radius:
		return Dir2, true
	case c.Q < -radius:
		return Dir3, true
	case c.R > radius:
		return Dir4, true
	case c.Q+c.R > radius:
		return Dir5, true
	default:
		return 0, false
	}
}

func edgePosition(c Coord, d LocalDirection, radius int) (int, bool) {
	switch d {
	case Dir0:
		return c.R + radius, c.Q == radius && c.R >= -radius && c.R <= 0
	case Dir1:
		return c.Q, c.R == -radius && c.Q >= 0 && c.Q <= radius
	case Dir2:
		return c.Q + radius, c.Q+c.R == -radius && c.Q >= -radius && c.Q <= 0
	case Dir3:
		return c.R, c.Q == -radius && c.R >= 0 && c.R <= radius
	case Dir4:
		return c.Q + radius, c.R == radius && c.Q >= -radius && c.Q <= 0
	case Dir5:
		return c.Q, c.Q+c.R == radius && c.Q >= 0 && c.Q <= radius
	default:
		return 0, false
	}
}

func boundaryCoord(d LocalDirection, position, radius int) Coord {
	if position < 0 || position > radius {
		panic("wrex: invalid boundary position")
	}
	switch d {
	case Dir0:
		return Coord{Q: radius, R: -radius + position}
	case Dir1:
		return Coord{Q: position, R: -radius}
	case Dir2:
		return Coord{Q: -radius + position, R: -position}
	case Dir3:
		return Coord{Q: -radius, R: position}
	case Dir4:
		return Coord{Q: -radius + position, R: radius}
	case Dir5:
		return Coord{Q: position, R: radius - position}
	default:
		panic("wrex: invalid boundary direction")
	}
}

// differenceMagnitude returns the magnitude and sign of b-a without
// performing the potentially overflowing signed subtraction.
func differenceMagnitude(a, b int) (magnitude uint, negative bool) {
	if b >= a {
		return uint(b) - uint(a), false
	}
	return uint(a) - uint(b), true
}
