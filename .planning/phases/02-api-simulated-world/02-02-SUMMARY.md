---
phase: 02-api-simulated-world
plan: 02
subsystem: api
tags: [go, net/http, sqlite, modernc, guest-auth, scoring]

requires:
  - phase: 01-foundation
    provides: internal/domain types and docs/API.md
provides:
  - cmd/api server on port 8088
  - SQLite store with embedded migration 0001
  - Guest bootstrap and Bearer middleware
  - Scoring, settlement, badges and fan aggregates
  - Every read and write route in docs/API.md
affects: [phase-03, phase-04, phase-05]

tech-stack:
  added: [modernc.org/sqlite v1.61.0]
  patterns: ["per-request view merging sim and store data", "store errors panic as storeErr and become JSON 500s in one recoverer", "tests run each case against the real sim and a fake world"]

key-files:
  created: [cmd/api/main.go, internal/store/store.go, internal/store/migrations/0001_init.sql, internal/auth/auth.go, internal/domain/scoring.go, internal/domain/aggregate.go, internal/httpapi/server.go, internal/httpapi/view.go, internal/httpapi/reads.go, internal/httpapi/writes.go, Makefile]
  modified: [go.mod, go.sum]

requirements-completed: [PLAT-02, PLAT-03, PRED-03, PRED-04, PRED-05, PRED-06]

completed: 2026-10-22
status: complete
---

# Phase 2 Plan 2: Store, Auth, Scoring and HTTP API Summary

**A stdlib Go API with SQLite storage and guest accounts that serves the full contract, merges simulated and real activity per request, and scores predictions exactly as the Build decisions table says.**

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented
- **Commits:** `6743d09` (scoring and aggregates), `a8900c2` (store, guest auth and routes) (committed together on 2026-10-22, after Phase 6)

## Accomplishments

- **Server** (`cmd/api/main.go`): `PORT` (default 8088), `DB_PATH` (default `./innings.db`), optional `INNINGS_ANCHOR` to pin the sim clock; the sim clock then runs in real time. Graceful shutdown. `Makefile` targets: `api`, `app` (Metro on 8082), `test`, `vet`, `reset-db`.
- **Store** (`internal/store`): nine tables (users, logs, predictions, follows, likes, there_too, replies, top5, watching). Simulated ids are stored without foreign keys, so a real guest can like or reply to a simulated fan's stub.
- **Auth** (`internal/auth`): bootstrap creates "Guest Fan" with handle `fan-xxxxxx`, a 48-character hex token and demo history; middleware answers 401 JSON for a missing or unknown token.
- **Scoring** (`internal/domain/scoring.go`): `IsLocked`, `StreakMultiplier`, `ScoreOne`, `SettleHistory` (streaks carried across a user's predictions in match start order), `SummariseHistory` (accuracy ignores ties), and five badges.
- **Aggregates** (`internal/domain/aggregate.go`): half-ball buckets and averages, player averages sorted best first, `FansPotm` with `MinPotmRatings = 3`, watch shares, prediction shares.
- **Routes** (`internal/httpapi`): all 12 reads, 10 writes and `/healthz` from `docs/API.md`. Notable rules: Your Call prefers matches in the next three days and ranks India, then World Cup, then other internationals, then leagues; the Bragging Board covers the caller and everyone they follow over the last 7 days; notifications collect locks within 3 hours, friends' logs from the last 3 days, points scored, and replies and likes on my stubs (at most 50).
- **Validation on writes:** unknown match 400; logging an upcoming match 400 `not-started`; rating must be a half step from 0.5 to 5; player ratings 1 to 10 and only for players in the XI; review at most 2000 characters, place note at most 200; prediction 409 `locked` after the first ball and 400 if a pick isn't in the match; name 1 to 40 characters, bio at most 160; Top 5 at most five, no duplicates.

## Decisions Made

- The view loads whole tables per request. At demo size this is simpler and fast enough; `view.go` notes that it becomes scoped queries at scale.
- The upset share uses crowd picks, simulated friends' picks and real users' picks made before the start, so it is fixed once the match locks.

## Deviations from Plan

- **Fake world for parallel work:** the sim workstream (02-01) and this one ran at the same time, so `fakeworld_test.go` provides a small world with known numbers. `eachWorld` runs every API test against both; `TestMergeRules` pins exact values (fan rating 3.85 from 13 ratings, Fans' Player of the Match differing from the official pick, review order by likes).

## Issues Encountered

- Port 8080 was taken by another local API, so the default became 8088 (also in `src/data/client.ts`).

## Next Phase Readiness

The app can bootstrap a guest and read every screen's data. `DELETE /v1/logs/{id}` is implemented but not yet used by any screen.

## Self-Check: PASSED

`go test ./...` and `go vet ./...` pass (see `02-VALIDATION.md`).
