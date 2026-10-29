---
gsd_state_version: 1.0
milestone: v1.2
milestone_name: Real matches
current_phase: 09
current_phase_name: during-the-match
status: in-progress
stopped_at: match chat done; building predictions during a match
last_updated: "2026-10-29T15:20:00.000Z"
last_activity: 2026-10-29
last_activity_desc: Match chat (CHAT-01) on real matches; next predictions during a match
progress:
  total_phases: 9
  completed_phases: 8
  total_plans: 9
  completed_plans: 9
  percent: 100
---

# Project State

> **CURRENT STATUS (2026-10-29).** v1.0, v1.1 and v1.2 are complete: 8 phases, 73 requirements. Matches are now real, from ESPN's public scores feed (no key): live scores, the week's fixtures with announced squads, playing XIs at the toss, and results with top scorers and player of the match. Next: the match chat and predictions during a match. Before that, Fans sign in with a phone number and a 6-digit code, set up a fan card with a unique @username, and find friends from their contacts; Settings has sign out and delete account. The app was walked end to end on a new Android development build (Pixel 8 emulator, Android 16) and, on 2026-10-28, on an iPhone simulator, including two phones finding each other in their contacts. That walk found and fixed two keyboard bugs on iPhone. `go test ./...`, `go vet`, `npx tsc --noEmit` and 39 Jest tests pass, and a 70-check end-to-end run against a fresh server passed. Codes show on screen in dev mode; real texts need an SMS provider account. The code is on the private GitHub repo `swathivallabhaneni289/innings`, branch `main`. Next: the developer tests on their own phones.

## Project Reference

See: .planning/PROJECT.md (updated 2026-10-25)

**Core value:** Capture how fans felt about a match: the prediction, the rating, the memory of where they watched.
**Current focus:** The match chat and predictions during a match, on real matches
**Requirements document:** `docs/REQUIREMENTS.md` (the product requirements with acceptance checks); the build checklist is `.planning/REQUIREMENTS.md`

## Current Position

Phase: 07 of 07 (phone sign-in & onboarding): complete
Plan: 07-01 complete (`phases/07-accounts/07-01-SUMMARY.md`)
Status: v1.1 milestone complete
Last activity: 2026-10-28: sign-in walked on iPhone and across two phones; keyboard fixes on the fan card, the log and Settings

Progress: [██████████] 100%

## Performance Metrics

**Velocity:**

- Total plans completed: 9
- Average duration: not instrumented
- Total execution time: one evening for v1.0 (2026-10-22), one afternoon for v1.1 (2026-10-25), parallel sub-agents

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 | 1 | - | - |
| 02 | 2 | - | - |
| 03 | 1 | - | - |
| 04 | 1 | - | - |
| 05 | 2 | - | - |
| 06 | 1 | - | - |
| 07 | 1 | - | - |

*Timing was not recorded per plan.*

## Accumulated Context

### Decisions

Decisions are logged in the PROJECT.md Key Decisions table.
Recent decisions:

- [Whole project]: Go backend, Expo SDK 57 app, Go API plus SQLite in place of Supabase, API on 8088 and Metro on 8082 (developer, 2026-10-22)
- [Phase 02]: No cricket API key; `internal/sim` generates the world around server start. `INNINGS_ANCHOR` pins the clock for repeatable runs
- [Phase 02]: Build decisions settled the plan's gaps: streak x1 / x1.5 / x2, upset at lock time, tied top scorers all count, half-balls everywhere, Fans' Player of the Match needs 3 ratings, Bragging Board is the last 7 days
- [Phase 05]: The Vault ships 58 real matches; 3 independent checker agents confirmed all 58 (0 fixed, 0 dropped); sources in `docs/VAULT_SOURCES.md`
- [Phase 06]: Direction C, Matchday Ticket, chosen by the developer on 2026-10-22 and built across every screen. Red is kept for live, stamps, serials, the Log ball and small text actions (See all, PREDICT, ghost buttons); green for wins only
- [Phase 06]: Phones run a development build; Expo Go from the stores only opens the newest SDK
- [Phase 06]: The GitHub repo is private
- [Phase 07]: Phone and text-code sign-in like Beli, replacing guest accounts (developer, 2026-10-25). Twilio Verify when configured, dev mode (`make api`) otherwise, refused with neither. New accounts start empty
- [Phase 07]: Android draws edge to edge, so text screens use `KeyboardAware` (keyboard-height padding on Android) instead of KeyboardAvoidingView alone
- [2026-10-28]: The developer found the scrolling Welcome page too plain and too wordy, so it's one screen with nothing to scroll: a ticket plays the five parts of the app in turn (call, rate, stub, Bragging Board, Vault) with one line each and story-style progress bars. It waits for the launch splash, pauses for screen readers, follows Reduce Motion, and scrolls only when the system text is too large to fit (`src/features/auth/WelcomeStory.tsx`)
- [2026-10-29]: Polishing that story wasn't enough ("you kept the same design"), and real app screens in a phone frame weren't either, so the Welcome screen was redesigned: the whole screen is one match ticket on card stock, with giant team codes rolling like a scoreboard, your rating filling in with the app's word for it, LOGGED beside the codes, the venue, a fun fact on the stub and the buttons below the perforation (`src/features/auth/WelcomeTicket.tsx`, replacing `WelcomeStory.tsx`)
- [2026-10-29, earlier]: The developer wanted the Welcome story catchier but not busy, so it was cut back rather than redesigned: a short headline that wipes up and one line under it (no tag), one moment per ticket (LOCKED and LOGGED stamps press the card, the balls fill, you climb past Aditi on the Bragging Board), and the Vault's stack sits straight with no tilt
- [Phase 07, 2026-10-28]: `KeyboardAware` wraps the whole screen, because on iOS KeyboardAvoidingView measures itself against its parent. A screen that is one ScrollView of fields uses `automaticallyAdjustKeyboardInsets` with `KeyboardAware scrollInsets`, so iOS scrolls the focused field into view

### Pending Todos

- Developer: test on a real iPhone and Android phone. Each needs a new development build for sign-in (`npx expo run:ios --device`, `npx expo run:android --device`), with the phone on the same Wi-Fi as `make api`
- Developer: pick an SMS provider before friends sign up from their own phones (Twilio Verify is wired in; India needs DLT registration, unconfirmed for Twilio)
- Developer: confirm the stock colours for WPL, World Cup, BBL, SA20, The Hundred and Vault classics, and the darkened floodlights stocks (shown in the screenshots; no sign-off recorded)
- Fix the 13 lint errors and 1 warning that `npx expo lint` reports in code from earlier phases (React's rules on refs and setState in render or effects)
- Walk every screen in floodlights; the Phase 6 review captured four of them in the dark theme
- Jest tests for `src/features/vault/top5.ts` and the Log draft

### Blockers/Concerns

- Real matches come from ESPN's unofficial scores feed, which can change without notice; the feed's tests run on saved responses and need refreshing when it does
- Real text messages need an SMS provider account; the Twilio client is tested only against a mock
- Store builds (EAS, App Store, Play Store) need the developer's Apple and Google accounts
- Log and profile photos are device-local file URIs stored as strings; they don't upload, so other devices can't see them
- `DELETE /v1/logs/{id}` and `useActions().deleteLog` exist, but no screen offers delete yet
- Sessions don't expire, and CORS allows any origin

## Deferred Items

Items acknowledged and carried forward:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Data | Real cricket API (DATA-01) | v2 | 2026-10-22 |
| Sign-in | Real text messages: SMS provider, sender ID, India DLT (SMS-01) | v2, needs the developer's account | 2026-10-25 |
| Sign-in | Sign in with Apple and Google (ACCT-09) | v2 | 2026-10-25 |
| Media | Server photo storage (MEDIA-01) | v2 | 2026-10-22 |
| Release | EAS and store submission (STORE-01) | v2, needs developer accounts | 2026-10-22 |
| Features | Live reactions, Season Wrapped, fan clubs, AI review summaries | v2 (plan's "Later") | 2026-10-22 |

## Session Continuity

Last session: 2026-10-29
Stopped at: full-screen ticket Welcome pushed; the developer is trying it on their iPhone simulator
Resume file: none

- Run locally: `make api` (port 8088, SQLite file `./innings.db`, sign-in codes in dev mode; `make reset-db` wipes it) and `make app` (Metro on 8082; press `i` for iOS or `a` for Android). To sign in, type any number and tap the dev code shown on the screen.
- After pulling Phase 7, build the app again (`npx expo run:android --port 8082`, `npx expo run:ios`): sign-in added native modules.
- Android emulator: AVD `innings_pixel`; `adb reverse tcp:8082 tcp:8082` and `adb reverse tcp:8088 tcp:8088`; builds need `JAVA_HOME` set to JDK 17. The emulator also reaches the API at 10.0.2.2.
- On a phone, the app reaches the API through the Metro host's LAN address; `EXPO_PUBLIC_API_URL` overrides it.
- Screenshot tooling: start Metro with `EXPO_PUBLIC_QA_DRIVER=<url>` and the app polls that URL for a route or `theme:paper` / `theme:floodlights` (`src/lib/qaDriver.ts`).
