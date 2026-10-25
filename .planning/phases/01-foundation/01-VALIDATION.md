---
phase: 1
slug: foundation
status: automated_gates_passed
nyquist_compliant: false
wave_0_complete: true
created: 2026-10-22
---

# Phase 1: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | TypeScript compiler (`tsc --noEmit`, strict). Jest with `jest-expo` is configured in `package.json` (`testMatch: **/__tests__/**/*.test.ts?(x)`), but there are no test files yet. |
| **Config file** | `tsconfig.json` (extends `expo/tsconfig.base`, `strict: true`) |
| **Quick run command** | `npx tsc --noEmit` |
| **Full suite command** | `npx tsc --noEmit && go test ./...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task:** `npx tsc --noEmit`
- **After the phase:** the contract is exercised end to end by the Phase 2 API tests (`TestEveryGET`, `TestNoNulls`), which fail on any null array or map
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Behavior | Test Type | Automated Command | Status |
|---------|------|------|-------------|----------|-----------|-------------------|--------|
| Task 1 | 01-01 | 1 | PLAT-01 | App scaffold typechecks under strict mode | static | `npx tsc --noEmit` | green |
| Task 2 | 01-01 | 1 | PLAT-05 | TS contract compiles; every screen's data shape comes from `src/domain/api.ts` | static | `npx tsc --noEmit` | green |
| Task 2 | 01-01 | 1 | PLAT-05 | Go mirror compiles | static | `go build ./internal/domain` | green |
| Task 2 | 01-01 | 1 | PLAT-05 | No array or map key in any response is null (keys listed in `arrayKeys`) | integration | `go test ./internal/httpapi -run TestEveryGET -v` | green (Phase 2) |
| Task 3 | 01-01 | 1 | DESIGN-01, DESIGN-03, DESIGN-05 | Theme, type scale and components typecheck; `BallRating` accepts half values | static | `npx tsc --noEmit` | green |
| Task 4 | 01-01 | 1 | PLAT-05 | Hooks and `useActions` typecheck against the contract | static | `npx tsc --noEmit` | green |

*Status: pending · green · red · flaky*

---

## Results (run 2026-10-22)

```
$ npx tsc --noEmit
(no output)
tsc exit 0

$ go build ./internal/domain
(no output, exit 0)
```

---

## Wave 0 Requirements

- [x] TypeScript strict mode on
- [ ] A first Jest test under `src/**/__tests__/` (for example `src/features/matches/group.ts` or `src/lib/format.ts`) so app logic has unit coverage, not only types

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Fonts load and the splash hides only after fonts and stored prefs are ready | DESIGN-05 | Needs a running app | Cold-start the dev build; no flash of system font |
| Auto follows the system scheme; Paper and Floodlights override it | DESIGN-01 | Needs the OS appearance switch | Toggle system dark mode with Auto, then force each theme in Settings |
| Phone reaches the API on the laptop | PLAT-01 | Needs a device on the same network | Start `make api`, install a development build with `npx expo run:android --device` or `npx expo run:ios --device`, confirm Home loads |

These were not recorded as a formal walkthrough. They are folded into the Phase 6 screenshot QA, which covers every screen in both themes.

---

## Validation Sign-Off

- [x] All tasks have an automated verify
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [ ] `nyquist_compliant: true`: not set, because app logic has no unit tests yet (types only)

**Approval:** automated gates passed 2026-10-22; manual checks carried to Phase 6
