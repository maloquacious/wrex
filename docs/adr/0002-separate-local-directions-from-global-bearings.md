# ADR 0002: Separate local directions from global bearings

- Status: Accepted
- Date: 2026-07-14

## Context

Each playable region in Wrex is an independent axial hex map. The six axial
movement vectors are valid only inside that face's local coordinate frame. A
name such as `Northeast` is therefore ambiguous: it can mean the upper-right
vector in a drawing of one face, or a geographic bearing toward the world's
north pole. Those meanings diverge when movement crosses a polyhedral face
boundary.

The world also needs a geographic north pole. A conventional spherical
latitude/longitude system has a singularity at the pole because all meridians
converge there. Wrex already contains six inaccessible square defects, so that
singularity can be assigned to terrain that no player may occupy.

## Decision

Wrex will use two distinct direction systems.

1. `LocalDirection` values `Dir0` through `Dir5` represent axial coordinate
   operations in a face-local frame.
2. `Bearing` values represent the six global compass sectors: north, northeast,
   southeast, south, southwest, and northwest.

Each playable face stores which `LocalDirection` points most directly toward
the designated north-pole seam. `World.BearingFor` and
`World.LocalDirectionFor` convert between local directions and global bearings.
A caller following a global bearing must recompute the local direction after
crossing to another face.

Seam 0 is designated `NorthPole`, and seam 3 is designated `SouthPole`. Face
orientations are generated as a shortest-path gradient toward the north-pole
seam. On the four faces bordering that seam, the local north direction leads
directly into the inaccessible region and movement returns
`ErrImpassableSeam`.

The former compass-like local constants remain as deprecated aliases to avoid
unnecessary source breakage, but package implementation and documentation use
the neutral names.

## Consequences

### Positive

- Local coordinate math remains simple and uniform on every face.
- Global geography is explicit instead of being implied by a drawing
  convention.
- Repeated northward travel converges toward one well-defined polar region.
- The coordinate singularity is confined to inaccessible terrain.
- Rendering and user interfaces can display compass bearings without changing
  the movement engine.

### Negative

- A global bearing is not itself a movement vector; callers must convert it on
  the current face.
- The local direction corresponding to north may change at every face
  transition.
- Six-sector bearings are an abstraction and do not provide continuous angular
  headings.

## Movement at the pole

The north pole is not a cell. Several playable boundary routes approach
different edges of the square polar seam. Attempting to continue north is
blocked. To travel around the pole, a player chooses a different bearing and
follows playable face-to-face transitions around the inaccessible region.
