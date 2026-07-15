# Climate Generation on a Polyhedral World

> **Status: Exploratory essay**
>
> **Scope:** This is a possible client-side simulation built on a future,
> correct Wrex neighbor graph. Climate generation is not implemented by Wrex
> and is not part of its current API contract.

Traditional rectangular hex maps can model moisture by sweeping east-to-west
across rows. A sphere-like world has no single global row orientation, so a
client could instead define climate as movement over the same neighbor graph
used for navigation.

The sections below describe that proposal, not existing package behavior.

## Moisture Packets

A client could represent air as moisture packets carrying humidity,
temperature, and energy. Packets could originate over oceans or other moisture
sources and move by repeatedly following a preferred outbound bearing.

At each step:

1. Move to the neighboring cell indicated by the local bearing.
2. Lose moisture while climbing.
3. Warm and dry while descending.
4. Recharge humidity over open water.
5. Stop when humidity or energy is exhausted.

With a correct neighbor graph and public neighbor operation, packets would
cross internal face boundaries without client-side face handling.

## Local Rather Than Global

The simulation would use world-relative bearings rather than assuming global
rows. Face-local interpretation is an internal Wrex concern, not state that the
client should need to store.

## Mountains

Given client-owned elevation and terrain data, rain shadows could emerge from
moist air precipitating while climbing and warming while descending.

## Seams

Hidden irregular regions could become boundary conditions rather than playable
terrain. A client might inject cold, warmth, humidity, dryness, or setting-
specific effects into adjacent cells.

## Unified Design

Climate, rivers, migration, and trade routes could all consume the same Wrex
neighbor graph. Their data and rules would remain client responsibilities;
Wrex would supply only cell identity and topology/navigation primitives.
