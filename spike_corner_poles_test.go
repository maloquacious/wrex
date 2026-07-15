package wrex

// Temporary spike for issue #9: verify that the topology contains
// three-face corner points suitable as pointy-top poles. Findings are
// recorded in docs/adr/0004; this file is deleted afterward.

import (
	"fmt"
	"sort"
	"testing"
)

// spikeCornerCell maps a face-corner side (numbering from
// TestTopologyIsClosedSphere's hexCorner) to the corner cell coordinate.
func spikeCornerCell(side, radius int) Coord {
	switch side {
	case 0:
		return Coord{Q: radius, R: -radius}
	case 1:
		return Coord{Q: radius, R: 0}
	case 2:
		return Coord{Q: 0, R: radius}
	case 3:
		return Coord{Q: -radius, R: radius}
	case 4:
		return Coord{Q: -radius, R: 0}
	case 5:
		return Coord{Q: 0, R: -radius}
	}
	panic("bad side")
}

type spikeVertex struct {
	cells [3]Cell
	sides [3]int
}

// spikeHexVertices returns the three-hex corner vertices via the same
// union-find rotation-system walk as TestTopologyIsClosedSphere.
func spikeHexVertices(t *testing.T, world *World) []spikeVertex {
	t.Helper()

	const cornerCount = hexFaceCount*6 + seamCount*4
	parents := [cornerCount]int{}
	for i := range parents {
		parents[i] = i
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

	for _, face := range world.faces {
		for d, edge := range face.Edges {
			if edge.Kind != HexEdge || face.ID > edge.Face {
				continue
			}
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
			var direction LocalDirection
			for d, edge := range world.faces[face].Edges {
				if edge.Kind == SeamEdge && edge.Seam == seam.ID {
					direction = LocalDirection(d)
					break
				}
			}
			seamStart := hexFaceCount*6 + int(seam.ID)*4 + side
			seamEnd := hexFaceCount*6 + int(seam.ID)*4 + (side+1)%4
			join(seamStart, hexCorner(face, direction, 1))
			join(seamEnd, hexCorner(face, direction, 0))
		}
	}

	groups := map[int][]int{}
	touchesSeam := map[int]bool{}
	for corner := 0; corner < cornerCount; corner++ {
		root := find(corner)
		if corner >= hexFaceCount*6 {
			touchesSeam[root] = true
			continue
		}
		groups[root] = append(groups[root], corner)
	}

	var vertices []spikeVertex
	for root, corners := range groups {
		if touchesSeam[root] {
			continue
		}
		if len(corners) != 3 {
			t.Fatalf("hex vertex %d has %d corners, want 3", root, len(corners))
		}
		var v spikeVertex
		for i, corner := range corners {
			face := FaceID(corner / 6)
			side := corner % 6
			v.sides[i] = side
			v.cells[i] = Cell{Face: face, Hex: spikeCornerCell(side, world.radius)}
		}
		vertices = append(vertices, v)
	}
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].cells[0].Face < vertices[j].cells[0].Face ||
			(vertices[i].cells[0].Face == vertices[j].cells[0].Face &&
				vertices[i].sides[0] < vertices[j].sides[0])
	})
	return vertices
}

// spikeDistances runs a BFS over the playable cell graph from the given
// seed cells and returns per-cell distances indexed by cellRouteIndex.
func spikeDistances(t *testing.T, world *World, seeds []Cell) []int32 {
	t.Helper()
	side := 2*world.radius + 1
	distance := make([]int32, hexFaceCount*side*side)
	for i := range distance {
		distance[i] = -1
	}
	queue := make([]uint32, 0, world.CellCount())
	for _, cell := range seeds {
		index := world.cellRouteIndex(cell)
		if distance[index] == -1 {
			distance[index] = 0
			queue = append(queue, uint32(index))
		}
	}
	for head := 0; head < len(queue); head++ {
		cell := world.routeCell(int(queue[head]))
		for d := Dir0; d <= Dir5; d++ {
			next, _, _, blocked, err := world.moveValid(cell, d)
			if err != nil {
				t.Fatalf("moveValid: %v", err)
			}
			if blocked {
				continue
			}
			index := world.cellRouteIndex(next)
			if distance[index] != -1 {
				continue
			}
			distance[index] = distance[queue[head]] + 1
			queue = append(queue, uint32(index))
		}
	}
	if int64(len(queue)) != world.CellCount() {
		t.Fatalf("BFS reached %d of %d cells", len(queue), world.CellCount())
	}
	return distance
}

func spikeSeamCells(world *World, seam SeamID) []Cell {
	var cells []Cell
	for _, face := range world.seams[seam].Faces {
		for d, edge := range world.faces[face].Edges {
			if edge.Kind != SeamEdge || edge.Seam != seam {
				continue
			}
			for position := 0; position <= world.radius; position++ {
				cells = append(cells, Cell{Face: face, Hex: boundaryCoord(LocalDirection(d), position, world.radius)})
			}
		}
	}
	return cells
}

func TestSpikeCornerPoles(t *testing.T) {
	world, err := NewWorld(3)
	if err != nil {
		t.Fatal(err)
	}

	vertices := spikeHexVertices(t, world)
	if len(vertices) != 32 {
		t.Fatalf("three-hex vertices = %d, want 32", len(vertices))
	}

	// The three cells at each vertex must be mutually adjacent and on
	// three distinct faces.
	for _, v := range vertices {
		faces := map[FaceID]bool{}
		for _, c := range v.cells {
			faces[c.Face] = true
		}
		if len(faces) != 3 {
			t.Fatalf("vertex %v spans %d faces, want 3", v, len(faces))
		}
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if i == j {
					continue
				}
				adjacent := false
				for d := Dir0; d <= Dir5; d++ {
					next, moveErr := world.Move(v.cells[i], d)
					if moveErr != nil {
						continue
					}
					if next == v.cells[j] {
						adjacent = true
						break
					}
				}
				if !adjacent {
					t.Fatalf("vertex cells %v and %v are not adjacent", v.cells[i], v.cells[j])
				}
			}
		}
	}

	// Group vertices by the multiset of corner sides. The octahedron
	// face-center orbit should be the one whose corners all use the same
	// side on every face.
	bySides := map[string][]int{}
	for i, v := range vertices {
		s := []int{v.sides[0], v.sides[1], v.sides[2]}
		sort.Ints(s)
		key := fmt.Sprint(s)
		bySides[key] = append(bySides[key], i)
	}
	t.Logf("side-multiset groups:")
	for key, members := range bySides {
		t.Logf("  sides %s: %d vertices", key, len(members))
	}

	// Seam-distance signature per vertex: distance from the vertex's three
	// cells to each of the six seams.
	seamDistance := make([][]int32, seamCount)
	for s := SeamID(0); s < seamCount; s++ {
		seamDistance[s] = spikeDistances(t, world, spikeSeamCells(world, s))
	}
	bySignature := map[string][]int{}
	for i, v := range vertices {
		signature := make([]int32, seamCount)
		for s := range signature {
			best := int32(1 << 30)
			for _, c := range v.cells {
				if d := seamDistance[s][world.cellRouteIndex(c)]; d < best {
					best = d
				}
			}
			signature[s] = best
		}
		sorted := append([]int32(nil), signature...)
		sort.Slice(sorted, func(a, b int) bool { return sorted[a] < sorted[b] })
		key := fmt.Sprint(sorted)
		bySignature[key] = append(bySignature[key], i)
		t.Logf("vertex %2d cells=%v sides=%v seam distances=%v", i, v.cells, v.sides, signature)
	}
	t.Logf("distance-signature groups:")
	for key, members := range bySignature {
		t.Logf("  %s: vertices %v", key, members)
	}

	// Candidate polar orbit: the group of 8 (if present).
	var polar []int
	for _, members := range bySignature {
		if len(members) == 8 {
			polar = members
		}
	}
	if polar == nil {
		t.Fatalf("no distance-signature group of size 8; groups: %v", bySignature)
	}
	sort.Ints(polar)

	// Opposite pairs within the polar orbit by cell-graph separation.
	vertexDistance := func(a, b int) int32 {
		distance := spikeDistances(t, world, vertices[a].cells[:])
		best := int32(1 << 30)
		for _, c := range vertices[b].cells {
			if d := distance[world.cellRouteIndex(c)]; d < best {
				best = d
			}
		}
		return best
	}
	t.Logf("polar orbit separations:")
	opposite := map[int]int{}
	for _, a := range polar {
		bestVertex, bestDistance := -1, int32(-1)
		for _, b := range polar {
			if a == b {
				continue
			}
			d := vertexDistance(a, b)
			t.Logf("  vertex %d -> vertex %d: %d", a, b, d)
			if d > bestDistance {
				bestVertex, bestDistance = b, d
			}
		}
		opposite[a] = bestVertex
	}
	for _, a := range polar {
		if opposite[opposite[a]] != a {
			t.Fatalf("vertex %d's farthest partner %d is not mutual (partner's farthest is %d)", a, opposite[a], opposite[opposite[a]])
		}
	}

	// Convergence: from every cell, a greedy step toward the pole vertex
	// exists (a neighbor strictly closer), so per-cell recomputation routes
	// every cell to one of the three polar cells.
	north := polar[0]
	south := opposite[north]
	for _, pole := range []int{north, south} {
		distance := spikeDistances(t, world, vertices[pole].cells[:])
		side := 2*world.radius + 1
		for index := 0; index < hexFaceCount*side*side; index++ {
			if distance[index] <= 0 {
				continue
			}
			cell := world.routeCell(index)
			if !world.Contains(cell) {
				continue
			}
			stepFound := false
			for d := Dir0; d <= Dir5; d++ {
				next, _, _, blocked, moveErr := world.moveValid(cell, d)
				if moveErr != nil {
					t.Fatal(moveErr)
				}
				if blocked {
					continue
				}
				if distance[world.cellRouteIndex(next)] == distance[index]-1 {
					stepFound = true
					break
				}
			}
			if !stepFound {
				t.Fatalf("no descending step from %v toward vertex %d", cell, pole)
			}
		}
	}

	// Qualitative: bearing sequences of greedy routes to the north corner
	// pole from the cells farthest from it.
	distance := spikeDistances(t, world, vertices[north].cells[:])
	var farthest Cell
	best := int32(-1)
	side := 2*world.radius + 1
	for index := 0; index < hexFaceCount*side*side; index++ {
		cell := world.routeCell(index)
		if !world.Contains(cell) {
			continue
		}
		if d := distance[index]; d > best {
			best = d
			farthest = cell
		}
	}
	t.Logf("north pole = vertex %d (cells %v), south pole = vertex %d, max distance = %d", north, vertices[north].cells, south, best)
	cell := farthest
	var bearings []Bearing
	for distance[world.cellRouteIndex(cell)] > 0 {
		moved := false
		for d := Dir0; d <= Dir5; d++ {
			next, _, _, blocked, moveErr := world.moveValid(cell, d)
			if moveErr != nil {
				t.Fatal(moveErr)
			}
			if blocked {
				continue
			}
			if distance[world.cellRouteIndex(next)] == distance[world.cellRouteIndex(cell)]-1 {
				b, bErr := world.BearingFor(cell.Face, d)
				if bErr != nil {
					t.Fatal(bErr)
				}
				bearings = append(bearings, b)
				cell = next
				moved = true
				break
			}
		}
		if !moved {
			t.Fatal("greedy route stalled")
		}
	}
	t.Logf("greedy route %v -> vertex %d bearing sequence: %v", farthest, north, bearings)
}
