---
phase: 5
slug: vault-profile
status: automated_gates_passed
nyquist_compliant: false
wave_0_complete: true
created: 2026-10-22
---

# Phase 5: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | App: TypeScript compiler (strict); no Jest tests yet. Vault data and routes: Go `testing` with `httptest` |
| **Config file** | `tsconfig.json` |
| **Quick run command** | `npx tsc --noEmit` |
| **Full suite command** | `npx tsc --noEmit && go test ./...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task:** `npx tsc --noEmit` (screens) or `go test ./internal/sim -run TestClassics` (seed)
- **After the phase:** `go test ./...`
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Plan | Requirement | Behavior | Test Type | Automated Command | Status |
|------|-------------|----------|-----------|-------------------|--------|
| 05-01 | VAULT-06 | Every classic is well formed, starts at 12:00 UTC on its date, is ordered oldest first and is findable by id | unit | `go test ./internal/sim -run TestClassics` | green |
| 05-01 | VAULT-06 | One source row per match | data | `python3 -c "import json;print(len(json.load(open('internal/sim/classics.json'))))"` gives 58; 58 data rows in `docs/VAULT_SOURCES.md` | green |
| 05-01 | VAULT-06 | Facts checked by three independent agents | manual (agent) | Tally: 58 confirmed, 0 corrected, 0 dropped | green |
| 05-02 | VAULT-01, VAULT-04 | `/v1/vault` has 5 collections, `onThisDay`, `top5`, `eras` with no nulls | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| 05-02 | VAULT-02 | `/v1/vault/rankings` with and without `format=T20&era=2020s` | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| 05-02 | VAULT-05 | Six pins is 400; an unknown match is 400; an empty Top 5 is allowed; a pin shows in `/v1/vault` | integration | `go test ./internal/httpapi -run TestTop5AndMe` | green |
| 05-02 | PROF-07 | `PUT /v1/me` saves name and bio; a blank name is 400 | integration | `go test ./internal/httpapi -run TestTop5AndMe` | green |
| 05-02 | PROF-03, PROF-06 | `/v1/users/me` and a friend's page carry `stats` with every field, `logs`, `top5` | integration | `go test ./internal/httpapi -run TestEveryGET` | green |
| 05-02 | PROF-06 | Badge rules (first call, upset, 5-match streak, sharpshooter, stadium regular) | unit | `go test ./internal/domain -run TestBadges` | green |
| 05-02 | all | Screens, settings and share card typecheck | static | `npx tsc --noEmit` | green |

*Status: pending · green · red · flaky*

---

## Results (run 2026-10-22)

```
$ npx tsc --noEmit
(no output)
tsc exit 0

$ go test -count=1 ./...
ok  	github.com/swathivallabhaneni289/innings/internal/httpapi	2.024s
ok  	github.com/swathivallabhaneni289/innings/internal/sim	2.896s
(all five tested packages ok; full output in 02-VALIDATION.md)
```

On a throwaway server (scratch database, port 8097) a fresh guest's Vault showed collections of 23, 17, 12, 6 and 11 matches, eras 1970s to 2020s, rankings of 77 matches (58 classics plus 19 completed fixtures) and an On This Day four years back. The guest's profile showed 5 stubs, 3.5 average given, 67% accuracy, 115 points and the First Call, Called the Upset and Sharpshooter badges.

---

## Wave 0 Requirements

- [ ] Jest tests for `src/features/vault/top5.ts` (order kept, sixth pin refused) and `anniversary` / `matchYear`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Share story opens the system sheet with a 9:16 PNG | LOG-07 | Native share sheet | Save a log on a device, tap Share story |
| Pick a profile photo from the library | PROF-07 | OS permission sheet | Settings, Choose photo |
| Theme switch applies at once | DESIGN-01 | Visual | Settings, Appearance, each option |
| Log a classic from On This Day with a memory | VAULT-03 | Flow | Vault, "I watched this — log it", add a place note |

Not recorded as a walkthrough. Carried into the Phase 6 screenshot QA.

---

## Validation Sign-Off

- [x] All tasks have an automated verify
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [ ] `nyquist_compliant: true`: not set; app logic has no unit tests and the manual checks are pending

**Approval:** automated gates passed 2026-10-22; manual checks carried to Phase 6
