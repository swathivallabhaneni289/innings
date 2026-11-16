# Requirements: Innings

**Defined:** 2026-10-22
**Core Value:** Capture how fans felt about a match: the prediction, the rating, the memory of where they watched.

Derived from every bullet in `docs/PLAN.md`, with the gaps settled in its "Build decisions" table. A box is checked only where the code implements it.

## v1 Requirements

### Home

- [x] **HOME-01**: Top bar shows the wordmark with search and notifications (unread badge) on the right
- [x] **HOME-02**: "On Now" strip shows live and today's matches with teams and live score; a Predict button before the start that turns into Locked at the first ball
- [x] **HOME-03**: "Your Call" card picks the next big match; fan picks winner, top run-scorer and top wicket-taker in three taps and sees the friends' split
- [x] **HOME-04**: Friends' Feed shows friends' stubs with rating, short review, fan Player of the Match and where they watched (with how many people)
- [x] **HOME-05**: "Just Finished" lists matches that ended in the last 24 hours with the average fan rating and the top review
- [x] **HOME-06**: Bragging Board widget ranks the fan and the friends they follow by points from the last 7 days

### Matches

- [x] **MATCH-01**: Filter chips: Live, Upcoming, Recent, plus India (men and women), IPL, WPL, World Cup, BBL, SA20, The Hundred
- [x] **MATCH-02**: Match list grouped by day; cards show teams, format, stage, venue and the fan rating once finished
- [x] **MATCH-03**: Match Page before the start: header with teams, venue and start time, and an "I'm watching" place picker
- [x] **MATCH-04**: Match Page during the match: live score (simulated) and friends watching now
- [x] **MATCH-05**: Match Page after: big average fan rating with the 10 half-ball distribution
- [x] **MATCH-06**: Fan Player Ratings out of 10; Fans' Player of the Match (at least 3 ratings) beside the official pick, with "Fans disagree!" when they differ
- [x] **MATCH-07**: Prediction results for the fan and friends, with points awarded
- [x] **MATCH-08**: Reviews in two lists: popular (most liked, all fans) and friends'
- [x] **MATCH-09**: "Where fans watched" summary as shares by place
- [x] **MATCH-10**: On finished matches and classics: rate with a tap, "Log this match", pin to the Top 5, and write a review under Reviews

### Predictions

- [x] **PRED-01**: Prediction panel on the Match Page (and Home) with the friends' predictions split bar; picks can change until the lock
- [x] **PRED-02**: Predictions lock at the first ball: the API answers 409 `locked`, the app shows the picks read-only
- [x] **PRED-03**: Points: correct winner +10, top run-scorer +15, top wicket-taker +15
- [x] **PRED-04**: Upset bonus +10 when the winning pick had under 30% of fan picks at lock time
- [x] **PRED-05**: Streak multiplier on consecutive correct winners: 1 to 2 x1, 3 to 4 x1.5, 5 or more x2; a wrong winner resets, a tie leaves it
- [x] **PRED-06**: Tied top scorers or wicket-takers all count as correct
- [x] **PRED-07**: Points appear on the Bragging Board, the Profile and in prediction results

### Log

- [x] **LOG-01**: Centre tab is a big red ball that opens the Log flow as a full-screen modal
- [x] **LOG-02**: Step 1, pick the match: live now or ended in the last 24 hours, plus search for older matches and classics
- [x] **LOG-03**: Step 2, rate 0.5 to 5 balls in half-ball steps, with an optional "Instant classic" tag
- [x] **LOG-04**: Step 3, rate players: both XIs, a 1 to 10 slider each, skip any, crown your own Player of the Match
- [x] **LOG-05**: Step 4, where I watched it: 7 places, a place note, head count, tagged friends, optional review
- [x] **LOG-06**: Save prints a ticket stub with the LOGGED stamp animation
- [x] **LOG-07**: The stub can be shared as a 9:16 story card image through the share sheet
- [x] **LOG-08**: Logging the same match again updates the stub and keeps its stub number

### Vault

- [x] **VAULT-01**: Featured collections: World Cup Finals, Greatest Chases, Test Match Thrillers, Super Over Madness, Women's Cricket Classics
- [x] **VAULT-02**: All-time fan rankings, filterable by format and era
- [x] **VAULT-03**: "I watched this" logging for old matches, with the place note as the memory
- [x] **VAULT-04**: On This Day: a classic from today's date (or the next anniversary coming up), with a nudge to log it
- [x] **VAULT-05**: Your Top 5: pin and unpin up to five matches from any match page or the Vault, and edit it (reorder, take one out)
- [x] **VAULT-06**: The Vault is seeded with about 50 real classic matches, each fact-checked (58 shipped)

### Profile

- [x] **PROF-01**: Header with photo, name, a badge for each team the fan backs and one-line fan bio
- [x] **PROF-02**: Top 5 Matches shown as large ticket stubs
- [x] **PROF-03**: Stats row: matches logged, average rating given, prediction accuracy, bragging points
- [x] **PROF-04**: Stub Collection: every logged match
- [x] **PROF-05**: Where I've Watched breakdown and a stadiums-visited list
- [x] **PROF-06**: Predictor Rank among the fan and the people they follow, with badges (Called the Upset, 5-Match Streak and others)
- [x] **PROF-07**: Edit the fan card: photo, name, bio, the teams they back, avatar colour

### Social

- [x] **SOC-01**: Follow and unfollow friends from the Friends screen and from profiles
- [x] **SOC-02**: Like, reply and "I was there too" on any stub; a stub page shows the full log and replies
- [x] **SOC-03**: Notifications: predictions about to lock, friends' new logs, points scored, replies and likes on my stubs
- [x] **SOC-04**: Search across teams, grounds, cities, stages, competitions, years and fans
- [x] **SOC-05**: Open any fan's profile from the feed, replies or search

### Design

- [x] **DESIGN-01**: Two themes: a light default and a dark "floodlights" mode, with Auto, Paper and Floodlights in Settings
- [x] **DESIGN-02**: The ticket stub is the signature element: perforated, torn stub with teams, date, venue, rating and stub number
- [x] **DESIGN-03**: Ratings drawn as five cricket balls, half-balls included
- [x] **DESIGN-04**: Motion: a ball spin when a rating is set, a stamp when a match is saved
- [x] **DESIGN-05**: Three-role type system: condensed display for scores and headings, a text face for reviews, mono for stats
- [x] **DESIGN-06**: Direction C, Matchday Ticket, applied across the design system and every screen in both themes
- [x] **DESIGN-07**: Screenshot QA with defects fixed: 22 screens captured (4 in floodlights) and scored in a 6-pillar review, every prioritized fix applied

### Platform

- [x] **PLAT-01**: One Expo codebase runs on iPhone and Android as a native development build (Expo Go from the stores only opens the newest SDK)
- [x] **PLAT-02**: Go API with SQLite, embedded migrations and Postgres-portable SQL; `make api` needs no setup
- [x] **PLAT-03**: Guest accounts: `POST /v1/bootstrap` creates or resumes a guest; the token lives on the device *(v1.0; replaced in v1.1 by phone sign-in and in v1.3 by email and password, ACCT-01 to ACCT-07)*
- [x] **PLAT-04**: A deterministic simulated world (fixtures, XIs, live scores, friends, fans, crowd) needs no API keys
- [x] **PLAT-05**: One JSON contract in TypeScript, mirrored in Go, consumed through typed react-query hooks
- [x] **PLAT-06**: Android readiness: the app checked on an Android emulator or device (icons, fonts, API host, system navigation)
- [x] **PLAT-07**: App icon, adaptive icon and splash in the chosen direction
- [x] **PLAT-08**: README and the GitHub repository (private, `swathivallabhaneni289/innings`)

## v1.1 Requirements: Accounts

Added 2026-10-25 when the developer asked for "a proper sign in like how Beli has". Moved from a phone number and a texted code to email and password on 2026-10-30 (Phase 10, v1.3); the surviving requirements kept their IDs.

- [x] **ACCT-01**: Sign up with an email address and a password, confirmed by a 6-digit code emailed to that address; then sign in with them on any number of phones *(a phone number and a texted code until 2026-10-30)*
- [x] **ACCT-02**: Codes expire after 10 minutes and allow 5 tries; a new code after 30 seconds, at most 5 an hour per address and 20 code emails an hour per IP address; 10 wrong passwords per account in 15 minutes, or 50 per IP address in an hour, make sign-in wait
- [x] **ACCT-03**: Codes are emailed through an SMTP server when one is configured (Gmail with an app password works); in dev mode (`make api` with no SMTP settings in `.env`) the code shows on screen and in the server log; with neither, no codes are sent *(replaced texting them through Twilio Verify)*
- [x] **ACCT-04**: New fans set up a fan card: name, a unique @username checked as they type, the teams they back (ACCT-12) and an optional photo
- [x] **ACCT-05**: Find friends from contacts (email addresses hashed on the phone, matched on the server, never stored), invite the rest through the share sheet, and follow suggested fans
- [x] **ACCT-06**: Sign out of this phone; delete the account and everything it made, after typing the password again
- [x] **ACCT-07**: The session token lives in the phone's secure storage and only its hash on the server; a signed-out or unknown token sends the fan back to sign-in
- [x] **ACCT-08**: A brand-new account with no history looks right on every screen

## v1.2 Requirements: Real matches

Added 2026-10-29: the plan asked for a cricket data API in week 1; nothing paid, no sign-ups.

- [x] **LIVE-03**: Fans make their own calls from the chat (a batter's 30/50/100, a team's total, the winner), refused once decided, settled like the app's
- [x] **LIVE-02**: Calls during a match (where will they finish, will they chase it, will he make 50), locking at set points and settling on their own, 5 points each on the Bragging Board
- [x] **CHAT-01**: A chat room for every real match, open from a day before the first ball until a day after it, with the score above and new messages within seconds
- [x] **DATA-01**: Today's real matches with live scores, the week's fixtures with announced squads, playing XIs at the toss, and results with top scorers, top wicket-takers and player of the match, from ESPN's public scores feed (no key)

## v1.3 Requirements: Email sign-in and close-out

Added 2026-10-30, when sign-in moved to email and password (ACCT-01 to ACCT-06 above were reworded in place).

- [x] **ACCT-10**: Reset a forgotten password with a 6-digit code emailed to the address; the new password signs this phone in and every other device out
- [x] **ACCT-11**: Change the password in Settings, given the current one; every other device is signed out and this one stays in
- [x] **ACCT-12**: Back up to five teams, main team first, chosen in sign-up and Settings from the international sides and all 10 IPL sides

## v1.4 Requirements: Past matches

Added 2026-11-13: fans couldn't find the older matches they wanted to log, and ESPN's feed only answers for recent days.

- [x] **DATA-02**: Find and log any match from every World Cup (men's and women's, ODI and T20), Champions Trophy, Asia Cup and IPL season since 1973, from an archive Innings builds itself with Wikipedia's tournament pages as the reference

## v2 Requirements

From the plan's "Later" list, plus items deferred by build decisions on 2026-10-22.

### Later (from the plan)

- **LIVE-01**: Live mood reactions during a match
- **WRAP-01**: Season Wrapped, an end-of-season shareable summary (the Profile shows a "Later this season" placeholder)
- **CLUB-01**: Team fan clubs
- **AI-01**: AI summaries of fan reviews

### Deferred by decision

- **SMS-01**: *Replaced 2026-10-30 (Phase 10): codes go by email instead (ACCT-01, ACCT-03), so no SMS provider is needed.* It was real text messages, with an SMS provider account, a sender ID and, for Indian numbers, DLT registration
- **ACCT-09**: Sign in with Apple and Google
- **MEDIA-01**: Upload photos to server storage: the profile picture (today a device-local file URI), and photos on stubs if the log asks for them again
- **STORE-01**: EAS builds and App Store / Play Store submission

## Out of Scope

| Feature | Reason |
|---------|--------|
| Supabase | Replaced by the Go API and SQLite |
| Heritage scorebook look (paper grain, ruled lines, Oswald / Source Serif) | Built in Phase 1, superseded by direction C on 2026-10-22 |
| Directions A Almanack and B Pavilion Scoreboard | Sketched, not chosen |
| Launch tactics (timing, seeding, growth) | Not build tasks for this milestone |
| Phone numbers and texted codes | Replaced by email and password in Phase 10 (2026-10-30) |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| PLAT-01, PLAT-05 | Phase 1 | Complete |
| DESIGN-01, DESIGN-03, DESIGN-05 | Phase 1 | Complete |
| PLAT-02, PLAT-03, PLAT-04 | Phase 2 | Complete |
| PRED-03, PRED-04, PRED-05, PRED-06 | Phase 2 | Complete |
| MATCH-01, MATCH-02, MATCH-03, MATCH-04, MATCH-05, MATCH-08, MATCH-09, MATCH-10 | Phase 3 | Complete |
| LOG-01, LOG-02, LOG-03, LOG-04, LOG-05, LOG-06, LOG-08 | Phase 3 | Complete |
| DESIGN-02, DESIGN-04 | Phase 3 | Complete |
| HOME-01, HOME-02, HOME-03, HOME-04, HOME-05, HOME-06 | Phase 4 | Complete |
| PRED-01, PRED-02, PRED-07 | Phase 4 | Complete |
| MATCH-06, MATCH-07 | Phase 4 | Complete |
| SOC-01, SOC-02, SOC-03, SOC-04 | Phase 4 | Complete |
| VAULT-01, VAULT-02, VAULT-03, VAULT-04, VAULT-05, VAULT-06 | Phase 5 | Complete |
| PROF-01, PROF-02, PROF-03, PROF-04, PROF-05, PROF-06, PROF-07 | Phase 5 | Complete |
| SOC-05, LOG-07 | Phase 5 | Complete |
| DESIGN-06, DESIGN-07 | Phase 6 | Complete |
| PLAT-06, PLAT-07, PLAT-08 | Phase 6 | Complete |
| ACCT-01, ACCT-02, ACCT-03, ACCT-04, ACCT-05, ACCT-06, ACCT-07, ACCT-08 | Phase 7 (ACCT-01 to ACCT-06 reworked in Phase 10) | Complete |
| DATA-01 | Phase 8 | Complete |
| CHAT-01, LIVE-02, LIVE-03 | Phase 9 | Complete |
| ACCT-10, ACCT-11, ACCT-12 | Phase 10 | Complete |
| SMS-01 (v2) | Phase 10 | Replaced by ACCT-01 and ACCT-03 |
| DATA-02 | Quick task, 2026-11-13 | Complete |

**Coverage:**

- v1 requirements: 64 total (64 complete, 0 pending)
- v1.1 requirements: 8 total (8 complete, 0 pending)
- v1.2 requirements: 4 total (4 complete, 0 pending)
- v1.3 requirements: 3 total (3 complete, 0 pending)
- v1.4 requirements: 1 total (1 complete, 0 pending)
- Mapped to phases: 79, plus DATA-02 done as a quick task
- Unmapped: 0

---
*Requirements defined: 2026-10-22*
*Last updated: 2026-11-13, adding DATA-02 (past matches) after Phase 10. The readable version with acceptance checks is `docs/REQUIREMENTS.md`.*
