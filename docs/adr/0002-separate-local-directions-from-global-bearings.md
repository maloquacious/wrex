# ADR 0002: Separate local directions from global bearings

- Status: Accepted; partially superseded by ADR 0003
- Date: 2026-07-14

## Supersession

[ADR 0003](0003-move-compass-semantics-to-child-package.md) supersedes the API
placement and compatibility decisions in this record. Compass-named bearings,
pole roles, and deprecated compass-like local aliases do **not** live in the
root package. The distinction between face-local directions and world-relative
bearings remains accepted.

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
2. `Bearing` values represent six neutral world-relative sectors. The optional
   `compass` child package names those sectors north, northeast, southeast,
   south, southwest, and northwest.

Each playable face stores which `LocalDirection` points most directly toward
the designated north-pole seam. `World.BearingFor` and
`World.LocalDirectionFor` convert between local directions and global bearings.
A caller following a global bearing must recompute the local direction after
crossing to another face.

The `compass` package designates seam 0 as its north pole and seam 3 as its
south pole. The intended orientation makes compass bearings converge on their
designated poles and returns `ErrImpassableSeam` when a step would enter one.
The current orientation does not reliably satisfy that behavior; see
[issue #2](https://github.com/maloquacious/wrex/issues/2).

The proposal to retain former compass-like local constants as deprecated
aliases was superseded by ADR 0003. Those aliases were removed while the API
was still experimental.

## Consequences

### Positive

- Local coordinate math remains simple and uniform on every face.
- Global geography is explicit instead of being implied by a drawing
  convention.
- The model can express travel toward a well-defined polar region once the
  orientation defect is corrected.
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

In the target model, the north pole is not a cell. Several playable boundary
routes approach different edges of the square polar seam. Attempting to
continue north is blocked. To travel around the pole, a client chooses a
different bearing and follows playable transitions around the inaccessible
region. This describes intended behavior, not a guarantee of the current
topology or orientation implementation.
