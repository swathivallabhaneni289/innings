# Phase 1: Foundation - Context

**Gathered:** 2026-10-22
**Status:** Complete

<domain>
## Phase Boundary

Everything the parallel workstreams in Phases 2 to 5 need to agree on before any of them starts: the Expo app scaffold, the JSON contract between the app and the Go API (TypeScript first, mirrored in Go), the first design system (tokens, fonts, core components, two themes) and the data layer screens read through. No screen content beyond the root layout and the tab bar.

</domain>

<decisions>
## Implementation Decisions

### Stack
- **D-01:** The phone app is Expo SDK 57 with React Native and TypeScript (strict), so one codebase runs on iPhone and Android. Routes live in `src/app/` (Expo Router); the `@/*` alias points at `src/`.
- **D-02:** The backend is Go, not Supabase (developer, 2026-10-22). Phase 1 only fixes the contract; Phase 2 builds the API.
- **D-03:** Server data goes through TanStack Query hooks in `src/data/index.ts`; on-device preferences (guest token, theme, notifications-seen time) go through a persisted Zustand store in `src/data/store.ts`. Screens import from `@/data` only.

### Contract
- **D-04:** `src/domain/types.ts` (entities) and `src/domain/api.ts` (one response type per endpoint) are the single source of truth; `docs/API.md` lists the routes; Go mirrors both in `internal/domain/types.go` and `internal/domain/api.go`.
- **D-05:** JSON rules both sides rely on: camelCase field names; a TS `?` field is `omitempty` in Go, a TS `X | null` field is always present; arrays and maps are never null; `teamIds` is a 2-element array; timestamps are RFC 3339 UTC; errors are `{"error", "message"}`.
- **D-06:** Ratings are 0.5 to 5 in half-ball steps everywhere (`RATING_STEPS` in `src/domain/constants.ts`); player ratings are 1 to 10.
- **D-07:** The "India" filter is a team filter (`ind`, `ind-w`), not a competition, so the competition list stays honest.

### Design system (first pass)
- **D-08:** Heritage scorebook from the plan's design section: paper #F5F0E6, ink navy #1C2333, cricket-ball red #B3261E, pitch green #2F6B3A for wins only; a dark "floodlights" palette; Oswald for scores and headings, Source Serif 4 for reviews, IBM Plex Mono for stats; faint grain and ruled lines. (Superseded on 2026-10-22 by direction C; see Phase 6.)
- **D-09:** Theme preference is Auto, Paper or Floodlights; Auto follows the system scheme.
- **D-10:** Paper and tickets have sharp corners; only balls and avatars are round (`radius` in `src/theme/tokens.ts`).

### Claude's Discretion
- The cricket ball drawing (`src/components/CricketBall.tsx`) and the shared component set in `src/components/ui.tsx`.
- Loading and error copy in cricket terms ("Walking out to bat…").

</decisions>

<canonical_refs>
## Canonical References

- `docs/PLAN.md`: the product plan and the Build decisions table
- `docs/API.md`: routes and JSON rules
- `AGENTS.md`: Expo rules for this repo (use `npx expo install`, never hand-edit `ios/` or `android/`)

</canonical_refs>

<code_context>
## Existing Code Insights

The repo started from a blank Expo template (`d9d233a Initial commit`, with `App.tsx` and `index.ts` at the root). Both were removed in favour of `expo-router/entry` and `src/app/`.

</code_context>

<deferred>
## Deferred Ideas

- Email sign-in: guest accounts only for v1 (decided in Phase 2).

</deferred>

---

*Phase: 1-Foundation*
*Context gathered: 2026-10-22*
