package wrex

import (
	"errors"
	"math"
	"testing"
)

func TestNewWorld(t *testing.T) {
	world, err := NewWorld(3)
	if err != nil {
		t.Fatalf("NewWorld: %v", err)
	}
	if got, want := world.Radius(), 3; got != want {
		t.Fatalf("Radius = %d, want %d", got, want)
	}
	if got, want := len(world.Faces()), 24; got != want {
		t.Fatalf("len(Faces) = %d, want %d", got, want)
	}
	if got, want := len(world.Seams()), 6; got != want {
		t.Fatalf("len(Seams) = %d, want %d", got, want)
	}
	if got, want := world.CellCount(), int64(888); got != want {
		t.Fatalf("CellCount = %d, want %d", got, want)
	}
}

func TestNewWorldRadiusRange(t *testing.T) {
	for _, tt := range []struct {
		radius  int
		wantErr bool
	}{
		{MinRadius - 1, true}, {MinRadius, false}, {MaxRadius, false}, {MaxRadius + 1, true},
	} {
		world, err := NewWorld(tt.radius)
		if tt.wantErr {
			if !errors.Is(err, ErrInvalidRadius) || world != nil {
				t.Fatalf("NewWorld(%d) = %#v, %v", tt.radius, world, err)
			}
		} else if err != nil {
			t.Fatalf("NewWorld(%d): %v", tt.radius, err)
		}
	}
}

func TestMaximumRadiusCellCount(t *testing.T) {
	world, _ := NewWorld(MaxRadius)
	const want int64 = 6_501_624
	if got := world.CellCount(); got != want {
		t.Fatalf("CellCount = %d, want %d", got, want)
	}
}

func TestCellIDRoundTrip(t *testing.T) {
	world, _ := NewWorld(MaxRadius)
	cells := []Cell{
		{Face: 0, Hex: Coord{}},
		{Face: 23, Hex: Coord{Q: 300, R: -300}},
		{Face: 12, Hex: Coord{Q: -300, R: 0}},
		{Face: 7, Hex: Coord{Q: 0, R: 300}},
	}
	for _, cell := range cells {
		id, err := world.EncodeCell(cell)
		if err != nil {
			t.Fatalf("EncodeCell(%#v): %v", cell, err)
		}
		got, err := world.DecodeCell(id)
		if err != nil {
			t.Fatalf("DecodeCell(%d): %v", id, err)
		}
		if got != cell {
			t.Fatalf("round trip = %#v, want %#v", got, cell)
		}
		if uint32(id)>>cellUsedBits != 0 {
			t.Fatalf("CellID %#08x uses reserved bits", uint32(id))
		}
	}
}

func TestCellIDIsStableAcrossWorldRadiusWhenCellExists(t *testing.T) {
	small, _ := NewWorld(3)
	large, _ := NewWorld(300)
	cell := Cell{Face: 5, Hex: Coord{Q: 2, R: -1}}
	a, _ := small.EncodeCell(cell)
	b, _ := large.EncodeCell(cell)
	if a != b {
		t.Fatalf("IDs differ: %d != %d", a, b)
	}
}

func TestCellIDRejectsInvalidValues(t *testing.T) {
	world, _ := NewWorld(3)
	if _, err := world.EncodeCell(Cell{Face: 24}); !errors.Is(err, ErrInvalidCell) {
		t.Fatalf("EncodeCell error = %v", err)
	}
	if _, err := world.EncodeCell(Cell{Face: 0, Hex: Coord{Q: 4}}); !errors.Is(err, ErrInvalidCell) {
		t.Fatalf("EncodeCell error = %v", err)
	}
	if _, err := world.DecodeCell(CellID(1 << 31)); !errors.Is(err, ErrInvalidCellID) {
		t.Fatalf("reserved-bit error = %v", err)
	}

	large, _ := NewWorld(300)
	id, _ := large.EncodeCell(Cell{Face: 0, Hex: Coord{Q: 300}})
	if _, err := world.DecodeCell(id); !errors.Is(err, ErrInvalidCellID) {
		t.Fatalf("out-of-radius error = %v", err)
	}
}

func TestCellIDRejectsExtremeCoordinates(t *testing.T) {
	world, _ := NewWorld(MaxRadius)
	centerID, err := world.EncodeCell(Cell{})
	if err != nil {
		t.Fatalf("EncodeCell(center): %v", err)
	}

	coordinates := []Coord{
		{Q: math.MinInt},
		{Q: math.MaxInt},
		{R: math.MinInt},
		{R: math.MaxInt},
		{Q: math.MinInt, R: math.MaxInt},
		{Q: math.MaxInt, R: math.MinInt},
	}
	for _, coordinate := range coordinates {
		cell := Cell{Hex: coordinate}
		if world.Contains(cell) {
			t.Errorf("Contains(%#v) = true, want false", cell)
		}
		id, err := world.EncodeCell(cell)
		if !errors.Is(err, ErrInvalidCell) {
			t.Errorf("EncodeCell(%#v) error = %v, want ErrInvalidCell", cell, err)
		}
		if err == nil && id == centerID {
			t.Errorf("EncodeCell(%#v) collided with center ID %d", cell, centerID)
		}
	}
}

func TestDistanceSaturatesOnOverflow(t *testing.T) {
	tests := []struct {
		a, b Coord
		want int
	}{
		{a: Coord{}, b: Coord{Q: math.MinInt}, want: math.MaxInt},
		{a: Coord{}, b: Coord{R: math.MaxInt}, want: math.MaxInt},
		{a: Coord{Q: math.MinInt}, b: Coord{Q: math.MaxInt}, want: math.MaxInt},
		{a: Coord{Q: math.MinInt, R: math.MaxInt}, b: Coord{Q: math.MaxInt, R: math.MinInt}, want: math.MaxInt},
		{a: Coord{}, b: Coord{Q: math.MaxInt, R: -math.MaxInt}, want: math.MaxInt},
	}
	for _, tt := range tests {
		if got := Distance(tt.a, tt.b); got != tt.want {
			t.Errorf("Distance(%#v, %#v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCellIDKnownLayout(t *testing.T) {
	world, _ := NewWorld(300)
	cell := Cell{Face: 23, Hex: Coord{Q: -300, R: 300}}
	id, err := world.EncodeCell(cell)
	if err != nil {
		t.Fatal(err)
	}
	want := CellID(uint32(23)<<20 | uint32(0)<<10 | uint32(600))
	if id != want {
		t.Fatalf("CellID = %#08x, want %#08x", uint32(id), uint32(want))
	}
}

func TestMoveWithinFace(t *testing.T) {
	world, _ := NewWorld(3)
	got, err := world.Move(Cell{Face: 0}, Dir0)
	if err != nil {
		t.Fatal(err)
	}
	want := Cell{Face: 0, Hex: Coord{Q: 1}}
	if got != want {
		t.Fatalf("Move = %#v, want %#v", got, want)
	}
}

func TestMoveIntoSeamIsBlocked(t *testing.T) {
	world, _ := NewWorld(3)
	start := Cell{Face: 0, Hex: Coord{Q: 3, R: -1}}
	got, err := world.Move(start, Dir0)
	if !errors.Is(err, ErrImpassableSeam) {
		t.Fatalf("Move error = %v", err)
	}
	if got != start {
		t.Fatalf("blocked Move = %#v, want %#v", got, start)
	}
}

func TestTopologyCountsAndReciprocity(t *testing.T) {
	world, _ := NewWorld(3)
	seamIncidences, hexIncidences := 0, 0
	for _, face := range world.Faces() {
		for d, edge := range face.Edges {
			if edge.Kind == SeamEdge {
				seamIncidences++
				continue
			}
			hexIncidences++
			back := world.faces[edge.Face].Edges[edge.Entry]
			if back.Kind != HexEdge || back.Face != face.ID || back.Entry != LocalDirection(d) {
				t.Fatalf("face %d edge %d is not reciprocal", face.ID, d)
			}
		}
	}
	if seamIncidences != 24 {
		t.Fatalf("seam incidences = %d, want 24", seamIncidences)
	}
	if hexIncidences != 120 {
		t.Fatalf("hex incidences = %d, want 120", hexIncidences)
	}
	for _, seam := range world.Seams() {
		seen := map[FaceID]bool{}
		for _, face := range seam.Faces {
			seen[face] = true
		}
		if len(seen) != 4 {
			t.Fatalf("seam %d borders %d unique faces, want 4", seam.ID, len(seen))
		}
	}
}

func TestLocalDirectionDeltas(t *testing.T) {
	want := []Coord{
		{Q: 1, R: 0},
		{Q: 1, R: -1},
		{Q: 0, R: -1},
		{Q: -1, R: 0},
		{Q: -1, R: 1},
		{Q: 0, R: 1},
	}
	for d, delta := range want {
		if got := LocalDirection(d).Delta(); got != delta {
			t.Fatalf("Dir%d.Delta() = %#v, want %#v", d, got, delta)
		}
	}
}
