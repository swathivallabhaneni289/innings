---
phase: 01-foundation
plan: 01
subsystem: app-shell
tags: [expo, expo-router, typescript, json-contract, design-tokens, react-query, zustand]

# Dependency graph
requires: []
provides:
  - Expo SDK 57 app with Expo Router routes under src/app and the @/* alias on src/
  - The JSON contract (src/domain/types.ts, src/domain/api.ts) and its Go mirror (internal/domain)
  - docs/API.md with every route and the JSON rules
  - Paper and floodlights palettes, type scale and the core component set
  - The data layer (src/data) every screen reads through
affects: [phase-02 API and sim, phase-03 to 05 screens, phase-06 re-skin]

tech-stack:
  added: [expo ~57.0.31, react-native 0.86.3, react 19.2.3, typescript ~6.0.3, expo-router, "@tanstack/react-query 5", "zustand 5", "@react-native-async-storage/async-storage", react-native-reanimated 4.5, react-native-svg, "@expo-google-fonts/oswald", "@expo-google-fonts/source-serif-4", "@expo-google-fonts/ibm-plex-mono"]
  patterns: ["contract-first: TS types are the source of truth, Go mirrors them", "screens import from @/data only", "QueryState renders loading and error so 'no data yet' is never shown as empty"]

key-files:
  created: [docs/API.md, src/domain/types.ts, src/domain/api.ts, src/domain/constants.ts, internal/domain/types.go, internal/domain/api.go, src/theme/tokens.ts, src/theme/index.tsx, src/components/index.ts, src/data/client.ts, src/data/store.ts, src/data/index.ts, src/app/_layout.tsx, "src/app/(tabs)/_layout.tsx"]
  modified: [package.json, app.json, tsconfig.json]

requirements-completed: [PLAT-01, PLAT-05, DESIGN-01, DESIGN-03, DESIGN-05]

completed: 2026-10-22
status: complete
---

# Phase 1 Plan 1: Scaffold, Contract, Design System, Data Layer Summary

**An Expo Router shell with five tabs, one JSON contract in TypeScript and Go, a two-theme heritage scorebook design system, and a react-query data layer, so the later workstreams could run in parallel without touching shared files.**

> These shared files are being reworked in Phase 6 for direction C. This summary records the state Phase 1 left them in.

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented (one evening for Phases 1 to 5)
- **Commits:** none yet; the work is uncommitted on top of `d9d233a Initial commit`

## Accomplishments

- Replaced the blank template (`App.tsx`, `index.ts`) with `expo-router/entry`; routes live in `src/app/`, the tab group in `src/app/(tabs)/`.
- Wrote the contract once: entities in `src/domain/types.ts`, one response type per route in `src/domain/api.ts`, shared lists (watch places, competitions, collections, rating steps, `MAX_TOP5`) in `src/domain/constants.ts`. Go mirrors it in `internal/domain/types.go` and `internal/domain/api.go`.
- `docs/API.md` lists 12 read routes (bootstrap among them), 10 write routes and `GET /healthz`, plus the JSON rules (camelCase, never-null arrays and maps, RFC 3339 times, `{"error", "message"}` errors).
- `src/theme/tokens.ts`: paper (#F5F0E6 ground, #1C2333 ink, #B3261E red, #2F6B3A green) and floodlights palettes, spacing, sharp-corner radii, three font families and a 12-role type scale. `ThemeProvider` resolves Auto, Paper or Floodlights.
- Components: `Paper` (grain texture from `assets/textures/`, since removed in Phase 6, plus SVG ruled lines), `Screen`, `Txt`, `States` (`Loading`, `ErrorState`, `Empty`, `QueryState`), `ui` (Button, Chip, Panel, Rule, SectionHeader, SplitBar, Stat, TeamBadge, Avatar, LiveTag), `CricketBall`, `BallRating` (half-balls).
- Root layout loads ten font faces, waits for stored preferences, bootstraps the guest and only then renders the stack; the Log screen is a full-screen modal.
- Tab bar: Home, Matches, a raised red cricket ball for Log, Vault, Profile; the Log tab opens `/log` instead of switching tabs.
- Data layer: `api()` adds the Bearer token and maps failures to `ApiError`; when the server is unreachable it says so and suggests `make api`. The API host comes from `EXPO_PUBLIC_API_URL`, else the Metro host's LAN address (so a phone in Expo Go reaches the laptop), else 127.0.0.1, port 8088.

## Files Created/Modified

- `src/domain/*`, `internal/domain/types.go`, `internal/domain/api.go`, `docs/API.md`: the contract
- `src/theme/*`, `src/components/*` (except `TicketStub.tsx` and `Motion.tsx`, which are Phase 3): the design system
- `src/data/client.ts`, `src/data/store.ts`, `src/data/index.ts`: the data layer
- `src/app/_layout.tsx`, `src/app/(tabs)/_layout.tsx`: shell and tab bar
- `package.json`, `app.json`, `tsconfig.json`: scaffold

## Decisions Made

- Teams and players from bootstrap are kept in an in-memory catalog so `getTeam`, `getPlayer` and `matchTitle` stay synchronous in render code.
- `useActions()` invalidates every query except `me` after a write, trading a few extra requests for never showing stale counts.

## Deviations from Plan

None.

## Next Phase Readiness

- Phase 2 can code the API and the sim against `internal/domain` in parallel.
- Phases 3 to 5 can build screens from `@/components`, `@/data` and `@/theme`.

## Self-Check: PASSED

`npx tsc --noEmit` exits 0 (see `01-VALIDATION.md`).
