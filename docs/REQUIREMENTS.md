# Innings: Requirements

**Version:** v1.0 · **Status:** 64 of 64 requirements built · **Updated:** 25 October 2026

This document covers what Innings v1.0 has to do, how to check that it does it, and where each part is built. It comes from the product plan in [`PLAN.md`](PLAN.md). Where the plan left something open (how streaks multiply, when an upset counts), the decision is recorded here as a rule.

The requirement IDs match the build checklist in [`.planning/REQUIREMENTS.md`](../.planning/REQUIREMENTS.md), which the GSD roadmap tracks phase by phase.

## Progress

| Area | Built | Phase |
| --- | --- | --- |
| Home | 6 / 6 | 4 |
| Matches and the Match Page | 10 / 10 | 3, 4 |
| Predictions and bragging points | 7 / 7 | 2, 4 |
| Logging a match | 8 / 8 | 3, 5 |
| The Vault | 6 / 6 | 5 |
| Profile | 7 / 7 | 5 |
| Friends and social | 5 / 5 | 4, 5 |
| Design | 7 / 7 | 1, 3, 6 |
| Platform | 8 / 8 | 1, 2, 6 |
| **Total** | **64 / 64** | |

"Built" means the code does it and passes the check in its row. Real fans haven't used the app yet, and it hasn't been run on a physical phone. Both are the next step; see [Open items](#open-items).

## The problem

Cricbuzz tells you what happened. CricHeroes helps you score your own games. Nothing records how fans felt about a match: who they backed before the toss, how good it was, which player they'd crown, and where they were when it finished.

Innings v1.0 has to make that whole loop work for one fan and their friends:

1. **Before the first ball:** make your call (winner, top run-scorer, top wicket-taker).
2. **During:** follow the score and see which friends are watching.
3. **After:** rate the match and the players, say where you watched it, and get a ticket stub.
4. **Over time:** collect stubs, earn bragging points, and pin your all-time Top 5.

## Who it's for

A fan who watches most matches, usually with other people, and argues about them afterwards. They already have a scores app, so Innings shouldn't compete on scores. It has to be quicker to log a match than to type a group-chat message about it, and the stub has to look good enough to post.

## Scope

**In v1.0**

- Five tabs: Home, Matches, Log, Vault, Profile
- The Match Page before, during and after play
- Predictions, locked at the first ball
- Logging in four steps, the ticket stub and a shareable story card
- Friends, likes, replies, notifications and search
- 58 real classic matches
- iPhone and Android from one codebase
- A Go API with SQLite, guest accounts and a simulated cricket world

**Not in v1.0**

| Left out | Why |
| --- | --- |
| Live cricket data | No API key yet. A deterministic simulation stands in behind the same interface (DATA-01). |
| Email or social sign-in | Guest accounts are enough to test the loop (ACCT-01). |
| Photo upload | Photos stay on the device for now (MEDIA-01). |
| App Store and Play Store releases | These need the developer's Apple and Google accounts (STORE-01). |
| Live mood reactions, Season Wrapped, fan clubs, AI summaries of reviews | The plan marks these as later (LIVE-01, WRAP-01, CLUB-01, AI-01). |

## Functional requirements

Each table lists the requirement, a check that proves it, and where it lives.

- **Screens** are under `src/app/` and `src/features/`.
- **API handlers** are in `internal/httpapi/`.
- **Rules** are in `internal/domain/`.
- **Test names** run with `go test ./... -run <Name>`.

### Home

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| HOME-01 | Show the wordmark, with search and notifications on the right; notifications carry an unread badge | Open Home with unseen notifications: the badge shows a count, and opening Notifications clears it | `features/home/HomeHeader.tsx` |
| HOME-02 | Show an "On now" strip of live and today's matches with the live score. Before the start, a match offers Predict; at the first ball it reads Locked | The strip opens at the first ticket, live matches first. An upcoming ticket shows "Locks in" with a countdown and PREDICT; a live one shows the score | `features/home/OnNowStrip.tsx` |
| HOME-03 | Offer "Your call" on the next big match: winner, top run-scorer and top wicket-taker in three taps, with the friends' split | The card picks the highest-profile open match in the next three days. The picks save and the split bar shows friends' shares | `features/home/YourCallCard.tsx`, `yourCall()` in `reads.go` |
| HOME-04 | Show friends' stubs with rating, short review, fan Player of the Match and where they watched, with how many people | Each feed card shows all five, e.g. "📍 Hostel room, with 6 people" | `features/feed/FeedCard.tsx` |
| HOME-05 | List matches that finished in the last 24 hours, with the average fan rating and the top review | A match that ended 25 hours ago is not listed | `features/home/JustFinished.tsx` |
| HOME-06 | Rank the fan and the people they follow by points from the last 7 days | Points from a match that ended 8 days ago don't count | `features/home/BraggingBoard.tsx` |

### Matches and the Match Page

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| MATCH-01 | Filter by Live, Upcoming and Recent, and by India (men and women), IPL, WPL, World Cup, BBL, SA20 and The Hundred | Each chip narrows the list. An unknown state is rejected (`TestEveryGET` sends `state=bogus` and expects 400) | `features/matches/MatchFilters.tsx` |
| MATCH-02 | Group matches by day; each card shows teams, format, stage, venue, and the fan rating once finished | Headings read Today, Tomorrow, Yesterday or a date | `features/matches/MatchList.tsx`, `group.ts` |
| MATCH-03 | Before the start: teams, venue and start time, plus an "I'm watching" place picker | Setting a place saves it and shows it on the page (`TestWatching`) | `features/match/Header.tsx`, `Watching.tsx` |
| MATCH-04 | During the match: the live score and the friends watching now | The score moves with the clock (`TestLiveScoresThroughout`); friends watching are listed (`TestFriendsWatching`) | `features/match/LiveScoreboard.tsx` |
| MATCH-05 | After the match: the average fan rating with a chart of the 10 half-ball buckets | `TestBucketsAndAverage` | `features/match/FanRating.tsx` |
| MATCH-06 | Show fan player ratings out of 10 and crown the Fans' Player of the Match beside the official one, with "Fans disagree!" when they differ | A player needs at least 3 ratings to be crowned (`TestPlayerAveragesAndPotm`) | `features/match/PlayerRatings.tsx` |
| MATCH-07 | Show prediction results for the fan and friends, with the points each earned | `TestScoreOne` | `features/match/PredictionResults.tsx` |
| MATCH-08 | Show reviews in two lists: popular (most liked, all fans) and friends' | Popular is sorted by likes | `features/match/Reviews.tsx` |
| MATCH-09 | Summarise where fans watched as shares by place | Shares add up to 100% (`TestWatchShares`) | `features/match/WatchSummary.tsx` |
| MATCH-10 | Offer "Log this match" on finished matches and Vault classics | The button opens the Log flow with this match already picked | `features/match/LogAction.tsx` |

### Predictions and bragging points

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| PRED-01 | Let the fan pick winner, top run-scorer and top wicket-taker on the Match Page and on Home, show the friends' split, and allow changes until the lock | Changing a pick before the start saves the new one (`TestPredictionUpcoming`) | `features/predict/PredictionPicker.tsx`, `features/match/PredictionSection.tsx` |
| PRED-02 | Lock predictions at the first ball | After the start, `PUT /v1/matches/{id}/prediction` returns 409 `locked` (`TestPredictionLocked`, `TestIsLocked`), and the app shows the picks read-only | `domain/scoring.go`, `writes.go` |
| PRED-03 | Award points: winner +10, top run-scorer +15, top wicket-taker +15 | `TestScoreOne` | `domain/scoring.go` |
| PRED-04 | Add a +10 upset bonus when the winning team had under 30% of fan picks at lock time | `TestScoreOne` | `domain/scoring.go` |
| PRED-05 | Multiply points on a run of correct winners: ×1 for 1–2, ×1.5 for 3–4, ×2 for 5 or more. A wrong winner resets the run; a tie or no result leaves it alone | `TestStreakMultiplier`, `TestSettleHistoryOrdersByStartAndCarriesStreak` | `domain/scoring.go` |
| PRED-06 | Count every tied top scorer or wicket-taker as a correct pick | `TestScoreOne` | `domain/scoring.go` |
| PRED-07 | Show points on the Bragging Board, the Profile and the prediction results | The same total appears in all three places | `BraggingBoard.tsx`, `StatsRow.tsx`, `PredictionResults.tsx` |

### Logging a match

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| LOG-01 | Open the Log flow from the red ball in the middle of the tab bar, as a full-screen modal | Tap the ball from any tab | `app/(tabs)/_layout.tsx`, `app/log/index.tsx` |
| LOG-02 | Step 1: suggest matches live now or finished in the last 24 hours, with search for older matches and classics | Search for "2011" to find the World Cup final | `features/log/PickMatch.tsx` |
| LOG-03 | Step 2: rate from 0.5 to 5 balls in half-ball steps, with an optional "Instant classic" tag | The API rejects 0.3 or 5.5 with `bad-rating` (`TestValidRating`) | `features/log/RateMatch.tsx` |
| LOG-04 | Step 3: rate players in both XIs with a 1–10 slider, skip anyone, and crown your own Player of the Match | The API rejects a player who isn't in the match (`bad-player`) | `features/log/RatePlayers.tsx` |
| LOG-05 | Step 4: where you watched (7 places), a place note, how many people, tagged friends, an optional photo and an optional review | Places: Stadium, Home, Friend's place, Hostel or PG, Office, Café or pub, On the go | `features/log/WhereWatched.tsx` |
| LOG-06 | Print a ticket stub with the LOGGED stamp animation when the log is saved | The stamp lands on the stub after Save | `features/log/Printed.tsx`, `components/Motion.tsx` |
| LOG-07 | Share the stub as a 9:16 story image through the system share sheet | Tap Share story after saving | `features/share/StoryCard.tsx`, `useShareStory.ts` |
| LOG-08 | Update the stub when the same match is logged again, keeping its stub number | `TestLogCreateUpdateKeepsStub` | `writes.go` |

### The Vault

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| VAULT-01 | Show five collections: World Cup Finals, Greatest Chases, Test Match Thrillers, Super Over Madness, Women's Cricket Classics | `/v1/vault` returns all five (`TestEveryGET`) | `app/(tabs)/vault.tsx`, `app/vault/collection/[id].tsx` |
| VAULT-02 | Rank matches by fan rating of all time, filterable by format and era | `/v1/vault/rankings?format=T20&era=2020s` (`TestEveryGET`) | `app/vault/rankings.tsx` |
| VAULT-03 | Let fans log an old match with "I watched this" and keep the place note as the memory | Log a classic: the note shows on its stub | `features/vault/ClassicRow.tsx` |
| VAULT-04 | Show On This Day: a classic played on today's date, or the next anniversary, with a nudge to log it | The card shows the year and how long ago it was | `features/vault/OnThisDay.tsx` |
| VAULT-05 | Let fans pin and unpin up to five matches as their Top 5 | A sixth pin, a duplicate or an unknown match returns 400; an empty Top 5 is allowed (`TestTop5AndMe`) | `features/vault/top5.ts`, `Top5Pager.tsx` |
| VAULT-06 | Hold about 50 real classics, each fact-checked | 58 shipped. Each was checked by three independent reviewers, with sources in [`VAULT_SOURCES.md`](VAULT_SOURCES.md) (`TestClassics`) | `internal/sim/classics.json` |

### Profile

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| PROF-01 | Show photo, name, favourite team badge and a one-line fan bio | Initials on a colour show when there's no photo | `features/profile/ProfileHeader.tsx` |
| PROF-02 | Show the Top 5 as large ticket stubs | Read-only here; pinning happens in the Vault | `features/profile/Top5Strip.tsx` |
| PROF-03 | Show matches logged, average rating given, prediction accuracy and bragging points | `/v1/users/me` carries every stat (`TestEveryGET`) | `features/profile/StatsRow.tsx` |
| PROF-04 | Show every logged match as a stub | Each stub opens its stub page | `features/profile/StubGrid.tsx` |
| PROF-05 | Break down where the fan watched, and list the stadiums visited | Counts by place, then "Grounds visited" with the count and each ground | `features/profile/WatchBreakdown.tsx` |
| PROF-06 | Show the fan's rank among the people they follow, with badges | Badges: First Call, Called the Upset, 5-Match Streak, Sharpshooter, Stadium Regular (`TestBadges`) | `features/profile/PredictorRank.tsx` |
| PROF-07 | Let the fan edit photo, name, bio, favourite team and avatar colour | A blank name returns 400 (`TestTop5AndMe`); name 1–40 characters, bio up to 160 | `app/settings.tsx`, `features/profile/SettingsForm.tsx` |

### Friends and social

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| SOC-01 | Follow and unfollow from the Friends screen and from profiles | Follow toggles; you can't follow yourself (`TestFollowToggle`) | `app/friends.tsx`, `features/profile/FollowButton.tsx` |
| SOC-02 | Like, reply and say "I was there too" on any stub; the stub page shows the full log and the replies | Like toggles (`TestLikeToggleOnSimLog`); replies must be 1–500 characters | `app/stub/[id].tsx`, `features/feed/StubActions.tsx` |
| SOC-03 | Notify about predictions about to lock, friends' new logs, points scored, and replies and likes on your stubs | A match starting within 3 hours that you haven't called raises a lock reminder | `app/notifications.tsx`, `notifications` in `reads.go` |
| SOC-04 | Search teams, grounds, cities, stages, competitions, years and fans | "Wankhede", "IPL", "2011" and a friend's name all return results | `app/search.tsx`, `haystack()` in `reads.go` |
| SOC-05 | Open any fan's profile from the feed, replies or search | Tap a name | `app/user/[id].tsx`, `features/profile/ProfileView.tsx` |

### Design

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| DESIGN-01 | Offer a light default and a dark "floodlights" theme, chosen as Auto, Paper or Floodlights in Settings | The switch applies at once | `src/theme/`, `features/profile/AppearancePicker.tsx` |
| DESIGN-02 | Make the ticket stub the signature element: perforated, with a tear-off stub, teams, date, venue, rating and stub number | See [Design rules](#design-rules) | `components/TicketStub.tsx` |
| DESIGN-03 | Draw ratings as five cricket balls, including half-balls | 3.5 draws three full balls and one half | `components/BallRating.tsx` |
| DESIGN-04 | Spin the ball when a rating is set, and stamp LOGGED when a match is saved | | `components/Motion.tsx` |
| DESIGN-05 | Use three type roles: condensed display for codes and numbers, a text face for UI and reviews, mono for fine print | Big Shoulders Display, Schibsted Grotesk, IBM Plex Mono | `src/theme/tokens.ts` |
| DESIGN-06 | Apply the Matchday Ticket direction across every screen in both themes | No paper grain left (`grep -r grain src` finds nothing) | `src/theme/tokens.ts`, `src/components/` |
| DESIGN-07 | Check the screens from screenshots and fix what's wrong | 22 screens captured (4 in floodlights) and scored in a design review; all 12 prioritized fixes applied ([review](../.planning/phases/06-design-release/06-UI-REVIEW.md)) | |

### Platform

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| PLAT-01 | Run on iPhone and Android from one Expo codebase | Run as a development build on the iPhone 17 Pro simulator and a Pixel 8 emulator (Android 16) | `src/`, `app.json` |
| PLAT-02 | Serve a Go API with SQLite and embedded migrations, needing no setup | `make api` on a fresh clone creates `innings.db` and serves on port 8088 (`TestStoreRoundTrip`) | `cmd/api`, `internal/store` |
| PLAT-03 | Create or resume a guest with `POST /v1/bootstrap` and keep the token on the device | `TestBootstrapResumes`, `TestUnauthorized`, `TestBearerToken` | `internal/auth`, `src/data/store.ts` |
| PLAT-04 | Generate a deterministic cricket world (fixtures, XIs, live scores, friends, fans, crowd) with no API keys | The same anchor time gives the same world (`TestDeterminism`); live, upcoming and finished matches all exist (`TestStateCoverageAtAnchor`) | `internal/sim` |
| PLAT-05 | Keep one JSON contract in TypeScript, mirrored in Go, used through typed hooks | No endpoint returns `null` for a list (`TestNoNulls`, `TestEveryGET`) | `src/domain/`, `internal/domain/`, `src/data/index.ts` |
| PLAT-06 | Work on Android: icons, fonts, the API host and the system navigation bar | Home loads with data on the emulator; the emulator reaches the API at 10.0.2.2 | `src/data/client.ts` |
| PLAT-07 | Ship an app icon, an Android adaptive icon and a splash screen in the chosen look | A red ball on the concrete ground #ECEAE4 | `assets/`, `app.json` |
| PLAT-08 | Have a README and a GitHub repository | The README's run steps work from a fresh clone. The repository is private | `README.md` |

## Rules

### Bragging points

| Event | Points |
| --- | --- |
| Correct winner | +10 |
| Correct top run-scorer | +15 |
| Correct top wicket-taker | +15 |
| Upset: your winner had under 30% of fan picks when predictions locked | +10 |

The streak multiplier applies to the whole match total, upset bonus included. The streak counts correct winners in a row: ×1 for the first two, ×1.5 for the third and fourth, ×2 from the fifth. A wrong winner resets it; a tie or no result leaves it unchanged.

For example, the fifth correct winner in a row, calling an upset and the top scorer, earns (10 + 15 + 10) × 2 = 70 points.

### Other rules

- **Predictions lock at the first ball.** The fans' split for the upset bonus is frozen at that moment, so late picks can't move it.
- **Tied top performers.** If two batters tie for top score, a pick of either is correct.
- **Ratings.** Ratings run from 0.5 to 5 in half-ball steps, everywhere. The distribution chart has 10 bars.
- **Fans' Player of the Match.** This is the highest average rating among players with at least 3 ratings.
- **Bragging Board.** It counts points from matches that finished in the last 7 days.
- **Logging rules.**
  - A match can be logged once it has started.
  - Logging it again updates the same stub and keeps its number.
- **Top 5.** At most five matches, each pinned once.

### Input limits (enforced by the API)

| Field | Limit | Error code |
| --- | --- | --- |
| Match rating | 0.5–5 in steps of 0.5 | `bad-rating` |
| Player rating | 1–10 | `bad-player-rating` |
| Review | up to 2,000 characters | `too-long` |
| Place note | up to 200 characters | `too-long` |
| Reply | 1–500 characters | `bad-text` |
| Name | 1–40 characters | `bad-name` |
| Bio | up to 160 characters | `bad-bio` |
| Top 5 | at most 5, no duplicates, known matches only | `too-many`, `duplicate`, `unknown-match` |
| Prediction | winner must be one of the two teams; players must be in the match | `bad-pick` |
| Request body | 1 MB | |

## Non-functional requirements

### Platforms

- **Phones:** iPhone and Android from one Expo SDK 57 codebase (React Native 0.86, TypeScript strict).
- **Installing:** phones install a development build:
  - `npx expo run:ios --device`
  - `npx expo run:android --device`

  Expo Go from the app stores can't be used, because it only opens projects on the newest SDK.
- **Server:** Go 1.26 with the standard library `net/http` and pure-Go SQLite (`modernc.org/sqlite`). The SQL is kept portable to Postgres.

### Data

- **Matches:** every live match, fixture, fan and crowd number is simulated. The simulation is deterministic: `INNINGS_ANCHOR` pins the clock, so tests can assert exact numbers.
- **The Vault:** real matches only. An entry that can't be verified against a public scorecard is dropped.
- **The contract:** lists and maps are never `null` in the JSON; nullable fields are explicit. `src/domain/types.ts` and `src/domain/api.ts` define it, and `internal/domain` mirrors it.

### Security and privacy

- **Authentication:** every `/v1` route except `POST /v1/bootstrap` needs `Authorization: Bearer <token>`. Guest tokens come from `crypto/rand`.
- **Server limits:**
  - Request bodies are capped at 1 MB.
  - The server times out slow request headers after 10 seconds.
  - Every write is validated (see the limits above).
- **Photos:** log and profile photos stay on the device as file paths. They are never uploaded.
- **Known limits, accepted for a local build:**
  - The token is kept in AsyncStorage on the phone and stored as-is in SQLite.
  - CORS allows any origin.
  - Both get hardened with real accounts (ACCT-01).

### Accessibility

- Interactive elements carry accessibility roles (61) and labels (46).
- Filter chips are at least 40 pt tall.
- Ratings and split bars have spoken labels, e.g. "IND 62%, AUS 38%".
- Not done yet: a full pass with VoiceOver and TalkBack.

### Performance

- No load-time targets have been measured.
- Two fixes came from testing on Android:
  - The barcode on each ticket is drawn as a single shape.
  - Each Vault shelf shows 8 tickets.

  Together they made the Vault render quickly on the emulator.
- The plan's goal of logging a match in under 60 seconds is a design target, not a measurement.

### Running it

- `make api` runs with no setup on port 8088.
- `make app` starts Metro on port 8082.
- `make test` runs the Go tests; `npx tsc --noEmit` and `npx jest` check the app.

## Design rules

The look is called **Matchday Ticket**: every match is printed as a ticket on card stock.

- **Ground and cards:** the ground is concrete grey #ECEAE4. Cards are flat white, with no borders or shadows.
- **Ticket stock** is coloured by competition:
  - international: pale blue
  - IPL: salmon
  - Test: cream
  - WPL, World Cup, BBL, SA20, The Hundred and Vault classics each have their own stock
  - floodlights keeps the stocks, darkened
- **Ticket details:** a dotted perforation, two notches, and a tear-off stub with a red serial number and a barcode.
- **Colour roles:**

  | Colour | Hex | Used for |
  | --- | --- | --- |
  | Ink | #16171A | Text; primary buttons are solid ink |
  | Red | #D23A2E | Only for live, stamps, serials and the Log ball |
  | Green | #2F6B3A | Only for wins |

- **Type:**
  - Big Shoulders Display for team codes and big numbers
  - Schibsted Grotesk for UI and reviews
  - IBM Plex Mono for ticket fine print

## Later (v2)

| ID | Requirement |
| --- | --- |
| DATA-01 | Fixtures, squads and live scores from CricketData.org or SportMonks, behind the same `World` methods as the simulation |
| ACCT-01 | Email sign-in that keeps a guest's history, plus hardened token storage |
| MEDIA-01 | Upload log and profile photos so friends can see them |
| STORE-01 | EAS builds and App Store / Play Store releases |
| LIVE-01 | Live mood reactions during a match |
| WRAP-01 | Season Wrapped, an end-of-season summary to share |
| CLUB-01 | Team fan clubs |
| AI-01 | AI summaries of fan reviews |

## Open items

- [ ] Test on a real iPhone and a real Android phone
- [ ] Walk every screen in floodlights (the review captured four)
- [ ] Confirm the ticket stock colours for WPL, World Cup, BBL, SA20, The Hundred and Vault classics
- [ ] Add a way to delete a log in the app (the API already supports `DELETE /v1/logs/{id}`)
- [ ] Unit tests for the Top 5 helper and the Log draft
