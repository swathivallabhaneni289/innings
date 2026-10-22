---
phase: 3
slug: core-loop
status: automated_gates_passed
nyquist_compliant: false
wave_0_complete: true
created: 2026-10-22
---

# Phase 3: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | App: TypeScript compiler (strict); no Jest tests exist yet. Server rules behind these screens: Go `testing` with `httptest` |
| **Config file** | `tsconfig.json`; Jest config in `package.json` (unused) |
| **Quick run command** | `npx tsc --noEmit` |
| **Full suite command** | `npx tsc --noEmit && go test ./...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task:** `npx tsc --noEmit`
- **After the plan:** `go test ./internal/httpapi`
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task | Requirement | Behavior | Test Type | Automated Command | Status |
|------|-------------|----------|-----------|-------------------|--------|
| 1 | MATCH-01, MATCH-02 | `/v1/matches` answers for no filter and each state with `matches` and `fanRatings`; a bad state is 400 | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| 1 | MATCH-02 | Matches screens typecheck against `MatchesResponse` | static | `npx tsc --noEmit` | green |
| 2 | MATCH-03 to 10 | Match page for the first live, upcoming and recent match and the top-ranked match has `match`, `locked`, `split`, `mine` (prediction, outcome, log, watching), `friendsWatching`, `fans`, `results`; unknown id is 404 | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| 2 | MATCH-03 | Setting "I'm watching" shows on the match page; clearing it with `null` works | integration | `go test ./internal/httpapi -run TestWatching` | green |
| 3 | LOG-03 to 05, LOG-08 | Create a log, update it, stub number and id kept; a 4.2 or 0 rating, an unknown place, a player rated 11 and an unknown match are all 400; the log can then be liked, replied to and deleted (404 afterwards) | integration | `go test ./internal/httpapi -run TestLogCreateUpdateKeepsStub` | green |
| 3 | LOG-03 | Only 0.5 to 5 in half steps is valid | unit | `go test ./internal/domain -run TestValidRating` | green |
| 3, 4 | LOG-01 to 06, DESIGN-02, DESIGN-04 | Flow, ticket and motion components typecheck | static | `npx tsc --noEmit` | green |

*Status: pending · green · red · flaky*

---

## Results (run 2026-10-22)

```
$ npx tsc --noEmit
(no output)
tsc exit 0

$ go test -count=1 ./...
ok  	github.com/swathivallabhaneni289/innings/internal/httpapi	2.024s
(all five tested packages ok; full output in 02-VALIDATION.md)
```

---

## Wave 0 Requirements

- [ ] Jest unit tests for `src/features/matches/group.ts` (day grouping around midnight) and `src/features/log/draft.ts` (`toNewLog`, `draftFromLog`)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Match Page layout for upcoming, live, completed and classic | MATCH-03 to 10 | Visual | Open one of each from the Matches tab and the Vault |
| Live score moves without a pull | MATCH-04 | Needs time to pass | Leave a live match open for a minute |
| Log a fixture in four overs, a classic in three | LOG-01 to 05 | Gesture flow | Centre ball, pick, rate, rate players, place, Save & print |
| Ticket feeds out, stamp lands, success haptic | LOG-06, DESIGN-04 | Motion and haptics | Save a log on a device |
| Re-log pre-fills and keeps the stub number | LOG-08 | Flow | Log the same match again from its Match Page |
| Photo from the library | LOG-05 | OS permission sheet | Over 4, Add a photo |

Not recorded as a walkthrough. Carried into the Phase 6 screenshot QA (every screen, both themes).

---

## Validation Sign-Off

- [x] All tasks have an automated verify (typecheck or API test)
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [ ] `nyquist_compliant: true`: not set; app logic has no unit tests and the manual checks are pending

**Approval:** automated gates passed 2026-10-22; manual checks carried to Phase 6
