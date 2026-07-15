# Wrex

Wrex is an experimental, game-agnostic Go package for defining and navigating
a finite, sphere-like world of regular hexagonal cells. It is intended for
games that need a lightweight global hex grid and can tolerate a small, fixed
number of inaccessible irregular locations.

The project began for a fantasy game, but fantasy rules and terminology are not
part of the package. Wrex supplies cell identity and topology/navigation
operations. Applications own persistence, terrain, entities, movement costs,
and all other game state.

A true sphere cannot be tiled entirely with regular hexagons: some non-hexagonal
regions are mathematically unavoidable. Wrex's target abstraction exposes every
regular hexagon as a playable cell and hides six non-standard regions behind
navigation errors. The package does not currently provide geometric vertices
or a renderer, but that does not mean its intended topology is unrelated to a
sphere or can never support rendering.

> **Experimental status:** the current API is a partial implementation of this
> model. The topology is a closed polyhedron, but do not yet rely on compass
> bearings converging at the poles. See [Current status](#current-status).

The module path is `github.com/maloquacious/wrex`.

## Scope

### Target client model

Wrex is intended to let a client:

- enumerate every playable cell as a stable exported `CellID`;
- ask for the neighboring cell in one of six world-relative bearings;
- receive a defined error when that step encounters a hidden non-standard
  region; and
- build storage, simulation, and game rules on top of that cell graph.

Face-local coordinates, face boundaries, and edge transforms are needed to
implement the graph, but are not intended to be the primary client model.

### Non-goals

Wrex does not own:

- client persistence, maps, entities, or other game state;
- fantasy-game-specific rules, names, terrain, climate, or movement costs;
- pathfinding or simulation policy;
- a claim that a true sphere can consist exclusively of regular hexagons; or
- currently, Euclidean geometry, mesh generation, projection, or rendering.

## Core model and terminology

| Term | Meaning |
|---|---|
| **Cell** | One playable regular hexagon. In the target API clients identify it only by `CellID`; the current API also exports its implementation address as `Cell{Face, Hex}`. |
| **Cell ID** | A stable value identifying one playable cell. It is suitable for application records and references, subject to the current encoding defect described below. |
| **Neighbor** | The location reached by one topological step from a cell in a bearing. A requested neighbor may instead be a hidden irregular point, in which case navigation fails with a defined error. |
| **Bearing** | One of six world-relative orientation sectors. Unlike a face-local direction, its meaning is intended to remain consistent across face boundaries. |
| **Face** | An implementation patch containing a finite axial hex map. Faces partition cells so the topology can be represented compactly; they are not intended as gameplay maps. |
| **Edge** | One side of an implementation face. Crossing it either transforms the address onto an adjacent face or encounters an inaccessible region. An edge is not the side of an individual cell in this documentation. |
| **Seam** | Current topology metadata representing one inaccessible square region and the face edges incident to it. A seam contains no playable cells. Future APIs may hide this representation completely. |
| **Irregular point** | The client-level concept for an unavoidable non-hexagonal location. It is not a playable cell and has no `CellID`; attempting to enter it must return an error. The current implementation represents each one as a square `Seam`. |

The intended model uses 24 playable hexagonal implementation faces and six
hidden square irregular regions. The choice of square rather than pentagonal
defects is explained by [ADR 0001](docs/adr/0001-use-square-defects.md). That is
design rationale, not evidence that the current face graph is correct.

## Current API

### Creating a world

```go
world, err := wrex.NewWorld(3)
if err != nil {
    panic(err)
}

fmt.Println(world.Radius())      // 3
fmt.Println(len(world.Faces())) // 24 implementation faces
fmt.Println(len(world.Seams())) // 6 inaccessible regions
fmt.Println(world.CellCount())  // 888
```

Supported radii are `MinRadius` through `MaxRadius`, currently 3 through 300.
Radius `r` means a face center plus `r` complete rings, or
`1 + 3r(r + 1)` cells per face. `NewWorld` returns an error wrapping
`ErrInvalidRadius` for an unsupported radius.

Always construct a world with `NewWorld`. A zero-value `World` is invalid:
cell and navigation operations return `ErrInvalidWorld`, `Contains` returns
false, and its face and seam collections are empty.

### Transitional cell addressing and IDs

The current implementation publicly exposes its face-local address:

```go
cell := wrex.Cell{
    Face: 7,
    Hex:  wrex.Coord{Q: 12, R: -4},
}

id, err := world.EncodeCell(cell)
sameCell, err := world.DecodeCell(id)
```

`CellID` is a packed `uint32` containing a face and biased axial `q` and `r`
coordinates. IDs do not include the world radius, so decoding validates the ID
against the receiving world. Reserved high bits must be zero.

This coordinate/face API is a mismatch with the intended ID-only client model:
the package does not yet enumerate IDs directly, and navigation currently
requires decoding IDs and using implementation types. Encoding validates the
face-local coordinate before packing it, including extreme integer values that
cannot be represented by the ID format.

### Transitional navigation

`World.Move` accepts a `Cell` and a face-local `LocalDirection`. An in-face step
adds the corresponding axial vector; an edge step applies the stored face-edge
transform. A step toward an inaccessible seam returns an error wrapping
`ErrImpassableSeam`.

The root package also exposes neutral `Bearing0` through `Bearing5`. Because a
bearing is world-relative rather than an axial vector, clients of the current
API must convert it on the cell's current face before every step:

```go
local, err := world.LocalDirectionFor(cell.Face, wrex.Bearing0)
if err != nil {
    return err
}
next, err := world.Move(cell, local)
if errors.Is(err, wrex.ErrImpassableSeam) {
    // The requested neighbor is a hidden non-standard region.
}
```

This is not yet the target “neighbor by `CellID` and bearing” operation. The
graph is the dual of a class-III (2,1) octahedral subdivision: its 24 hexagons
and six square defects form a closed spherical polyhedron.

## Orientation and optional compass semantics

The root package keeps bearings neutral so applications can assign their own
world-relative meaning. The optional `github.com/maloquacious/wrex/compass`
package interprets them clockwise as:

| Core bearing | Compass interpretation |
|---|---|
| `Bearing0` | `compass.North` |
| `Bearing1` | `compass.Northeast` |
| `Bearing2` | `compass.Southeast` |
| `Bearing3` | `compass.South` |
| `Bearing4` | `compass.Southwest` |
| `Bearing5` | `compass.Northwest` |

The compass package designates seam 0 as the north pole and seam 3 as the south
pole. Layout controls only how these bearings are drawn:

- for **flat-top** hexagons, north is the single upward direction;
- for **pointy-top** hexagons, northwest and northeast are the two upward
  directions.

Those screen directions do not replace world-relative semantics. In the target
model, repeatedly choosing north should approach the north irregular point, and
choosing south from near the south pole should continue toward that pole—not
turn away because of a face-local drawing orientation. The current orientation
assignment does not reliably provide this behavior; see
[issue #2](https://github.com/maloquacious/wrex/issues/2).

The separation between neutral bearings and optional compass names is recorded
in [ADR 0002](docs/adr/0002-separate-local-directions-from-global-bearings.md)
and [ADR 0003](docs/adr/0003-move-compass-semantics-to-child-package.md).

## Current status

### Implemented and tested as code behavior

- construction for radii 3 through 300 and deterministic cell counts;
- compact `Cell`, `Coord`, and 32-bit `CellID` value types;
- validation and encode/decode round trips for ordinary valid cells;
- local movement within a face and configured edge transitions;
- `ErrImpassableSeam` at configured blocked edges;
- a closed 24-hexagon / 6-square topology with spherical Euler characteristic
  and complete three-face vertex cycles;
- neutral bearing/local-direction conversion; and
- optional compass names and polar seam lookup.

The topology tests verify reciprocal joins, its complete rotation system,
three-face vertex cycles, and Euler characteristic.

### Known correctness and API gaps

- [#2: compass bearings do not reliably reach their designated poles](https://github.com/maloquacious/wrex/issues/2).
- The target ID-only enumeration and neighbor-by-bearing API is not yet
  implemented; face, coordinate, edge, and local-direction details remain
  exported.

## Proposed and exploratory work

The following are directions, not package guarantees:

- **Core Wrex candidates:** a correct closed topology, direct ID enumeration,
  neighbor-by-ID-and-bearing navigation, and possibly topology-aware distance
  or path primitives if they can remain policy-free.
- **Client functionality:** terrain storage, weighted pathfinding, climate,
  rivers, migration, trade, and game movement rules should ordinarily be built
  over Wrex's neighbor graph by applications or separate packages.
- **Exploratory essay:** [Climate Generation on a Polyhedral World](docs/climate-generation.md)
  sketches one possible client simulation. It is not implemented by Wrex.

Architecture decision records in [`docs/adr`](docs/adr) preserve design history
and rationale. This README, not every historical statement in an ADR, describes
the current product contract and implementation status.
