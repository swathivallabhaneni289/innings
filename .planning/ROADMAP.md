# Roadmap: Innings

## Overview

Innings delivers a cricket fan's diary in six phases. Phase 1 fixes the shared JSON contract, the Expo scaffold, the first design system and the data layer. Phase 2 builds the Go API and the simulated cricket world it serves. Phase 3 builds the core loop (Matches tab, the Match Page in its three states, the Log flow and the ticket stub). Phase 4 adds the social layer and predictions on top. Phase 5 adds the Vault and the Profile. Phase 6 re-skins the app in the chosen Matchday Ticket direction and prepares the repo to be shown.

The developer planned and built all six phases on 2026-10-22 in one evening by prompting Claude Code, which split the work across parallel sub-agents with strict file ownership. Milestone v1.0 is complete: 8 of 8 plans, 64 of 64 v1 requirements.

Milestone v1.1 (Accounts) added Phase 7 on 2026-10-25 at the developer's request: phone sign-in and a Beli-style onboarding, replacing the guest accounts. It is complete. Milestone v1.2 (Real matches) added Phase 8 on 2026-10-29: the app had shipped on simulated matches, and the plan asked for real data in week 1. Milestone v1.3 (Email sign-in and close-out) added Phase 10 on 2026-10-30: sign-in moved from phone numbers to email and password, fans can back up to five teams, and the loose ends from earlier phases were closed. The Phase 1 contract (`src/domain/types.ts`, `src/domain/api.ts`, `docs/API.md`) was the fixed interface every workstream built against.

After v1.3, two additions were built directly, without a phase. A quick task on 2026-11-13 added past matches (DATA-02): 2,948 games from every World Cup, Champions Trophy, Asia Cup and IPL season since 1973, in an archive Innings builds itself from Wikipedia's tournament pages. Later that day, at the developer's request, every finished match's page gained a tap-to-rate panel, "Pin to Top 5" for any match and a review box under Reviews, and the Top 5 gained an editor to reorder it or take games out (MATCH-10 and VAULT-05, reworded in place).

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): planned milestone work
- Decimal phases (2.1, 2.2): urgent insertions (none so far)

- [x] **Phase 1: Foundation** - Expo scaffold, JSON contract, first design system, data layer (completed 2026-10-22)
- [x] **Phase 2: Go API & simulated world** - Simulation, store, guest auth, every endpoint, scoring (completed 2026-10-22)
- [x] **Phase 3: Core loop** - Matches tab, Match Page in three states, Log flow, ticket stub (completed 2026-10-22)
- [x] **Phase 4: Social & predictions** - Home, friends' feed, predictions with the lock, Fans' Player of the Match, Bragging Board (completed 2026-10-22)
- [x] **Phase 5: Vault & Profile** - Fact-checked classics, rankings, Top 5, profile stats, share card (completed 2026-10-22)
- [x] **Phase 6: Design pass & release prep** - Matchday Ticket re-skin, UI review and fixes, icons, Android, README, GitHub (completed 2026-10-22)

**Milestone v1.1: Accounts**

- [x] **Phase 7: Phone sign-in & onboarding** - Phone number and text code, fan card with @username, find friends, sign out, delete account (completed 2026-10-25)

**Milestone v1.2: Real matches**

- [x] **Phase 8: Real matches** - Today's matches, the week's fixtures, squads, playing XIs and results from ESPN's public scores feed (completed 2026-10-29)
- [x] **Phase 9: During the match** - A chat room per match, and calls on how it goes while it's on (completed 2026-10-29)

**Milestone v1.3: Email sign-in and close-out**

- [x] **Phase 10: Email sign-in and close-out** - Email and password in place of phone numbers and texted codes, forgot and change password, up to five teams, and the earlier phases' loose ends (completed 2026-10-30)

## Phase Details

### Phase 1: Foundation

**Goal**: One codebase, one contract and one design system that every later workstream builds on
**Depends on**: Nothing (first phase)
**Requirements**: PLAT-01, PLAT-05, DESIGN-01, DESIGN-03, DESIGN-05
**Success Criteria** (what must be TRUE):

  1. The Expo SDK 57 app typechecks with zero errors (`npx tsc --noEmit`)
  2. Every entity and response shape exists once in TypeScript and is documented in `docs/API.md`
  3. Screens read data only through the hooks in `src/data/index.ts`
  4. Paper and floodlights palettes, the type scale and the core components exist and are exported from `src/components/index.ts`

**Plans**: 1/1 plans executed

Plans:

- [x] 01-01-PLAN.md: Scaffold, JSON contract, design system, data layer

**UI hint**: yes

### Phase 2: Go API & simulated world

**Goal**: A zero-setup Go API that serves a believable, deterministic cricket world merged with real guests' activity
**Depends on**: Phase 1 (the contract)
**Requirements**: PLAT-02, PLAT-03, PLAT-04, PRED-03, PRED-04, PRED-05, PRED-06
**Success Criteria** (what must be TRUE):

  1. `make api` starts the API on port 8088 with no configuration and no keys
  2. The same anchor time always produces the same world; live, upcoming, just-finished and recent fixtures exist at server start
  3. `POST /v1/bootstrap` creates or resumes a guest, who starts with demo history and follows 5 of 8 friends
  4. Every route in `docs/API.md` answers, and no array or map is ever encoded as null
  5. Scoring follows the Build decisions table exactly

**Plans**: 2/2 plans executed

Plans:
**Wave 1** *(parallel, built against the `World` interface)*

- [x] 02-01-PLAN.md: The simulated world (`internal/sim`)
- [x] 02-02-PLAN.md: Store, guest auth, HTTP API and scoring (`internal/store`, `internal/auth`, `internal/httpapi`, `internal/domain`)

### Phase 3: Core loop

**Goal**: A fan can find a match, read its page in any state, and log it as a ticket stub in under a minute
**Depends on**: Phase 1, Phase 2
**Requirements**: MATCH-01, MATCH-02, MATCH-03, MATCH-04, MATCH-05, MATCH-08, MATCH-09, MATCH-10, LOG-01, LOG-02, LOG-03, LOG-04, LOG-05, LOG-06, LOG-08, DESIGN-02, DESIGN-04
**Success Criteria** (what must be TRUE):

  1. The Matches tab filters by state and by competition or India, grouped by day
  2. The Match Page shows the right layout for upcoming, live and completed matches and for Vault classics
  3. The Log flow saves a stub in four steps (three for classics, which have no XI) and prints it with the LOGGED stamp
  4. Logging the same match again updates the stub and keeps its number

**Plans**: 1/1 plans executed

Plans:

- [x] 03-01-PLAN.md: Matches tab, Match Page, Log flow and ticket stub

**UI hint**: yes

### Phase 4: Social & predictions

**Goal**: Home becomes the place to see what's on, make the call, and see what friends thought
**Depends on**: Phase 3
**Requirements**: HOME-01, HOME-02, HOME-03, HOME-04, HOME-05, HOME-06, PRED-01, PRED-02, PRED-07, MATCH-06, MATCH-07, SOC-01, SOC-02, SOC-03, SOC-04
**Success Criteria** (what must be TRUE):

  1. A prediction can be made and changed until the first ball, then shows as Locked; the API refuses late picks with 409
  2. Home shows On Now, Your Call, Friends' Feed, Just Finished and the Bragging Board
  3. Finished matches crown a Fans' Player of the Match and flag "Fans disagree!" when it differs from the official pick
  4. Likes, replies, "I was there too" and follows work and show in notifications

**Plans**: 1/1 plans executed

Plans:

- [x] 04-01-PLAN.md: Home, feed, predictions, Fans' Player of the Match, Bragging Board, notifications, search

**UI hint**: yes

### Phase 5: Vault & Profile

**Goal**: The app stays useful on days with no cricket, and every fan has a profile worth showing
**Depends on**: Phase 3, Phase 4
**Requirements**: VAULT-01, VAULT-02, VAULT-03, VAULT-04, VAULT-05, VAULT-06, PROF-01, PROF-02, PROF-03, PROF-04, PROF-05, PROF-06, PROF-07, SOC-05, LOG-07
**Success Criteria** (what must be TRUE):

  1. The Vault holds about 50 real classics, each with a public source, in five collections
  2. Rankings filter by format and era; On This Day always shows a match; Top 5 pins up to five
  3. The Profile shows header, Top 5, stats, stub collection, where watched and predictor rank for any fan
  4. A saved stub can be shared as a story card image

**Plans**: 2/2 plans executed

Plans:
**Wave 1** *(parallel)*

- [x] 05-01-PLAN.md: Vault seed and triple fact-check (`internal/sim/classics.json`, `docs/VAULT_SOURCES.md`)
- [x] 05-02-PLAN.md: Vault and Profile screens, Settings, share card

**UI hint**: yes

### Phase 6: Design pass & release prep

**Goal**: The app looks deliberate rather than generated, in the Matchday Ticket direction, and the repo is ready to show
**Depends on**: Phase 5
**Requirements**: DESIGN-06, DESIGN-07, PLAT-06, PLAT-07, PLAT-08
**Success Criteria** (what must be TRUE):

  1. Every screen uses the Matchday Ticket tokens, ticket stocks and type, in both themes
  2. A screenshot of every screen in both themes has been reviewed and its defects fixed
  3. The app icon, adaptive icon and splash match the direction
  4. The app has been checked on an Android emulator or device
  5. A README describes the project and how to run it; the code is pushed to GitHub

*Outcome:* criteria 1, 3, 4 and 5 met. Criterion 2 was narrowed: 22 screens were captured and reviewed (4 of them in floodlights), and every prioritized fix was applied. The rest of the floodlights walk-through is a manual check in `06-VALIDATION.md`.

**Plans**: 1/1 plans executed

Plans:

- [x] 06-01-PLAN.md: Matchday Ticket re-skin, screenshot QA, icons, Android check, README, GitHub

**UI hint**: yes

### Phase 7: Phone sign-in & onboarding

**Goal**: Fans sign in with their phone number like on Beli, set up a fan card, find their friends, and can sign out or delete their account
**Depends on**: Phase 6
**Requirements**: ACCT-01, ACCT-02, ACCT-03, ACCT-04, ACCT-05, ACCT-06, ACCT-07, ACCT-08
**Success Criteria** (what must be TRUE):

  1. A new number gets a code, verifies, picks a name and a unique @username, finds friends and lands on Home
  2. The same number on another phone opens the same account
  3. Wrong, expired and reused codes and too many requests are refused with clear messages; an unconfigured server hands out no codes
  4. Sign out ends this device's session; delete account removes everything the fan made
  5. A brand-new account with no history looks right on every screen

*Outcome:* all five criteria met on the Android emulator and in a 70-check end-to-end API run. Real text messages still need an SMS provider account (SMS-01); in dev mode the code shows on screen. Phase 10 later replaced texts with emailed codes.

**Plans**: 1/1 plans executed

Plans:

- [x] 07-01-PLAN.md: Contract, server, app screens, Android run

**UI hint**: yes

### Phase 8: Real matches

**Goal**: The app shows the real cricket being played, so fans predict and log real matches
**Depends on**: Phase 7
**Requirements**: DATA-01
**Success Criteria** (what must be TRUE):

  1. Today's live matches show their real scores and status, refreshed within a minute or two
  2. The week's fixtures appear with their announced squads, so fans can predict days ahead
  3. The playing XIs replace the squads once the toss is done
  4. Finished matches settle predictions with the real winner, top scorers and top wicket-takers, and show the official player of the match
  5. No account, key or payment is needed, and the app never calls the feed itself

*Outcome:* all five met against ESPN's live feed on 29 October 2026 (England v Sri Lanka, South Africa v Australia and Bangladesh v West Indies live; UAE v Oman and a WBBL game settled; New Zealand v India with both squads). The feed is unofficial, so it may change without notice.

**Plans**: no separate plan file; built directly in `internal/feed`

**UI hint**: no (the app picks up new teams and players by reloading its catalog)

### Phase 10: Email sign-in and close-out

**Goal**: Fans sign up with an email and a password instead of a phone number, so sign-up costs nothing and every address is proven, and the loose ends from earlier phases are closed
**Depends on**: Phase 9
**Requirements**: ACCT-01 to ACCT-06 (reworded in place), ACCT-10, ACCT-11, ACCT-12; replaces SMS-01
**Success Criteria** (what must be TRUE):

  1. A new fan gives an email and a password, types the 6-digit code emailed to them, then sets up a fan card and finds friends; the account can't sign in until the code is in
  2. The same email and password open the same account on another phone
  3. Forgot password sets a new one with an emailed code; resetting or changing the password signs out every other device, and deleting the account asks for the password
  4. Wrong codes, wrong passwords and too many requests are refused with clear messages, and a server with neither mail settings nor dev mode sends no codes
  5. A fan backs up to five teams, main team first, chosen in sign-up and Settings
  6. Phone numbers are no longer collected, and contacts match on SHA-256 hashes of lower-cased email addresses

*Outcome:* all six met. The app was walked on the Innings QA iPhone simulator and the Android emulator in dev mode, where the code shows on screen. `go test ./...`, `go vet`, `npx tsc --noEmit` and `npx expo lint` pass, with the 13 lint errors from earlier phases fixed; 62 Jest tests pass (new ones cover the log draft, the Top 5, the sign-up checks, contact emails and teams); and `scripts/smoke.py` passes 88 checks against a fresh server. The phase also added "Delete this stub" on your own stub and took the photo picker out of the log's last step. Codes go out through any SMTP server, such as Gmail with an app password, but real delivery is untested until one is set in `.env`.

**Plans**: no separate plan file; built directly in `internal/auth`, `internal/httpapi`, `internal/store` and `src/app/(auth)`

**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 (2's two plans ran in parallel, as did 5's; 7's server and app ran in parallel against one contract)

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 1/1 | Complete | 2026-10-22 |
| 2. Go API & simulated world | 2/2 | Complete | 2026-10-22 |
| 3. Core loop | 1/1 | Complete | 2026-10-22 |
| 4. Social & predictions | 1/1 | Complete | 2026-10-22 |
| 5. Vault & Profile | 2/2 | Complete | 2026-10-22 |
| 6. Design pass & release prep | 1/1 | Complete | 2026-10-22 |
| 7. Phone sign-in & onboarding | 1/1 | Complete | 2026-10-25 |
| 8. Real matches | n/a | Complete | 2026-10-29 |
| 9. During the match | n/a | Complete | 2026-10-29 |
| 10. Email sign-in and close-out | n/a | Complete | 2026-10-30 |
