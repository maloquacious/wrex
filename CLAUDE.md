# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
go build ./...              # build all packages
go test ./...              # run all tests (root + compass)
go test -run TestNewWorld  # run a single test by name
go vet ./...               # static checks
```

There is no separate lint step or Makefile; the toolchain is plain `go`. Go 1.26.4 (see `go.mod`). The only external dependency is `github.com/maloquacious/semver`.

## What this package is

Wrex is a game-agnostic Go library for navigating a finite, sphere-like world of regular hexagonal cells. It provides **cell identity and topology/navigation only** — applications own persistence, terrain, entities, movement costs, and all game rules. There is no binary, server, or renderer; it is imported as a library (`github.com/maloquacious/wrex`).

The README is the authoritative product contract and the source of truth for the *current* vs. *intended* API. Much of the exported surface is explicitly transitional and marked as such — do not assume an exported type is the intended long-term client model. When the README and an ADR disagree, the README wins.

## Architecture

The whole world is derived from two hardcoded tables in `world.go`; nearly everything else is computed from them.

- **`faceTopology` and `seamTopology` (`world.go`)** are the authoritative spherical rotation system: the dual of a class-III (2,1) octahedral subdivision — **24 hexagonal faces + 6 square "seams"**. A true sphere cannot be tiled with only regular hexagons, so the six unavoidable non-hexagonal regions are modeled as square *seams* that are inaccessible (no playable cells, no `CellID`). Everything topological — edge transforms, reciprocity, Euler characteristic — is a consequence of these tables. Editing them changes the world; the topology tests (`TestTopologyIsClosedSphere`, `TestTopologyCountsAndReciprocity`) guard their invariants.
- **A `World` (`NewWorld(radius)`)** materializes those tables via `initTopology()`. Each of the 24 faces is a local axial hex map of the given radius (`1 + 3r(r+1)` cells/face). **A zero-value `World` is invalid** — every operation guards with `w.valid()` and returns `ErrInvalidWorld` (or `nil`/`false`/`0`). Always construct with `NewWorld`.
- **Two coordinate frames, kept deliberately separate:**
  - `LocalDirection` (`Dir0..Dir5`) — face-local axial directions, counterclockwise, used by `Move`.
  - `Bearing` (`Bearing0..Bearing5`) — world-relative sectors, clockwise, with **no compass meaning in the core package**. Each face stores a `Bearing0` local direction (assigned by BFS from seam 0 in `assignBearing0Directions`) so `BearingFor`/`LocalDirectionFor` translate between the frames. This separation is the core design decision (ADRs 0002, 0003) — keep compass/geographic naming out of the root package.
- **Movement (`Move`)** adds an axial delta for in-face steps; at a face boundary it applies the stored edge transform (`Edge.Entry` + `Reverse`) to land on the adjacent face, or returns a typed `*ImpassableSeamError` (wrapping `ErrImpassableSeam`) when the step hits a seam.
- **Convergent routing is per-cell, not per-face.** No single local direction per face yields a convergent route over the curved topology. `DirectionTowardSeam` runs a reverse BFS from a seam's edges over the whole cell graph (`buildSeamRoute`), caches the result per seam (`seamRouteCache`), and must be **recomputed at every cell** while navigating. `Move`'s own `moveValid` is the edge oracle that BFS walks, so routing and movement can never disagree.
- **`CellID`** is a packed `uint32` (`R|Q` biased by `MaxRadius`, then `FaceID`; high 7 bits reserved and must be zero). It does **not** encode radius, so `DecodeCell` validates against the receiving world. `EncodeCell` validates the coordinate (including extreme int values) before packing.

### Package layout

- `world.go` — everything above: types, topology tables, `World`, encoding, movement, routing, geometry helpers.
- `errors.go` — sentinel errors (`ErrInvalid*`, `ErrImpassableSeam`) and `ImpassableSeamError`.
- `compass/` — **optional child package** layering geographic names (North=`Bearing0`, clockwise) and polar routing (`DirectionTowardPole`, seam 0 = north pole, seam 3 = south pole) on top of the neutral core. Depends on the root, not vice versa.
- `internal/cerrs/` — `cerrs.Error`, a string-backed constant-error type used for sentinels.
- `docs/adr/` — architecture decision records (design history/rationale, not the current contract).

## Conventions

- **Constant sentinel errors** via `cerrs.Error`; wrap with `%w`. Typed errors (e.g. `ImpassableSeamError`) implement `Unwrap` so both `errors.Is` and `errors.As` work — preserve this when adding errors.
- **Integer-overflow safety is deliberate.** `Distance` saturates instead of overflowing and encoding rejects extreme coordinates; there are dedicated tests (`TestDistanceSaturatesOnOverflow`, `TestCellIDRejectsExtremeCoordinates`). Don't "simplify" these back into naive signed arithmetic.
- **Tests assert whole-world invariants**, not just examples: reversibility of every move, shortest-path routing from every cell, closed-sphere topology. When changing topology or movement, expect these exhaustive tests to be the real gate.
