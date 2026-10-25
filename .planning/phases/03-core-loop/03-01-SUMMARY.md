---
phase: 03-core-loop
plan: 01
subsystem: ui
tags: [expo-router, react-query, reanimated, react-native-svg, slider, image-picker, haptics]

requires:
  - phase: 01-foundation
    provides: components, data hooks, contract
  - phase: 02-api-simulated-world
    provides: matches, match page, log suggestions, search, logs and watching routes
provides:
  - Matches tab with combined filters and day grouping
  - Match Page for upcoming, live, completed and classic matches
  - The /log modal with four overs and the Printed step
  - TicketStub, MatchTicket, SpinningBall and Stamp components
affects: [phase-04 Home and predictions, phase-05 Profile and share card, phase-06 re-skin]

tech-stack:
  added: []
  patterns: ["one route per screen, feature code under src/features/<area>", "keep previous data dimmed while a filter loads", "draft state in the flow, converted to NewLog only on save"]

key-files:
  created: ["src/app/(tabs)/matches.tsx", "src/app/match/[id].tsx", "src/app/log/index.tsx", src/features/matches/group.ts, src/features/match/MatchPage.tsx, src/features/log/LogFlow.tsx, src/features/log/draft.ts, src/features/log/Printed.tsx, src/components/TicketStub.tsx, src/components/Motion.tsx]
  modified: []

requirements-completed: [MATCH-01, MATCH-02, MATCH-03, MATCH-04, MATCH-05, MATCH-08, MATCH-09, MATCH-10, LOG-01, LOG-02, LOG-03, LOG-04, LOG-05, LOG-06, LOG-08, DESIGN-02, DESIGN-04]

completed: 2026-10-22
status: complete
---

# Phase 3 Plan 1: Matches, Match Page, Log Flow and Ticket Stub Summary

**The core loop works end to end: filter the fixture list, open a Match Page that fits the match's state, and log it in four overs to a printed ticket stub with the LOGGED stamp.**

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented
- **Commits:** `8270558` (Matches and the Match Page), `b7bf2dd` (Log flow and stub) (committed together on 2026-10-22, after Phase 6)

## Accomplishments

- **Matches tab** (`src/app/(tabs)/matches.tsx`, `src/features/matches/`): Live, Upcoming and Recent chips plus India, IPL, WPL, World Cup, BBL, SA20 and The Hundred; tapping a selected chip clears it. `group.ts` groups by day with "Live now" first, then Today (still to come, then finished), Tomorrow, Yesterday and dated groups. `MatchCard` shows competition, format, stage, teams, venue and either the live score, the first-ball time or the result with the fan rating in balls and the count ("No fan ratings yet" otherwise). The empty copy names the filter ("Nothing live in the IPL at the moment").
- **Match Page** (`src/app/match/[id].tsx`, `src/features/match/`): one page for fixtures and classics, refetched every 20 seconds and on pull. BEFORE: header, "I'm watching" with the seven places, "Friends tuning in". DURING: live scoreboard with the batting side marked and the status line, a reactions row shown as coming soon, "Friends watching now". AFTER: result card, "Log this match" (or "Edit your log"), fan rating with the big average and ten half-ball bars, Popular and Friends review tabs, where fans watched as shares.
- **Log flow** (`src/app/log/index.tsx`, `src/features/log/`): a full-screen modal titled like a scorebook ("Over 2 of 4") with an over strip. Over 1 lists "Live now / just ended" and "Recent" from `/v1/log/suggestions` and searches older matches and classics. Over 2 is a half-ball rating with a spinning ball and an "Instant classic" box. Over 3 lists both XIs with a 1 to 10 slider, Skip and a crown for your Player of the Match. Over 4 has the seven places, a free-text "Where exactly?", a head count stepper, tags for friends you follow, a photo from the library and the review.
- **Printed** (`Printed.tsx`): the ticket feeds out, the stamp lands with a success haptic, then your Player of the Match, the review line, "Share story" (wired in Phase 5) and Done.
- **Ticket and motion** (`src/components/TicketStub.tsx`, `src/components/Motion.tsx`): SVG ticket outline with a notched perforation and a torn far edge; `MatchTicket` prints teams, date, venue, rating, Instant classic and the stub number. `Stamp` and `SpinningBall` use Reanimated.

## Decisions Made

- Opening the flow from a Match Page or the Vault passes `matchId` and starts at Rate it, so the over count drops by one; the flow is keyed by match id so switching matches resets it.
- A classic has no XI, so Rate the players is dropped: a classic picked in Over 1 runs three overs, and one opened from the Vault runs two.

## Deviations from Plan

- **Live reactions** are a disabled row labelled as coming soon, not working buttons. The plan marks them optional for later; they are v2 (LIVE-01).
- **Delete** is not offered in the app. `DELETE /v1/logs/{id}` and `useActions().deleteLog` exist (Phase 2), but no screen calls them.

## Issues Encountered

None recorded.

## Next Phase Readiness

`MatchPage.tsx` leaves the prediction panel, Fans' Player of the Match and prediction results for Phase 4. `Printed` takes the share hook from Phase 5.

## Self-Check: PASSED

`npx tsc --noEmit` exits 0; the related API tests pass (see `03-VALIDATION.md`). Screens have not had a recorded walkthrough; that is folded into the Phase 6 screenshot QA.
