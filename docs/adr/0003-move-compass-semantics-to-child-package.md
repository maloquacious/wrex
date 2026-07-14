# ADR 0003: Move compass semantics to a child package

- Status: Accepted
- Date: 2026-07-14
- Supersedes: the API placement chosen in ADR 0002

## Context

ADR 0002 correctly separated face-local axial directions from world-relative
bearings, but the root `wrex` package still named bearings `North`, `South`, and
so forth. It also stored a field named `Face.North` and assigned polar roles to
seams.

Those names mixed two concerns:

1. the core topology needs six neutral orientation sectors and conversion to
   face-local movement vectors;
2. a particular game or presentation layer may interpret those sectors as
   compass directions and may designate two inaccessible seams as poles.

The second concern is optional. Other consumers may use the same topology for a
space station, elemental world, cube-relative axes, or a world with different
geographic terminology.

## Decision

The root package will expose only neutral concepts:

- `LocalDirection` values `Dir0` through `Dir5`;
- `Bearing` values `Bearing0` through `Bearing5`;
- `Face.Bearing0`, the local direction corresponding to world-relative
  `Bearing0`;
- `World.BearingFor` and `World.LocalDirectionFor` for conversion.

The root package will not export compass-named constants, seam roles, pole
lookups, or deprecated local aliases such as `East` and `Northeast`.

The child package `github.com/maloquacious/wrex/compass` will provide:

- compass names mapped onto the six neutral bearings;
- `NorthPoleSeam` and `SouthPoleSeam` conventions;
- conversion helpers phrased in compass terminology;
- polar seam lookup.

The topology continues to orient `Bearing0` as a shortest-path gradient toward
seam 0. The root package describes that only as a reference orientation. The
`compass` package interprets it as geographic north and seam 0 as the north
pole.

## Consequences

### Positive

- The core package remains geometric and topological rather than geographic.
- Applications can choose their own names for bearings.
- Compass and polar concepts are opt-in through a small child package.
- The API makes it difficult to confuse a local axial vector with a global
  geographic direction.
- Future child packages can provide alternative interpretations without
  changing movement code.

### Negative

- Applications wanting ordinary compass names must import a second package.
- Moving the names is a source-breaking change for users of the earlier
  experimental API.
- The neutral name `Bearing0` conveys less immediate meaning without the child
  package or documentation.

## Compatibility

This project is still experimental. The misleading deprecated compass aliases
are removed rather than retained, because retaining them would undermine the
separation this decision establishes.
