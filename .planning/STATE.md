---
gsd_state_version: 1.0
milestone: v1.3
milestone_name: Email sign-in and close-out
current_phase: 10
current_phase_name: email-sign-in-and-close-out
status: complete
stopped_at: v1.3 complete, then past matches and the match page's rating, Top 5 pin, review box and Top 5 editor (2026-11-13), then the repo made public with a PRD and the build's prompts (2026-11-16); sign-up codes stay on screen in dev mode, since the developer won't use a personal email password, and `make demo` added (2026-11-17)
last_updated: "2026-11-17T08:47:37.000Z"
last_activity: 2026-11-17
last_activity_desc: Sign-up codes stay on screen rather than use a personal email password; make demo runs a ready-made demo account on today's real matches (or the simulation)
progress:
  total_phases: 10
  completed_phases: 10
  total_plans: 9
  completed_plans: 9
  percent: 100
---

# Project State

> **CURRENT STATUS (2026-11-16).** All ten phases are complete (v1.0 to v1.3, 79 requirements), and two additions came after them. On 2026-11-13 a quick task added past matches (DATA-02): every World Cup, Champions Trophy, Asia Cup and IPL season since 1973, 2,948 matches in `internal/sim/archive.json`, which `scripts/archive.py` builds from Wikipedia's tournament pages and checks match by match. Fans find them in search, open their pages and log them; they never show on the Matches tab or Home. Later that day, at the developer's request, every finished match's page gained the fan's own say: a tap on the balls rates it and saves at once, "Pin to Top 5" works for any match, a box under Reviews takes their review, and Edit on the Top 5 opens `/top5` to reorder it or take games out (MATCH-10 and VAULT-05, reworded). `go test ./...`, `go vet`, `npx tsc --noEmit`, `npx expo lint` and 69 Jest tests pass, and `scripts/smoke.py` passed 88 checks against a fresh server on 2026-11-16. The new match-page panels were checked on the "Innings QA" iPhone simulator only. Phase 10 moved sign-in from phone numbers to email and password. Sign-up asks only for an email and a password; a 6-digit code emailed to the address confirms it, and then come the fan card and find-friends steps. Forgot password works with an emailed code, Settings has change password, and both sign out the fan's other devices; delete account asks for the password. Phone numbers are no longer collected, and contacts match on SHA-256 hashes of lower-cased email addresses. Fans can back up to five teams, main team first. Matches are real, from ESPN's public scores feed, with a chat room and calls during each match (Phases 8 and 9). Phase 10 was walked on the "Innings QA" iPhone simulator and the Android emulator. In dev mode the code shows on screen; real emails go out once `.env` holds a mail account's settings (README → Real emails), which hasn't been tried, since the developer won't use a personal email password. The repo is the public GitHub repo `swathivallabhaneni289/innings` (private until 2026-11-16), branch `main`. Next: an Android pass on the new match-page panels, real email delivery, then the developer's own phones.

## Project Reference

See: .planning/PROJECT.md (updated 2026-10-30)

**Core value:** Capture how fans felt about a match: the prediction, the rating, the memory of where they watched.
**Current focus:** An Android pass on the match page's new panels, then real email delivery and the developer's own phones
**Requirements document:** `docs/REQUIREMENTS.md`, every requirement with its acceptance check. It has been the only requirements list since 2026-11-16, when the GSD checklist copy was merged into it

## Current Position

Phase: 10 of 10 (email sign-in and close-out): complete
Plan: none; built directly (see ROADMAP.md → Phase 10)
Status: v1.3 milestone complete
Last activity: 2026-11-16: the repo went public after a scan for secrets, the README was rewritten, and a PRD and the build's prompts were added. Before that, on 2026-11-13: past matches back to 1973 (DATA-02, a quick task), then the match page's rating, "Pin to Top 5" and review box, and the Top 5 editor

Progress: [██████████] 100%

## Performance Metrics

**Velocity:**

- Total plans completed: 9 (Phases 8, 9 and 10 were built without separate plan files)
- Average duration: not instrumented
- Total execution time: one evening for v1.0 (2026-10-22), one afternoon for v1.1 (2026-10-25), then v1.2 on 2026-10-29 and v1.3 on 2026-10-30, with parallel sub-agents

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
| 08 | - | - | - |
| 09 | - | - | - |
| 10 | - | - | - |

*Timing was not recorded per plan.*

## Accumulated Context

### Decisions

Decisions are logged in the PROJECT.md Key Decisions table and in `docs/PLAN.md` → Build decisions.
Recent decisions:

- [Whole project]: Go backend, Expo SDK 57 app, Go API plus SQLite in place of Supabase, API on 8088 and Metro on 8082 (developer, 2026-10-22)
- [Phase 02]: No cricket API key; `internal/sim` generates the world around server start. `INNINGS_ANCHOR` pins the clock for repeatable runs
- [Phase 02]: Build decisions settled the plan's gaps: streak x1 / x1.5 / x2, upset at lock time, tied top scorers all count, half-balls everywhere, Fans' Player of the Match needs 3 ratings, Bragging Board is the last 7 days
- [Phase 05]: The Vault ships 58 real matches; 3 independent checker agents confirmed all 58 (0 fixed, 0 dropped); sources in `docs/VAULT_SOURCES.md`
- [Phase 06]: Direction C, Matchday Ticket, chosen by the developer on 2026-10-22 and built across every screen. Red is kept for live, stamps, serials, the Log ball and small text actions (See all, PREDICT, ghost buttons); green for wins only
- [Phase 06]: Phones run a development build; Expo Go from the stores only opens the newest SDK
- [Phase 06]: The GitHub repo is private. Made public on 2026-11-16, after every commit was scanned for keys, passwords and personal data
- [Phase 07]: Phone and text-code sign-in like Beli, replacing guest accounts (developer, 2026-10-25). Twilio Verify when configured, dev mode (`make api`) otherwise, refused with neither. New accounts start empty. Replaced in Phase 10
- [Phase 07]: Android draws edge to edge, so text screens use `KeyboardAware` (keyboard-height padding on Android) instead of KeyboardAvoidingView alone
- [Phase 07, 2026-10-28]: `KeyboardAware` wraps the whole screen, because on iOS KeyboardAvoidingView measures itself against its parent. A screen that is one ScrollView of fields uses `automaticallyAdjustKeyboardInsets` with `KeyboardAware scrollInsets`, so iOS scrolls the focused field into view
- [2026-10-29]: The Welcome screen is one match ticket on card stock, with giant team codes rolling like a scoreboard, your rating filling in with the app's word for it, LOGGED beside the codes, the venue, a fun fact on the stub and the buttons below the perforation (`src/features/auth/WelcomeTicket.tsx`). It replaced a story of five tickets, which the developer found too plain even after polishing, and real app screens in a phone frame, which weren't enough either
- [Phase 08]: Matches are real, from ESPN's public scores feed: no key, nothing paid. One poller on the server saves what it learns in `feed_rows`, and request handlers never call ESPN. `INNINGS_MATCHES=sim` runs the simulation instead
- [Phase 09]: Every real match gets a chat room, and calls during the match are worth 5 points each; fans can make their own calls from the chat
- [Phase 10, 2026-10-30]: Email and password replace phone numbers and texted codes. Texts cost money for every sign-in, and without them nobody checked the number, so anyone could pose as another fan. An emailed code costs nothing through an ordinary Gmail account. Codes go through any SMTP server; `make api` reads `.env` through `INNINGS_ENV_FILE` and stays in dev mode, with the code on screen, until SMTP settings are there. Twilio is gone
- [Phase 10]: Passwords are PBKDF2-SHA256 at 600,000 iterations from Go's standard library. The rules follow NIST SP 800-63B: 8 to 128 characters, not the email address, not a common password, and no rules about mixing character types
- [Phase 10]: Resetting or changing a password signs out every other device. Deleting the account asks for the password, and a wrong current password is 403 `bad-password`, never 401, so the device stays signed in
- [Phase 10]: The migrations only add. Phone-era accounts keep their rows but can't sign in; `make reset-db` clears them
- [Phase 10]: Fans back up to five teams. `User.teamIds` lists them, main team first, and `favouriteTeamId` stays as the main team, so everything that shows one team keeps working. The picker offers the international sides and all 10 IPL sides
- [Phase 10]: The log's last step no longer offers a photo; photos are only for the profile picture

### Pending Todos

- Android: walk the match page's new panels on the emulator: the rating, "Pin to Top 5", the review box with the keyboard up, and the Top 5 editor. They were checked on the iPhone simulator only, and Android relies on `KeyboardAware`'s padding where iOS scrolls the field into view itself
- Real email delivery is untested; the mailer has only been tested against a mail server on the same machine. The developer won't put a personal email password into the project (2026-11-17), so codes show on screen in dev mode until a mail account is set up just for Innings (README → Real emails)
- Developer: test on a real iPhone and Android phone (`npx expo run:ios --device`, `npx expo run:android --device`), with the phone on the same Wi-Fi as `make api`
- Developer: store builds (EAS, App Store, Play Store) need the developer's Apple and Google accounts
- Photos are only for the profile picture, and they stay on the phone; uploading them is MEDIA-01
- Developer: confirm the stock colours for WPL, World Cup, BBL, SA20, The Hundred and Vault classics, and the darkened floodlights stocks (shown in the screenshots; no sign-off recorded)
- Walk every screen in floodlights; the Phase 6 review captured four of them in the dark theme
- Retake the README screenshots (most are from 2026-10-22, before chat, calls, past matches and the match page's rating) from `make demo`, which runs on real matches, so calls show whenever one is live; then record a short walkthrough for showing the app
- Six archive finals have a wrong stage label ("Match scorecard", "Chennai Super Kings innings"): Champions Trophy 2013 and 2017, IPL 2022 and 2023, and both 2026 T20 World Cup finals. Fix the stage parsing in `scripts/archive.py` and rebuild (it reads Wikipedia, so ask first)

### Blockers/Concerns

- Real matches come from ESPN's unofficial scores feed, which can change without notice; the feed's tests run on saved responses and need refreshing when it does
- Real email delivery hasn't been tried; until `.env` has SMTP settings, codes show on screen in dev mode
- Store builds need the developer's Apple and Google accounts
- The profile picture is a device-local file URI stored as a string; it doesn't upload, so other devices can't see it
- Sessions don't expire, and CORS allows any origin
- The per-IP and wrong-password limits live in memory, so a restart resets them; the per-address cap on code emails is kept in the database

## Deferred Items

Items acknowledged and carried forward:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Data | Real cricket API (DATA-01) | Built in v1.2 from ESPN's public scores feed | 2026-10-22 |
| Sign-in | Real text messages: SMS provider, sender ID, India DLT (SMS-01) | Replaced in v1.3 by emailed codes | 2026-10-25 |
| Sign-in | Sign in with Apple and Google (ACCT-09) | v2 | 2026-10-25 |
| Media | Server photo storage (MEDIA-01) | v2 | 2026-10-22 |
| Release | EAS and store submission (STORE-01) | v2, needs developer accounts | 2026-10-22 |
| Features | Live reactions, Season Wrapped, fan clubs, AI review summaries | v2 (plan's "Later") | 2026-10-22 |

## Session Continuity

Last session: 2026-11-17
Stopped at: the repo is public, with the README rewritten and `docs/PRD.md` and `docs/PROMPTS.md` added; sign-up codes stay on screen in dev mode, since the developer won't use a personal email password; `make demo` runs a ready-made demo account (`cmd/api/demo.go`) for showing the app
Resume file: none

- Run locally: `make api` (port 8088, SQLite file `./innings.db`; reads `.env` through `INNINGS_ENV_FILE`; sign-up codes show on screen in dev mode until `.env` has SMTP settings; `make reset-db` wipes the database) and `make app` (Metro on 8082; press `i` for iOS or `a` for Android). To sign up, use any email address and a password of at least 8 characters, then tap the dev code shown on the screen.
- `scripts/smoke.py` is the 88-check API run. Point it at a fresh dev-mode server with `INNINGS_MATCHES=sim`, on a spare port with a throwaway database, never at `make api`, which reads `.env`. It needs a live match, which the simulation always has and the real feed may not. It takes about 35 seconds, because it waits out the 30-second resend gap once.
- `python3 scripts/archive.py` rebuilds `internal/sim/archive.json` from Wikipedia's tournament pages. It goes online (to Wikipedia only), so ask the developer before running it.
- `make test` runs `go test ./...` and then Jest; `npx tsc --noEmit` and `npx expo lint` check the app.
- A build from before Phase 7 lacks sign-in's native modules, so build again (`npx expo run:android --port 8082`, `npx expo run:ios`). Phase 10 added none.
- iPhone: the "Innings QA" simulator. Android emulator: AVD `innings_pixel`; `adb reverse tcp:8082 tcp:8082` and `adb reverse tcp:8088 tcp:8088`; builds need `JAVA_HOME` set to JDK 17. The emulator also reaches the API at 10.0.2.2.
- On a phone, the app reaches the API through the Metro host's LAN address; `EXPO_PUBLIC_API_URL` overrides it.
- Screenshot tooling: start Metro with `EXPO_PUBLIC_QA_DRIVER=<url>` and the app polls that URL for a route or `theme:paper` / `theme:floodlights` (`src/lib/qaDriver.ts`).
