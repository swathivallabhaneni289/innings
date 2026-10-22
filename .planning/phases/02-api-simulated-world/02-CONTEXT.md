# Phase 2: Go API & simulated world - Context

**Gathered:** 2026-10-22
**Status:** Complete

<domain>
## Phase Boundary

The Go side of the app: a JSON API that serves every route in `docs/API.md`, backed by SQLite for real guests and by a deterministic simulated cricket world for everything else (fixtures, live scores, results, friends, public fans, the crowd). Also the scoring rules from the Build decisions table. No app screens.

Two workstreams ran in parallel against one interface: the simulation (`internal/sim`, plan 02-01) and the store, auth and HTTP layer (`internal/store`, `internal/auth`, `internal/httpapi`, `internal/domain` scoring and aggregates, plan 02-02). The API was tested against a small fake world (`internal/httpapi/fakeworld_test.go`) until the sim landed, and against both afterwards.

</domain>

<decisions>
## Implementation Decisions

### Backend shape
- **D-01:** Go with the standard library `net/http` router (method-and-path patterns), no framework. Middleware: CORS, request logging, panic recovery that turns store failures into a JSON 500.
- **D-02:** SQLite through pure-Go `modernc.org/sqlite` (no cgo), WAL mode, foreign keys on, migrations embedded with `go:embed`. The SQL avoids SQLite-only features so it ports to Postgres (TEXT timestamps, JSON as TEXT, `ON CONFLICT` upserts, no AUTOINCREMENT).
- **D-03:** Port 8088 by default (`PORT`), database `./innings.db` (`DB_PATH`). 8080 is taken by another local API.

### Accounts
- **D-04:** Guest accounts only. `POST /v1/bootstrap` with no token, or an unknown one, creates a guest and returns a 48-hex-character token; with a known token it resumes that guest. Every other route needs `Authorization: Bearer <token>` and answers 401 JSON otherwise.
- **D-05:** A new guest starts with demo history so no screen is empty on first open: a few logs (two of them Vault classics), settled predictions with at least one upset and one all-three-correct call, and follows of 5 of the 8 friends.

### Simulated world
- **D-06:** No cricket API key exists, so `internal/sim` generates the world. Fixture times are offsets from an anchor (server start, rounded down to the minute); the same anchor always builds the same world. `INNINGS_ANCHOR` pins it.
- **D-07:** Randomness comes from a splitmix64 generator seeded from strings (`internal/sim/rng.go`); `math/rand`'s global source is never used.
- **D-08:** Simulated rows and real rows never share a table. The API merges them per request (`internal/httpapi/view.go`): sim likes plus real likes, sim logs plus real logs, crowd picks plus friends' picks plus real picks.
- **D-09:** The real cricket API (CricketData.org or SportMonks) can replace `internal/sim` later behind the same `World` interface.

### Scoring (Build decisions table)
- **D-10:** Winner +10, top run-scorer +15, top wicket-taker +15, upset +10 when the winning pick had under 30% of fan picks at lock time.
- **D-11:** Streak multiplier on consecutive correct winners: 1 to 2 x1, 3 to 4 x1.5, 5 or more x2, applied to the whole match's points including the upset bonus. A wrong winner resets the streak; a tie or no result leaves it.
- **D-12:** Tied top scorers or wicket-takers all count. Fans' Player of the Match needs at least 3 ratings. Bragging Board sums points from matches finished in the last 7 days.

### Claude's Discretion
- Which fixtures, players and fan personas to simulate, and the review templates.
- Badge set: First Call, Called the Upset, 5-Match Streak, Sharpshooter (all three picks right in one match), Stadium Regular (3 stadium logs).

</decisions>

<canonical_refs>
## Canonical References

- `docs/PLAN.md` (Build decisions table)
- `docs/API.md` (routes, JSON rules)
- `src/domain/types.ts`, `src/domain/api.ts` (the contract the Go structs mirror)

</canonical_refs>

<code_context>
## Existing Code Insights

`internal/domain/types.go` and `internal/domain/api.go` exist from Phase 1. `internal/sim/sim.go` was written first as the contract between the two workstreams: exported method signatures with a note to keep them and fill the bodies.

</code_context>

<deferred>
## Deferred Ideas

- Email sign-in (v2, ACCT-01)
- Real cricket data (v2, DATA-01)
- Scoped queries: each request loads whole tables, which is fine at demo size and noted in `view.go` as the thing to change at scale

</deferred>

---

*Phase: 2-Go API & simulated world*
*Context gathered: 2026-10-22*
