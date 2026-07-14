# wrex

`wrex` is an experimental Go package for a finite fantasy world built from a
closed **24-hexagon / 6-square polyhedral topology**.

The world has:

- 24 playable hexagonal regions;
- 6 inaccessible square regions;
- 30 polyhedral faces total.

This is the combinatorial form associated with an octahedral Goldberg-style
world at the `(2,1)` scale. `wrex` currently models adjacency and movement, not
Euclidean vertex coordinates or a renderable solid.

The six square faces are permanently inaccessible terrain. In the package they
are represented as **seams**: topology metadata that blocks movement, not maps
that can contain players.

Architecture decisions are recorded in [`docs/adr`](docs/adr).

The Go module path is:

```text
github.com/maloquacious/wrex
```

## Creating a world

```go
package main

import (
    "fmt"

    "github.com/maloquacious/wrex"
)

func main() {
    world, err := wrex.NewWorld(3)
    if err != nil {
        panic(err)
    }

    fmt.Println(world.Radius())      // 3
    fmt.Println(len(world.Faces())) // 24
    fmt.Println(len(world.Seams())) // 6
    fmt.Println(world.CellCount())  // 888
}
```

`NewWorld(radius int)` returns `(*World, error)`. The radius range remains:

```text
3 <= radius <= 300
```

The upper bound is a round, practical limit slightly above the Earth-area
equivalent radius of approximately 286 when each cell has a 5 km apothem.

The package exports `MinRadius` and `MaxRadius`. An out-of-range value returns
an error wrapping `ErrInvalidRadius`.

## Radius and cell counts

The radius convention is:

- radius 3 means the center plus three complete rings;
- radius `r` means the center plus `r` complete rings.

One playable face contains:

```text
1 + 3r(r + 1)
```

At radius 3:

```text
1 + 3(3)(4) = 37 cells per face
24 × 37 = 888 playable cells
```

The seams contain no cells and do not contribute to the count.

`CellCount` returns `int64`. `World` does not allocate every possible cell; it
stores only the radius, 24 face descriptors, and 6 seam descriptors. Cells and
`CellID` values are compact values created as needed.

## Cell addresses and stable IDs

A playable location combines a face ID with a coordinate local to that face:

```go
type Cell struct {
    Face FaceID
    Hex  Coord
}
```

There is no valid `Cell` on a square seam.

For storage, APIs, events, URLs, and foreign keys, Wrex also provides a stable
32-bit identifier:

```go
type CellID uint32
```

Use the world to encode and decode IDs:

```go
cell := wrex.Cell{
    Face: 7,
    Hex:  wrex.Coord{Q: 12, R: -4},
}

id, err := world.EncodeCell(cell)
if err != nil {
    // errors.Is(err, wrex.ErrInvalidCell)
    panic(err)
}

sameCell, err := world.DecodeCell(id)
if err != nil {
    // errors.Is(err, wrex.ErrInvalidCellID)
    panic(err)
}
```

### CellID bit layout

`CellID` packs `(face, q, r)` directly. It is deterministic and requires no
registry or database lookup.

```text
31             25 24          20 19          10 9            0
+----------------+--------------+--------------+--------------+
| reserved: 0    | face (5 bit) | q + 300      | r + 300      |
| 7 bits         |              | 10 bits      | 10 bits      |
+----------------+--------------+--------------+--------------+
```

The fields are:

| Bits | Width | Meaning |
|---:|---:|---|
| 0–9 | 10 | `r + 300`, in the encoded range 0–600 |
| 10–19 | 10 | `q + 300`, in the encoded range 0–600 |
| 20–24 | 5 | playable face ID, 0–23 |
| 25–31 | 7 | reserved; currently required to be zero |

Ten bits can represent 0–1023, which comfortably contains the biased coordinate
range 0–600. Five bits can represent 0–31, which contains all 24 playable
faces. Only 25 of the 32 bits are currently used.

The encoding is stable across world radii. A cell that exists in both a
radius-3 and radius-300 world receives the same ID. The radius is therefore not
part of the ID; decoding validates that the represented coordinate exists in
the particular `World` used for decoding.

Not every possible 25-bit pattern is a valid cell. Decoding rejects:

- face values 24–31;
- axial coordinates outside the world's radius;
- coordinate pairs outside the hexagonal face even when each component is in
  range;
- IDs with any reserved high bit set.

This strict validation leaves the high seven bits available for a future,
explicitly versioned extension without silently reinterpreting old data.

### Why 32 bits are sufficient

At the maximum radius of 300, each face contains:

```text
1 + 3(300)(301) = 270,901 cells
```

Across 24 playable faces:

```text
24 × 270,901 = 6,501,624 cells
```

A sequential ID would already fit easily in a `uint32`. Packing the coordinates
instead gives the ID useful structure while remaining compact. Use `uint64`
only at an external boundary that standardizes all identifiers on 64 bits; it
is not required by Wrex's address space.

## Local coordinates and directions

Each playable face uses its own axial coordinate frame `(q, r)`. The omitted
cube coordinate is:

```text
s = -q - r
```

The center is `(0, 0)`. The six coordinate operations are deliberately neutral:

| Local direction | Δq | Δr | Conventional drawing only |
|---|---:|---:|---|
| `Dir0` | +1 | 0 | right |
| `Dir1` | +1 | -1 | upper right |
| `Dir2` | 0 | -1 | upper left |
| `Dir3` | -1 | 0 | left |
| `Dir4` | -1 | +1 | lower left |
| `Dir5` | 0 | +1 | lower right |

The values proceed counterclockwise around one face. They are coordinate
operations, not compass directions. Movement within a face is vector addition:

```text
next.q = current.q + Δq
next.r = current.r + Δr
```

The root package intentionally exports no constants named `North`, `East`, or
similar. A local direction can correspond to a different geographic bearing on
each polyhedral face.

## Neutral world-relative bearings

The root package also defines six neutral world-relative sectors:

```go
type Bearing uint8

const (
    Bearing0 Bearing = iota
    Bearing1
    Bearing2
    Bearing3
    Bearing4
    Bearing5
)
```

A face stores the local direction corresponding to `Bearing0`:

```go
type Face struct {
    ID       FaceID
    Bearing0 LocalDirection
    Edges    [6]Edge
}
```

The remaining bearing-to-direction relationships follow by 60-degree rotation.
Use the root package conversion methods when an application wants a neutral
world-relative frame without assigning compass semantics:

```go
local, err := world.LocalDirectionFor(cell.Face, wrex.Bearing0)
bearing, err := world.BearingFor(cell.Face, wrex.Dir2)
```

A bearing is not itself an axial movement vector. Recompute the local direction
after crossing to another face.

## Compass child package

Geographic interpretation lives in the optional child package:

```text
github.com/maloquacious/wrex/compass
```

It adopts this convention:

| Compass name | Core bearing |
|---|---|
| `compass.North` | `wrex.Bearing0` |
| `compass.Northeast` | `wrex.Bearing1` |
| `compass.Southeast` | `wrex.Bearing2` |
| `compass.South` | `wrex.Bearing3` |
| `compass.Southwest` | `wrex.Bearing4` |
| `compass.Northwest` | `wrex.Bearing5` |

Example:

```go
import (
    "errors"

    "github.com/maloquacious/wrex"
    "github.com/maloquacious/wrex/compass"
)

func moveNorth(world *wrex.World, cell wrex.Cell) (wrex.Cell, error) {
    local, err := compass.LocalDirection(world, cell.Face, compass.North)
    if err != nil {
        return cell, err
    }

    next, err := world.Move(cell, local)
    if errors.Is(err, wrex.ErrImpassableSeam) {
        return cell, err // the polar region or another blocked square seam
    }
    return next, err
}
```

The child package also identifies:

```go
compass.NorthPoleSeam // seam 0
compass.SouthPoleSeam // seam 3
```

and provides `compass.Pole` to retrieve those seam descriptors.

### Why compass is separate

The core movement engine needs only topology, axial vectors, and orientation
sectors. Terms such as north pole, northeast, and southwest are worldbuilding
or presentation choices. Keeping them in a child package means another game can
reuse Wrex while assigning different meanings to `Bearing0` through `Bearing5`,
or no geographic meaning at all.

The inaccessible square seam absorbs the polar singularity. Repeatedly
recomputing and following `compass.North` converges toward the north-pole seam.
The pole is not a cell, so movement into it returns `ErrImpassableSeam`. Players
must turn and navigate around it.

