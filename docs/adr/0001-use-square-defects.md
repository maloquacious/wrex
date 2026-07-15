# ADR 0001: Use square defects instead of pentagonal defects

- Status: Accepted (target design)
- Date: 2026-07-14

> This record explains the choice of irregular region. It does not establish
> that the current implementation realizes the intended closed topology. See
> [issue #1](https://github.com/maloquacious/wrex/issues/1).

## Context

A finite, closed world made primarily from hexagonal regions cannot be composed
only of regular hexagons. The topology needs a fixed amount of positive
curvature supplied by non-hexagonal faces.

Two practical Goldberg-style choices were considered:

1. an octahedral family using hexagons plus 6 squares; and
2. an icosahedral family using hexagons plus 12 pentagons.

In Wrex, non-hexagonal faces are permanently inaccessible terrain. Players may
approach one from a neighboring hexagonal region, but movement into it is
forbidden. The non-hexagonal faces therefore act as topological defects and
blocked seams rather than playable maps.

For a three-valent closed polyhedron:

- each square contributes 120 degrees of angular deficit, requiring 6 squares;
- each pentagon contributes 60 degrees of angular deficit, requiring 12
  pentagons.

Because each square has four edges and each pentagon has five, the two choices
produce different gameplay costs:

- squares: 6 inaccessible faces and 24 blocked face incidences;
- pentagons: 12 inaccessible faces and 60 blocked face incidences.

The pentagonal construction distributes curvature more evenly and produces a
rounder, more isotropic approximation of a globe. The square construction
concentrates curvature more strongly and produces a more visibly polyhedral
world.

## Decision

Wrex's target model will use an octahedral Goldberg-style topology made from
playable hexagonal regions and 6 inaccessible square defects.

The intended topology contains:

- 24 playable hexagonal faces;
- 6 inaccessible square faces represented in code as seams;
- 24 blocked hex-to-square face incidences.

Square seams are topology metadata. They contain no cells and cannot be valid
player locations.

## Rationale

Wrex is intended to provide a game-agnostic, sphere-like hex grid. It does not
need the most isotropic possible approximation, so the square construction
provides the smaller and simpler client-visible set of blocked locations:

- half as many inaccessible regions;
- 60 percent fewer blocked face incidences than the pentagonal alternative;
- fewer exceptional neighborhoods for movement and pathfinding clients;
- a simpler topology to explain, inspect, and test.

The reduced number of blocked interfaces is more important to gameplay than the
rounder appearance supplied by twelve pentagons.

## Consequences

### Positive

- Players encounter fewer impassable global defects.
- More region boundaries remain available for direct travel.
- Pathfinding has fewer blocked cross-face transitions.
- Clients can assign the six defects identities appropriate to their own
  setting.
- The topology remains compatible with additional octahedral Goldberg
  subdivisions that add playable hexagonal faces without adding square defects.

### Negative

- Curvature is concentrated into six stronger defects.
- The world is less sphere-like and less directionally uniform than an
  icosahedral pentagon-and-hexagon construction.
- Rendering the world as a rounded planet would show more distortion around the
  square defects.

### Reconsideration

This decision should be revisited only if non-hexagonal faces become playable,
or if visual approximation of a sphere becomes more important than minimizing
blocked travel boundaries.
