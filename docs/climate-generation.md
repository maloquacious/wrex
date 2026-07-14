# Climate Generation on a Polyhedral World

Traditional rectangular hex maps model moisture by sweeping east-to-west across rows. Wrex intentionally abandons the idea of global rows because the world is polyhedral and every face has its own local coordinate system.

Instead, climate is modeled as movement over the same neighbor graph used by players.

## Moisture Packets

Represent air as moisture packets carrying humidity, temperature, and energy. Packets originate over oceans or other moisture sources and move by repeatedly following the preferred outbound bearing of the current cell.

At each step:

1. Move to the neighboring cell indicated by the local bearing.
2. Lose moisture while climbing.
3. Warm and dry while descending.
4. Recharge humidity over open water.
5. Stop when humidity or energy is exhausted.

Because movement follows the graph, packets naturally cross face boundaries without any special-case code.

## Local Rather Than Global

No cell knows where 'east' is globally. Each cell simply knows which neighboring cell is next in the prevailing circulation. Bearings are interpreted locally by each face.

## Mountains

Rain shadows emerge naturally. Moist air climbing mountains precipitates. Descending air warms and dries, producing deserts without explicit biome rules.

## Seams

The inaccessible seams become boundary conditions rather than playable terrain. A seam may inject cold, warmth, humidity, dryness, or magical effects into adjacent cells.

## Unified Design

Movement, climate, rivers, migration, trade routes, and other simulations all operate over the same neighbor graph. The world generator therefore uses one fundamental abstraction for nearly every large-scale simulation.
