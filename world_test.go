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

func TestZeroValueWorldIsInvalid(t *testing.T) {
	var world World
	cell := Cell{}

	if world.Contains(cell) {
		t.Error("Contains(Cell{}) = true, want false")
	}
	if _, err := world.EncodeCell(cell); !errors.Is(err, ErrInvalidWorld) {
		t.Errorf("EncodeCell error = %v, want ErrInvalidWorld", err)
	}
	if _, err := world.DecodeCell(0); !errors.Is(err, ErrInvalidWorld) {
		t.Errorf("DecodeCell error = %v, want ErrInvalidWorld", err)
	}
	if got, err := world.Move(cell, Dir0); !errors.Is(err, ErrInvalidWorld) || got != cell {
		t.Errorf("Move = %#v, %v, want original cell and ErrInvalidWorld", got, err)
	}
	if _, err := world.BearingFor(0, Dir0); !errors.Is(err, ErrInvalidWorld) {
		t.Errorf("BearingFor error = %v, want ErrInvalidWorld", err)
	}
	if _, err := world.LocalDirectionFor(0, Bearing0); !errors.Is(err, ErrInvalidWorld) {
		t.Errorf("LocalDirectionFor error = %v, want ErrInvalidWorld", err)
	}
	if got := world.Faces(); got != nil {
		t.Errorf("Faces = %#v, want nil", got)
	}
	if got := world.Seams(); got != nil {
		t.Errorf("Seams = %#v, want nil", got)
	}
	if got := world.CellCount(); got != 0 {
		t.Errorf("CellCount = %d, want 0", got)
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
			if back.Kind != HexEdge || back.Face != face.ID || back.Entry != LocalDirection(d) || back.Reverse != edge.Reverse {
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

func TestTopologyIsClosedSphere(t *testing.T) {
	world, _ := NewWorld(3)

	const cornerCount = hexFaceCount*6 + seamCount*4
	parents := [cornerCount]int{}
	owners := [cornerCount]uint8{}
	for i := range parents {
		parents[i] = i
		if i < hexFaceCount*6 {
			owners[i] = uint8(i / 6)
		} else {
			owners[i] = hexFaceCount + uint8((i-hexFaceCount*6)/4)
		}
	}
	var find func(int) int
	find = func(i int) int {
		if parents[i] != i {
			parents[i] = find(parents[i])
		}
		return parents[i]
	}
	join := func(a, b int) {
		a, b = find(a), find(b)
		if a != b {
			parents[b] = a
		}
	}

	// Local side order around a hexagon is 0,5,4,3,2,1. Position increases
	// with that order on sides 0..2 and against it on sides 3..5.
	hexCorner := func(face FaceID, d LocalDirection, position int) int {
		side := 0
		if d != Dir0 {
			side = 6 - int(d)
		}
		forward := d <= Dir2
		if (position == 1) == forward {
			side = (side + 1) % 6
		}
		return int(face)*6 + side
	}

	edges := 0
	for _, face := range world.faces {
		for d, edge := range face.Edges {
			if edge.Kind != HexEdge || face.ID > edge.Face {
				continue
			}
			edges++
			for position := 0; position < 2; position++ {
				destinationPosition := position
				if edge.Reverse {
					destinationPosition = 1 - destinationPosition
				}
				join(
					hexCorner(face.ID, LocalDirection(d), position),
					hexCorner(edge.Face, edge.Entry, destinationPosition),
				)
			}
		}
	}
	for _, seam := range world.seams {
		for side, face := range seam.Faces {
			edges++
			var direction LocalDirection
			found := false
			for d, edge := range world.faces[face].Edges {
				if edge.Kind == SeamEdge && edge.Seam == seam.ID {
					direction = LocalDirection(d)
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("seam %d lists face %d without a reciprocal incidence", seam.ID, face)
			}
			seamStart := hexFaceCount*6 + int(seam.ID)*4 + side
			seamEnd := hexFaceCount*6 + int(seam.ID)*4 + (side+1)%4
			join(seamStart, hexCorner(face, direction, 1))
			join(seamEnd, hexCorner(face, direction, 0))
		}
	}

	vertices := map[int]map[uint8]bool{}
	vertexCorners := map[int]int{}
	for corner, owner := range owners {
		root := find(corner)
		if vertices[root] == nil {
			vertices[root] = map[uint8]bool{}
		}
		vertices[root][owner] = true
		vertexCorners[root]++
	}
	if got, want := len(vertices), 56; got != want {
		t.Fatalf("polyhedron vertices = %d, want %d", got, want)
	}
	for vertex, faces := range vertices {
		if len(faces) != 3 || vertexCorners[vertex] != 3 {
			t.Fatalf("vertex class %d contains %d corners from %d faces, want three distinct face corners", vertex, vertexCorners[vertex], len(faces))
		}
	}

	seen := [hexFaceCount + seamCount]bool{true}
	queue := []uint8{0}
	for len(queue) > 0 {
		face := queue[0]
		queue = queue[1:]
		if face < hexFaceCount {
			for _, edge := range world.faces[face].Edges {
				neighbor := uint8(edge.Face)
				if edge.Kind == SeamEdge {
					neighbor = hexFaceCount + uint8(edge.Seam)
				}
				if !seen[neighbor] {
					seen[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		} else {
			for _, neighbor := range world.seams[face-hexFaceCount].Faces {
				if !seen[neighbor] {
					seen[neighbor] = true
					queue = append(queue, uint8(neighbor))
				}
			}
		}
	}
	for face, visited := range seen {
		if !visited {
			t.Fatalf("face %d is disconnected", face)
		}
	}

	faces := len(world.faces) + len(world.seams)
	if got := len(vertices) - edges + faces; got != 2 {
		t.Fatalf("Euler characteristic = %d, want 2", got)
	}
}

func TestMoveAcrossFaceEdgesIsReversible(t *testing.T) {
	world, _ := NewWorld(3)
	for _, face := range world.faces {
		for d, edge := range face.Edges {
			if edge.Kind != HexEdge {
				continue
			}
			for position := 0; position <= world.radius; position++ {
				start := Cell{Face: face.ID, Hex: boundaryCoord(LocalDirection(d), position, world.radius)}
				next, err := world.Move(start, LocalDirection(d))
				if err != nil {
					t.Fatalf("Move from face %d edge %d position %d: %v", face.ID, d, position, err)
				}
				got, err := world.Move(next, edge.Entry)
				if err != nil || got != start {
					t.Fatalf("round trip from face %d edge %d position %d = %#v, %v; want %#v", face.ID, d, position, got, err, start)
				}
			}
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
