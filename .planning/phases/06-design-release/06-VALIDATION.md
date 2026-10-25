---
phase: 6
slug: design-release
status: complete
nyquist_compliant: false
wave_0_complete: true
created: 2026-10-22
updated: 2026-10-25
---

# Phase 6: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | App: TypeScript compiler (strict) and Jest with `jest-expo` (`src/lib/__tests__/format.test.ts`). Server: Go `testing` with `httptest`. Visual: screenshots through the QA driver (`src/lib/qaDriver.ts`) |
| **Config file** | `tsconfig.json`, `package.json` (`"jest": { "preset": "jest-expo" }`) |
| **Quick run command** | `npx tsc --noEmit` |
| **Full suite command** | `go vet ./... && go test ./... && npx tsc --noEmit && npx jest` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task:** `npx tsc --noEmit`
- **After every visual fix:** re-capture the screen through the QA driver and compare with the review
- **After the phase:** the full suite, then a fresh clone from GitHub
- **Max feedback latency:** 30 seconds (screenshots: about a minute per screen)

---

## Per-Task Verification Map

| Plan | Requirement | Behavior | Test Type | Automated Command | Status |
|------|-------------|----------|-----------|-------------------|--------|
| 06-01 | DESIGN-06 | Tokens, fonts and every screen typecheck after the re-skin | static | `npx tsc --noEmit` | green |
| 06-01 | DESIGN-06 | No paper grain is left in the app | static | `grep -r "grain" src` finds nothing | green |
| 06-01 | DESIGN-06 | Formatting rules on the tickets: innings lines, all-out as runs only, countdown, ticket date, stub number | unit | `npx jest` (14 tests) | green |
| 06-01 | DESIGN-07 | 22 screens captured and scored in a 6-pillar review; all 12 prioritized fixes applied and re-captured | manual (screenshots) | `06-UI-REVIEW.md`; fixes in `ff22e3f` | green |
| 06-01 | PLAT-06 | Home loads with data from the API on an Android emulator | manual (emulator) | `npx expo run:android --port 8082`, `adb reverse tcp:8088 tcp:8088` | green |
| 06-01 | PLAT-06 | Guest bootstrap and avatar colour | integration | `go test ./internal/auth ./internal/httpapi` | green |
| 06-01 | PLAT-07 | `app.json` points at the icon, adaptive icon layers and splash on #ECEAE4 | static | `grep -n "ECEAE4" app.json` | green |
| 06-01 | PLAT-08 | The README's steps work from a fresh clone | end to end | clone, `npm ci`, `go test ./...`, `npx tsc --noEmit`, `npx jest`, 32-check API run | green |

*Status: pending · green · red · flaky*

---

## Results

**Fresh clone of `28869cd` (2026-10-22).** 200 files. `gofmt` and `go vet` clean; `go test -count=1 ./...` ok in auth, domain, httpapi, sim and store; `npm ci`; `tsc` clean; Jest 14/14. An end-to-end run against an empty database on port 8096 passed 32 of 32 checks, including:

- a prediction after the first ball is refused with 409 `locked`
- logging a match again keeps its stub number
- likes and follows toggle
- a sixth Top 5 pin is refused
- deleting a log returns 204

**Re-run (2026-10-25), working copy at `28869cd`:**

```
$ go vet ./... && go test -count=1 ./...
ok  	github.com/swathivallabhaneni289/innings/internal/auth	0.765s
ok  	github.com/swathivallabhaneni289/innings/internal/domain	1.238s
ok  	github.com/swathivallabhaneni289/innings/internal/httpapi	2.236s
ok  	github.com/swathivallabhaneni289/innings/internal/sim	3.086s
ok  	github.com/swathivallabhaneni289/innings/internal/store	2.523s

$ npx tsc --noEmit
tsc exit 0

$ npx jest
Tests:       14 passed, 14 total
```

---

## Wave 0 Requirements

- [x] Jest configured and running (`jest-expo`, 14 tests)
- [ ] Jest tests for `src/features/vault/top5.ts` and the Log draft (carried from Phase 5)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Every screen in floodlights | DESIGN-07 | Visual | Settings, Appearance, Floodlights, then walk every tab. Only Home, Match Page after play, Profile and Vault were captured in floodlights for the review |
| Stock colours for WPL, World Cup, BBL, SA20, The Hundred and Vault classics | DESIGN-06 | Developer's call | Look at a ticket of each on the Matches tab and in the Vault |
| On a real phone | PLAT-01 | Needs the developer's device | `npx expo run:android --device` or `npx expo run:ios --device`, phone on the same Wi-Fi as the API |
| Share story opens the system sheet | LOG-07 | Native share sheet | Log a match, tap Share story |

---

## Validation Sign-Off

- [x] All tasks have an automated or recorded manual verify
- [x] No watch-mode flags
- [x] Feedback latency < 30s for the automated gates
- [ ] `nyquist_compliant: true`: not set; screens have no component tests and the manual checks above wait on the developer's phone testing

**Approval:** automated gates passed 2026-10-22 and again 2026-10-25; manual checks above are open for the developer's testing
