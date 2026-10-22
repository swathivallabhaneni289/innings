# Innings

## What This Is

A cricket fan's diary for iPhone and Android. Every match is treated like a film release: fans predict before the first ball, log the match afterwards as a ticket stub (a half-ball rating out of 5, player ratings out of 10, their own Player of the Match, where they watched and who with, a photo and a review), and build a personal collection of stubs and an all-time Top 5. A Vault of 58 real, fact-checked classic matches keeps the app useful on days with no cricket. "Innings" is a working name; the source plan lists Over & Out, Second Innings and Stumped as alternatives.

## Core Value

Capture how fans felt about a match, not what happened in it. Cricbuzz covers the scorecard and CricHeroes covers scoring your own games; Innings covers the fan's take: the prediction, the rating, the memory of where they watched. This milestone (v1.0) delivers the full loop end to end against a simulated cricket world, runnable locally with no API keys.

## Requirements

### Validated

(None yet. Built and checked by tests and typecheck on 2026-10-22; not yet used by real fans.)

### Active

- [x] Fans can browse live, upcoming and recent matches and open a Match Page that changes with the match state (before, during, after)
- [x] Fans can predict winner, top run-scorer and top wicket-taker, locked at the first ball, and earn bragging points
- [x] Fans can log a match in four quick steps and get a printed ticket stub with a LOGGED stamp, shareable as a story card
- [x] Fans can follow friends, see their stubs on Home, and like, reply or tap "I was there too"
- [x] Fans can browse the Vault (collections, rankings, On This Day), log old matches and pin a Top 5
- [x] Fans have a profile: Top 5 stubs, stats, stub collection, where they watched, predictor rank and badges
- [ ] The app wears the chosen Matchday Ticket design across every screen in both themes (Phase 6)
- [ ] The repo is ready to show: icons, Android check, README, GitHub (Phase 6)

See REQUIREMENTS.md for the full checkable list (64 v1 requirements across 6 phases).

### Out of Scope

- Supabase: replaced by the Go API and SQLite (see Key Decisions)
- Live cricket data from a paid API: the simulation stands in until a CricketData.org or SportMonks key exists
- Email or social sign-in: guest accounts only for v1
- App Store and Play Store builds: need the developer's Apple and Google accounts
- The heritage scorebook look: built first, then superseded by direction C (see Key Decisions)
- Launch tactics from the plan (launch before a big series, seed with friends and the developer's AI club): not build tasks
- The plan's "Later" list (live mood reactions, Season Wrapped, team fan clubs, AI summaries of reviews): see REQUIREMENTS.md v2

## Context

- Source plan: the claude.ai doc "Innings — Cricket Fan App Plan", copied into `docs/PLAN.md` with a "Build decisions" table that settles the plan's open gaps.
- Planned and built on 2026-10-22, in one evening, by directing Claude Code: a lead agent plus parallel sub-agents, each owning a fixed set of files. Ownership notes are still visible in the code (for example the "CONTRACT (owned by the Home agent)" header in `src/features/feed/FeedCard.tsx` and the Vault seed note in `internal/sim/classics.go`).
- Same split as the developer's other project, RNDMRoll: an Expo app talking to a Go API.
- There is no cricket API key. `internal/sim` builds a deterministic world around server start: 29 fixtures (live, upcoming, just finished, recent) across internationals, IPL, WPL, World Cup, BBL, SA20 and The Hundred; real current playing XIs (65 teams, 374 players in the catalog); 8 friends and 16 public fans with about 640 logs; anonymous crowd ratings and prediction picks; live scores that move with the clock.
- The Vault holds 58 real matches. Each was checked against web sources by 3 independent checker agents (58 confirmed, 0 corrected, 0 dropped). One source per match is listed in `docs/VAULT_SOURCES.md`.
- The developer's verdict on the first visual pass (paper texture, ruled lines, Oswald, Source Serif, Plex Mono) was that it felt AI-made. Three directions were sketched on a design canvas: A Almanack (editorial, Wisden-like), B Pavilion Scoreboard (manual scoreboard), C Matchday Ticket (collectible tickets). C was chosen on 2026-10-22.

## Constraints

- **Backend**: Go 1.26, standard library `net/http`, SQLite through pure-Go `modernc.org/sqlite`, migrations embedded in the binary, SQL kept portable to Postgres. `make api` runs with zero setup.
- **Mobile**: Expo SDK 57, React Native 0.86, React 19, TypeScript (strict), Expo Router, TanStack Query, Zustand with AsyncStorage. One codebase for iPhone and Android.
- **Ports**: API on 8088 (8080 is taken by another local API), Metro on 8082.
- **Contract**: `src/domain/types.ts` and `src/domain/api.ts` are the JSON contract; Go mirrors them in `internal/domain`. Arrays and maps are never null.
- **Data honesty**: the Vault holds real, verified matches only; an entry that cannot be verified is dropped.
- **Design**: no generic AI look (the plan names dark mode by default, neon gradients and rounded cards everywhere). Direction C, Matchday Ticket, is the target from Phase 6 on.
- **Distribution**: tonight the app runs as a native iOS dev build on the simulator and in Expo Go on phones. Store builds wait on the developer's accounts.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Backend in Go | Developer asked for Go; RNDMRoll already uses the same Expo plus Go split | Built 2026-10-22 (`cmd/api`, `internal/`) |
| Phone app in Expo SDK 57, React Native, TypeScript | One codebase runs on iPhone and Android; Go cannot build native phone screens well | Built 2026-10-22 (`src/`) |
| Go API plus SQLite replaces Supabase from the plan | Zero-setup local runs; embedded migrations; SQL stays portable to Postgres for production | Built 2026-10-22 (`internal/store`) |
| Guest accounts via `POST /v1/bootstrap` token | Fastest path to a usable loop; token stored on the device | Built; email sign-in deferred to v2 |
| Simulated world instead of a cricket API | No API key available; deterministic, so tests can pin exact numbers | Built (`internal/sim`); real API swap is v2 |
| Streak multiplier x1 / x1.5 / x2 | Plan said "streaks multiply" without numbers: 1 to 2 correct winners x1, 3 to 4 x1.5, 5 or more x2 | Built (`internal/domain/scoring.go`) |
| Upset measured at lock time | The split must be frozen at the first ball or late picks could move it | Built |
| Tied top scorers all count | A fan who named either of two tied players called it right | Built |
| Half-balls everywhere (0.5 to 5) | One rating scale for logs, averages and the 10-bar distribution | Built |
| Fans' Player of the Match needs at least 3 ratings | Stops one rating crowning a player | Built (`internal/domain/aggregate.go`) |
| Bragging Board covers the last 7 days | "This week" made concrete: points from matches finished in the last 7 days | Built |
| Vault: real matches only, triple fact-check | The archive is the app's credibility on no-match days | 58 confirmed, 0 fixed, 0 dropped |
| Heritage scorebook look (first pass) | Taken from the plan's design section | Built in Phase 1; superseded 2026-10-22 |
| Direction C, Matchday Ticket | Developer's choice from three sketched directions after the first pass felt AI-made | Chosen 2026-10-22; not yet built (Phase 6) |
| Directions A Almanack and B Pavilion Scoreboard | Sketched on the design canvas alongside C | Not chosen |
| API on port 8088, Metro on 8082 | 8080 is used by another local API on the developer's machine | In use |

---
*Last updated: 2026-10-22 after the design direction choice*
