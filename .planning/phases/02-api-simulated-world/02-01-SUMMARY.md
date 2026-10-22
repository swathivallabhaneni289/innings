---
phase: 02-api-simulated-world
plan: 01
subsystem: simulation
tags: [go, simulation, deterministic, splitmix64, fixtures, live-scores]

requires:
  - phase: 01-foundation
    provides: internal/domain types
provides:
  - A deterministic cricket world anchored to server start
  - 29 fixtures, 65 teams, 374 players, 8 friends, 16 public fans
  - About 640 simulated logs visible at server start, with predictions, replies and crowd data
  - Demo history for new guests
affects: [02-02 API, every app screen]

tech-stack:
  added: []
  patterns: ["string-seeded splitmix64, never math/rand's global source", "precompute everything with timestamps once, filter by now per call", "templates that skip themselves when a placeholder can't be filled"]

key-files:
  created: [internal/sim/sim.go, internal/sim/rng.go, internal/sim/teams.go, internal/sim/players.go, internal/sim/venues.go, internal/sim/fixtures.go, internal/sim/live.go, internal/sim/people.go, internal/sim/activity.go, internal/sim/reviews.go, internal/sim/crowd.go, internal/sim/demo.go, internal/sim/vault.go, internal/sim/classics.go, internal/sim/sim_test.go]
  modified: []

requirements-completed: [PLAT-04]

completed: 2026-10-22
status: complete
---

# Phase 2 Plan 1: Simulated World Summary

**`internal/sim` builds the same believable cricket world from the same anchor every time: matches in every state around server start, live scores that move with the clock, and a population of friends, public fans and an anonymous crowd who log, predict and reply.**

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented
- **Commits:** none yet (uncommitted)

## Accomplishments

- **Fixtures** (`fixtures.go`): 29 scripted fixtures placed in minutes relative to the anchor: an India v Australia T20I series, Sri Lanka v England Tests, India v South Africa women's ODIs, a women's T20 World Cup through to the semi-finals, and IPL, WPL, BBL, SA20 and The Hundred games (24 T20, 2 ODI, 2 Test, 1 Hundred). Probed at server start: 3 live, 7 upcoming, 19 completed, with 3 finished in the last 24 hours. One match ends with no result; some carry an upset flag or a fans' favourite who differs from the official Player of the Match, so "Fans disagree!" and upset bonuses happen naturally.
- **Live play** (`live.go`): limited-overs innings interpolated ball by ball from the scripted final score, with "need X off Y" during a chase; Tests run 90 overs a day in three sessions.
- **Catalog** (`teams.go`, `players.go`, `venues.go`): 65 teams and 374 players, playing XIs in batting order with current players, plus the nicknames fans use in reviews.
- **People** (`people.go`): 8 friends a guest can follow and 16 public fans who write the popular reviews, each with a favourite team and a one-line bio.
- **Activity** (`activity.go`, `reviews.go`): every log, prediction and reply precomputed once with timestamps and filtered by `now`. 637 logs were visible on a fresh server; watch places follow fixed weights (home most common). Reviews are filled from templates with real names, runs, wickets and margins.
- **Crowd** (`crowd.go`): per-match rating buckets, player tallies, watch places and winner picks frozen at lock time, so the prediction split and the upset rule have a realistic base.
- **Demo history** (`demo.go`): a new guest starts with 4 to 6 logs (two of them Vault classics), at least 5 settled predictions including an upset and an all-three-correct call, and follows 5 of the 8 friends.
- **Vault loader** (`classics.go`, `vault.go`): embeds `classics.json` and converts each seed to a completed match; the seed itself is plan 05-01.

## How the numbers were checked

Counts above come from a throwaway server started from the same code on a scratch database (port 8097, removed afterwards): `/v1/matches?state=` for fixture states, `/v1/bootstrap` for teams and players, and `/v1/users/{id}` summed over the 24 simulated people for logs.

## Decisions Made

- `World` never mutates after `NewWorld`, so it is safe for concurrent reads without locks.
- IDs never depend on the anchor, so a match id in a test or a deep link stays valid across restarts.

## Deviations from Plan

None.

## Next Phase Readiness

`var _ World = (*sim.World)(nil)` in `internal/httpapi/server.go` compiles, so the API serves the real sim.

## Self-Check: PASSED

`go test ./internal/sim` passes (10 tests; see `02-VALIDATION.md`).
