---
phase: 2
slug: api-simulated-world
status: automated_gates_passed
nyquist_compliant: true
wave_0_complete: true
created: 2026-10-22
---

# Phase 2: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` plus `net/http/httptest`; no third-party test libraries |
| **Config file** | none |
| **Quick run command** | `go test ./internal/<pkg> -run <TestName> -v` |
| **Full suite command** | `go test ./... && go vet ./...` |
| **Estimated runtime** | ~10 seconds uncached |

API tests use a temp SQLite file per test (`t.TempDir()`), a fixed anchor (2026-10-22 12:00 UTC) and a pinned clock 90 minutes later, so results don't depend on when they run. `eachWorld` runs every API case twice: against `sim.NewWorld` and against the fake world in `fakeworld_test.go`.

---

## Sampling Rate

- **After every task:** the package's own tests
- **After the wave:** `go test ./...` and `go vet ./...`
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Plan | Requirement | Behavior | Automated Command | Status |
|------|-------------|----------|-------------------|--------|
| 02-01 | PLAT-04 | Same anchor gives the same world | `go test ./internal/sim -run TestDeterminism` | green |
| 02-01 | PLAT-04 | At least 25 fixtures; live, upcoming and completed present; live always has a score; every competition and format present; exactly one winnerless result | `go test ./internal/sim -run TestStateCoverageAtAnchor` | green |
| 02-01 | PLAT-04 | Every fixture team has an XI; player ids resolve | `-run TestTeamsAndPlayers` | green |
| 02-01 | PLAT-04 | No nil slices or maps from any method | `-run TestNoNulls` | green |
| 02-01 | PLAT-04 | 8 friends, at least 14 public fans; logs valid, ordered, never in the future, no unfilled placeholders | `-run TestPeopleAndActivity` | green |
| 02-01 | PLAT-04 | Crowd buckets have 10 entries; picks present | `-run TestCrowd`, `-run TestFriendsWatching`, `-run TestLiveScoresThroughout` | green |
| 02-01 | PLAT-03 | Demo history: 4 to 6 logs (2 classics), at least 5 predictions with an upset and an all-three call, 5 follows | `-run TestDemoHistory` | green |
| 02-01 | VAULT-06 | Every classic well formed, ordered, findable | `-run TestClassics` | green |
| 02-02 | PLAT-02 | Every table round-trips; demo seeds get fresh ids | `go test ./internal/store` (`TestStoreRoundTrip`, `TestSeedHistoryUsesFreshIDs`) | green |
| 02-02 | PLAT-03 | Bearer parsing; bootstrap resumes; 401 without a token; CORS preflight 204 | `TestBearerToken`, `TestBootstrapResumes`, `TestUnauthorized`, `TestCORSPreflight` | green |
| 02-02 | PRED-03 to 06 | Lock time, multiplier bands, points, upset, ties, streak carried in start order, badges | `go test ./internal/domain` (`TestIsLocked`, `TestStreakMultiplier`, `TestScoreOne`, `TestSettleHistoryOrdersByStartAndCarriesStreak`, `TestBadges`) | green |
| 02-02 | PRED-06, MATCH-06 | Buckets, averages, valid ratings, player averages and Fans' Player of the Match, watch and pick shares | `TestBucketsAndAverage`, `TestValidRating`, `TestPlayerAveragesAndPotm`, `TestWatchShares`, `TestShares` | green |
| 02-02 | PLAT-05 | Every GET returns its documented keys with no null arrays; bad state 400; unknown match 404 | `go test ./internal/httpapi -run TestEveryGET` | green |
| 02-02 | PLAT-04 | Sim plus real merge gives exact numbers on the fake world | `-run TestMergeRules` | green |

*Status: pending · green · red · flaky*

---

## Results (run 2026-10-22, uncached)

```
$ go test -count=1 ./...
?   	github.com/swathivallabhaneni289/innings/cmd/api	[no test files]
ok  	github.com/swathivallabhaneni289/innings/internal/auth	0.577s
ok  	github.com/swathivallabhaneni289/innings/internal/domain	1.134s
ok  	github.com/swathivallabhaneni289/innings/internal/httpapi	2.024s
ok  	github.com/swathivallabhaneni289/innings/internal/sim	2.896s
ok  	github.com/swathivallabhaneni289/innings/internal/store	2.310s

$ go vet ./...
(clean)
```

35 top-level test functions passed, 0 failed, 0 skipped (`go test -count=1 -v ./...`): auth 1, domain 10, httpapi 12, sim 10, store 2. The 11 `eachWorld` API tests each run a `sim` and a `fake` subtest.

---

## Wave 0 Requirements

- [x] `go test` runs with no external services (SQLite file per test)
- [x] Fake world for API tests while the sim was being built

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Server starts with no setup | PLAT-02 | Process-level | `make api`, then `curl localhost:8088/healthz` prints `ok` |
| A fresh database gets the migration once | PLAT-02 | File-level | `make reset-db && make api`, restart, confirm no migration error |

---

## Validation Sign-Off

- [x] All tasks have an automated verify
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** automated gates passed 2026-10-22
