---
phase: 4
slug: social-predictions
status: automated_gates_passed
nyquist_compliant: false
wave_0_complete: true
created: 2026-10-22
---

# Phase 4: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | App: TypeScript compiler (strict); no Jest tests yet. Rules behind the screens: Go `testing` with `httptest`, each API case run against the sim and the fake world |
| **Config file** | `tsconfig.json` |
| **Quick run command** | `npx tsc --noEmit` |
| **Full suite command** | `npx tsc --noEmit && go test ./...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task:** `npx tsc --noEmit`
- **After the plan:** `go test ./internal/httpapi ./internal/domain`
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task | Requirement | Behavior | Test Type | Automated Command | Status |
|------|-------------|----------|-----------|-------------------|--------|
| 1 | PRED-02 | A prediction on a started match gets 409 `locked` | integration | `go test ./internal/httpapi -run TestPredictionLocked` | green |
| 1 | PRED-01 | On an upcoming match: a winner not in the match is 400; a valid call saves and shows as `mine.prediction` | integration | `go test ./internal/httpapi -run TestPredictionUpcoming` | green |
| 1 | PRED-02 | `IsLocked` is true from the start time on, and for any non-upcoming match | unit | `go test ./internal/domain -run TestIsLocked` | green |
| 1 | MATCH-06, MATCH-07 | Fake world: fan rating 3.85 from 13 ratings; Fans' Player of the Match differs from the official one and `fansDisagree` is true; reviews ordered by likes; friend reviews; prediction results with the expected outcome | integration | `go test ./internal/httpapi -run TestMergeRules` | green |
| 1 | MATCH-06 | Fans' Player of the Match needs at least 3 ratings | unit | `go test ./internal/domain -run TestPlayerAveragesAndPotm` | green |
| 2 | HOME-01 to 06 | `/v1/home` has `onNow`, `yourCall`, `yourCallSplit`, `yourCallMine`, `feed`, `justFinished`, `board`, `unread`, with no null arrays | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| 3 | SOC-02 | Like toggles on a simulated friend's stub add to the simulated base count and toggle back | integration | `go test ./internal/httpapi -run TestLikeToggleOnSimLog` | green |
| 3 | SOC-02 | Like, reply and delete on my own stub | integration | `go test ./internal/httpapi -run TestLogCreateUpdateKeepsStub` | green |
| 3 | SOC-01 | Follow toggles on and off; following yourself is 400 | integration | `go test ./internal/httpapi -run TestFollowToggle` | green |
| 4 | SOC-03, SOC-04 | `/v1/notifications` and `/v1/search?q=` answer with `items`, `matches`, `users` | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| all | all | Screens typecheck against the contract | static | `npx tsc --noEmit` | green |

*Status: pending · green · red · flaky*

---

## Results (run 2026-10-22)

```
$ npx tsc --noEmit
(no output)
tsc exit 0

$ go test -count=1 ./...
ok  	github.com/swathivallabhaneni289/innings/internal/domain	1.134s
ok  	github.com/swathivallabhaneni289/innings/internal/httpapi	2.024s
(all five tested packages ok; full output in 02-VALIDATION.md)
```

A throwaway server on a scratch database (port 8097) showed, for a fresh guest: 6 matches On Now, a Your Call match, 40 feed items, 3 Just Finished, 6 Bragging Board rows, 24 notifications.

---

## Wave 0 Requirements

- [ ] Jest tests for `PredictionPicker` step logic (`firstMissing`, `samePicks`) and the unread count

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Picker flips to Locked at the first ball with the screen open | PRED-02 | Needs the clock to pass a start time | Open an upcoming match a few minutes before its start and wait |
| Like and "I was there too" feel instant and roll back offline | SOC-02 | Needs a dropped connection | Stop the API, tap like, confirm the count returns |
| Bell badge clears after opening notifications | SOC-03 | Device store | Open the bell, go back, badge gone |

Not recorded as a walkthrough. Carried into the Phase 6 screenshot QA.

---

## Validation Sign-Off

- [x] All tasks have an automated verify
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [ ] `nyquist_compliant: true`: not set; app logic has no unit tests and the manual checks are pending

**Approval:** automated gates passed 2026-10-22; manual checks carried to Phase 6
