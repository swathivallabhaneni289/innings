# Roadmap: Innings

## Overview

Innings delivers a cricket fan's diary in six phases. Phase 1 fixes the shared JSON contract, the Expo scaffold, the first design system and the data layer. Phase 2 builds the Go API and the simulated cricket world it serves. Phase 3 builds the core loop (Matches tab, the Match Page in its three states, the Log flow and the ticket stub). Phase 4 adds the social layer and predictions on top. Phase 5 adds the Vault and the Profile. Phase 6 re-skins the app in the chosen Matchday Ticket direction and prepares the repo to be shown.

Phases 1 to 5 were planned and built on 2026-10-22 in one evening, by a lead agent directing parallel sub-agents with strict file ownership. The Phase 1 contract (`src/domain/types.ts`, `src/domain/api.ts`, `docs/API.md`) was the fixed interface every workstream built against.

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): planned milestone work
- Decimal phases (2.1, 2.2): urgent insertions (none so far)

- [x] **Phase 1: Foundation** - Expo scaffold, JSON contract, first design system, data layer (completed 2026-10-22)
- [x] **Phase 2: Go API & simulated world** - Simulation, store, guest auth, every endpoint, scoring (completed 2026-10-22)
- [x] **Phase 3: Core loop** - Matches tab, Match Page in three states, Log flow, ticket stub (completed 2026-10-22)
- [x] **Phase 4: Social & predictions** - Home, friends' feed, predictions with the lock, Fans' Player of the Match, Bragging Board (completed 2026-10-22)
- [x] **Phase 5: Vault & Profile** - Fact-checked classics, rankings, Top 5, profile stats, share card (completed 2026-10-22)
- [ ] **Phase 6: Design pass & release prep** - Direction C chosen 2026-10-22; re-skin under way; screenshot QA, icons, Android check, README, GitHub still to do

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

**Plans**: 0/1 plans executed (06-01 in progress)

Plans:

- [ ] 06-01-PLAN.md: Matchday Ticket re-skin, screenshot QA, icons, Android check, README, GitHub

**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 (2's two plans ran in parallel, as did 5's)

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 1/1 | Complete | 2026-10-22 |
| 2. Go API & simulated world | 2/2 | Complete | 2026-10-22 |
| 3. Core loop | 1/1 | Complete | 2026-10-22 |
| 4. Social & predictions | 1/1 | Complete | 2026-10-22 |
| 5. Vault & Profile | 2/2 | Complete | 2026-10-22 |
| 6. Design pass & release prep | 0/1 | In progress (direction C chosen, re-skin under way) | - |
