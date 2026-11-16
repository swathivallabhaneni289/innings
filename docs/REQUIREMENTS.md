# Innings: Requirements

**Version:** v1.4 · **Status:** 80 of 80 requirements built · **Updated:** 16 November 2026

This document covers what Innings has to do, how to check that it does it, and where each part is built. It comes from the product plan in [`PLAN.md`](PLAN.md). Where the plan left something open (how streaks multiply, when an upset counts), the decision is recorded here as a rule.

This is the project's only requirements list. The problem Innings solves, who it's for, its goals and what's out of scope are in the [PRD](PRD.md). The [Progress](#progress) table shows which phase built each area, and the [roadmap](../.planning/ROADMAP.md) has the detail of each phase.

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
| Accounts (v1.1, v1.3) | 11 / 11 | 7, 10 |
| Real matches (v1.2, v1.4) | 2 / 2 | 8, quick task |
| During the match (v1.2) | 3 / 3 | 9 |
| **Total** | **80 / 80** | |

"Built" means the code does it and passes the check in its row. Real fans haven't used the app yet, and it hasn't been run on a physical phone. Both are the next step; see [Open items](#open-items).

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
| MATCH-09 | Summarise where fans watched as shares by place | Each place's share of all logs, largest first (`TestWatchShares`) | `features/match/WatchSummary.tsx` |
| MATCH-10 | On finished matches and Vault classics, let the fan rate the match with a tap, open "Log this match", pin it to the Top 5, and write a review under Reviews | A tap on the balls saves a log at once, watched where the fan logged from last (else Home), and says so; the button opens the Log flow with this match already picked; the review box posts a review, rating it too if it isn't rated yet, and saving it again keeps the stub number (`TestArchiveMatches`); `relog` keeps the rest of a log as it was (`draft.test.ts`) | `features/match/YourTake.tsx`, `ReviewBox.tsx`, `features/log/draft.ts` |

### Predictions and bragging points

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| PRED-01 | Let the fan pick winner, top run-scorer and top wicket-taker on the Match Page and on Home, show the friends' split, and allow changes until the lock | A pick saved before the start shows on the Match Page, and a winner who isn't playing is rejected (`TestPredictionUpcoming`). Saving again replaces the pick | `features/predict/PredictionPicker.tsx`, `features/match/PredictionSection.tsx` |
| PRED-02 | Lock predictions at the first ball | After the start, `PUT /v1/matches/{id}/prediction` returns 409 `locked` (`TestPredictionLocked`, `TestIsLocked`), and the app shows the picks read-only | `domain/scoring.go`, `writes.go` |
| PRED-03 | Award points: winner +10, top run-scorer +15, top wicket-taker +15 | `TestScoreOne` | `domain/scoring.go` |
| PRED-04 | Add a +10 upset bonus when the winning team had under 30% of fan picks at lock time | `TestScoreOne` | `domain/scoring.go` |
| PRED-05 | Multiply points on a run of correct winners: ×1 for 1–2, ×1.5 for 3–4, ×2 for 5 or more. A wrong winner resets the run; a tie or no result leaves it alone | `TestStreakMultiplier`, `TestSettleHistoryOrdersByStartAndCarriesStreak` | `domain/scoring.go` |
| PRED-06 | Count every tied top scorer or wicket-taker as a correct pick | `TestScoreOne` | `domain/scoring.go` |
| PRED-07 | Show points on the Bragging Board, the Profile and the prediction results | The same total appears in all three places | `BraggingBoard.tsx`, `StatsRow.tsx`, `PredictionResults.tsx` |

### Logging a match

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| LOG-01 | Open the Log flow from the red ball in the middle of the tab bar, as a full-screen modal | Tap the ball from any tab | `app/(app)/(tabs)/_layout.tsx`, `app/(app)/log/index.tsx` |
| LOG-02 | Step 1: suggest matches live now or finished in the last 24 hours, with search for older matches and classics | Search for "2011" to find the World Cup final | `features/log/PickMatch.tsx` |
| LOG-03 | Step 2: rate from 0.5 to 5 balls in half-ball steps, with an optional "Instant classic" tag | The API rejects 0.3 or 5.5 with `bad-rating` (`TestValidRating`) | `features/log/RateMatch.tsx` |
| LOG-04 | Step 3: rate players in both XIs with a 1–10 slider, skip anyone, and crown your own Player of the Match | The API rejects a player who isn't in the match (`bad-player`) | `features/log/RatePlayers.tsx` |
| LOG-05 | Step 4: where you watched (7 places), a place note, how many people, tagged friends and an optional review | Places: Stadium, Home, Friend's place, Hostel or PG, Office, Café or pub, On the go | `features/log/WhereWatched.tsx` |
| LOG-06 | Print a ticket stub with the LOGGED stamp animation when the log is saved | The stamp lands on the stub after Save | `features/log/Printed.tsx`, `components/Motion.tsx` |
| LOG-07 | Share the stub as a 9:16 story image through the system share sheet | Tap Share story after saving | `features/share/StoryCard.tsx`, `useShareStory.ts` |
| LOG-08 | Update the stub when the same match is logged again, keeping its stub number | `TestLogCreateUpdateKeepsStub` | `writes.go` |

### The Vault

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| VAULT-01 | Show five collections: World Cup Finals, Greatest Chases, Test Match Thrillers, Super Over Madness, Women's Cricket Classics | `/v1/vault` returns all five (`TestEveryGET`) | `app/(app)/(tabs)/vault.tsx`, `app/(app)/vault/collection/[id].tsx` |
| VAULT-02 | Rank matches by fan rating of all time, filterable by format and era | `/v1/vault/rankings?format=T20&era=2020s` (`TestEveryGET`) | `app/(app)/vault/rankings.tsx` |
| VAULT-03 | Let fans log an old match with "I watched this" and keep the place note as the memory | Log a classic: the note shows on its stub | `features/vault/ClassicRow.tsx` |
| VAULT-04 | Show On This Day: a classic played on today's date, or the next anniversary, with a nudge to log it | The card shows the year and how long ago it was | `features/vault/OnThisDay.tsx` |
| VAULT-05 | Let fans pin and unpin up to five matches as their Top 5, from any match page or the Vault, and edit it: move a game up or down, or take one out | A sixth pin or an unknown match returns 400 and an empty Top 5 is allowed (`TestTop5AndMe`); a duplicate is refused in `writes.go`; an archive match and a recent one pin in order (`TestArchiveMatches`); Edit changes nothing until Save | `features/vault/top5.ts`, `Top5Pager.tsx`, `Top5Editor.tsx`, `app/(app)/top5.tsx` |
| VAULT-06 | Hold about 50 real classics, each fact-checked | 58 shipped. Each was checked three times against public scorecards by independent checker agents, with sources in [`VAULT_SOURCES.md`](VAULT_SOURCES.md) (`TestClassics`) | `internal/sim/classics.json` |

### Profile

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| PROF-01 | Show photo, name, a badge for each team the fan backs and a one-line fan bio | Initials on a colour show when there's no photo | `features/profile/ProfileHeader.tsx` |
| PROF-02 | Show the Top 5 as large ticket stubs | Pinning happens on any match page or in the Vault; Edit opens the Top 5 editor | `features/profile/Top5Strip.tsx` |
| PROF-03 | Show matches logged, average rating given, prediction accuracy and bragging points | `/v1/users/me` carries every stat (`TestEveryGET`) | `features/profile/StatsRow.tsx` |
| PROF-04 | Show every logged match as a stub | Each stub opens its stub page | `features/profile/StubGrid.tsx` |
| PROF-05 | Break down where the fan watched, and list the stadiums visited | Counts by place, then "Grounds visited" with the count and each ground | `features/profile/WatchBreakdown.tsx` |
| PROF-06 | Show the fan's rank among the people they follow, with badges | Badges: First Call, Called the Upset, 5-Match Streak, Sharpshooter, Stadium Regular (`TestBadges`) | `features/profile/PredictorRank.tsx` |
| PROF-07 | Let the fan edit photo, name, bio, the teams they back and avatar colour | A blank name returns 400 (`TestTop5AndMe`); name 1–40 characters, bio up to 160 | `app/(app)/settings.tsx`, `features/profile/SettingsForm.tsx` |

### Friends and social

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| SOC-01 | Follow and unfollow from the Friends screen and from profiles | Follow toggles; you can't follow yourself (`TestFollowToggle`) | `app/(app)/friends.tsx`, `features/profile/FollowButton.tsx` |
| SOC-02 | Like, reply and say "I was there too" on any stub; the stub page shows the full log and the replies | Like toggles (`TestLikeToggleOnSimLog`); replies must be 1–500 characters | `app/(app)/stub/[id].tsx`, `features/feed/StubActions.tsx` |
| SOC-03 | Notify about predictions about to lock, friends' new logs, points scored, and replies and likes on your stubs | A match starting within 3 hours that you haven't called raises a lock reminder | `app/(app)/notifications.tsx`, `notifications` in `reads.go` |
| SOC-04 | Search teams, grounds, cities, stages, competitions, years and fans | "Wankhede", "IPL", "2011" and a friend's name all return results | `app/(app)/search.tsx`, `haystack()` in `reads.go` |
| SOC-05 | Open any fan's profile from the feed, replies or search | Tap a name | `app/(app)/user/[id].tsx`, `features/profile/ProfileView.tsx` |

### Accounts (v1.1, email sign-in in v1.3)

Added when the developer asked for "a proper sign in like how Beli has". v1.1 built it on a phone number and a texted code. Phase 10 (v1.3) moved it to an email address and a password, with a 6-digit code emailed to confirm the address, because texts cost money and a phone number nobody checked would let someone pose as another fan (see [`PLAN.md`](PLAN.md) → Build decisions). The surviving requirements kept their IDs; ACCT-10 to ACCT-12 are new.

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| ACCT-01 | Sign up with an email address and a password of 8 to 128 characters (not the address, not a common password), confirmed by a 6-digit code emailed to that address; then sign in with them on any number of phones | The account can't sign in until the code is in, and a second device gets the same user and its own session (`TestSignUpAndSecondDevice`); password rules (`TestSignUpRules`, `TestPasswordProblem`); walked on the Innings QA iPhone simulator and the Android emulator | `app/(auth)/sign-up.tsx`, `app/(auth)/confirm.tsx`, `app/(auth)/sign-in.tsx`, `internal/auth/flows.go` |
| ACCT-02 | Refuse bad codes, wrong passwords and too many requests: codes last 10 minutes and allow 5 tries; a new code after 30 seconds, at most 5 an hour per address and 20 code emails an hour per IP; 10 wrong passwords for an account in 15 minutes, or 50 from one IP in an hour, make sign-in wait | `TestWrongCodeCountsDown`, `TestCodeExpires`, `TestResendConfirmation`, `TestHourlyCapPerAddress`, `TestIPCap`, `TestWrongPasswordCap`, `TestSignInErrors`, `TestCodeLimitsAndTries`, `TestFailureCounting` | `internal/auth/codes.go`, `internal/auth/flows.go`, `internal/auth/ratelimit.go`, `internal/store/accounts.go` |
| ACCT-03 | Email codes through an SMTP server when one is configured (Gmail with an app password works); in dev mode show the code on screen and in the server log; with neither, send no codes. This replaced texting them (SMS-01) | Against a mail server on the same machine: sends, and refuses a server without STARTTLS or a bad login (`TestSMTPMailerSends`, `TestSMTPMailerRefusesPlainTextAndBadLogins`); half an SMTP setup stops the server (`TestMailerFromEnv`); `TestNoMailer`, `TestCodeEmail`, `TestLoadEnvFile`. Real delivery hasn't been tried yet | `internal/auth/mail.go`, `cmd/api/main.go`, `cmd/api/env.go` |
| ACCT-04 | Let a new fan set up a fan card: name, a unique @username checked as they type, the teams they back (ACCT-12) and an optional photo | Usernames follow the rules and can't clash with any fan, real or simulated (`TestHandleRules`, `TestHandles`); sign-up can't finish without a name and username (`TestOnboarding`) | `app/onboarding/profile.tsx`, `features/onboarding/HandleField.tsx` |
| ACCT-05 | Find friends from contacts, invite the rest, and suggest fans to follow; contacts' email addresses are hashed on the phone, matched on the server and never stored | `TestContactsMatch`; addresses are trimmed, lower-cased and de-duplicated, and the fan's own is left out (`contactEmails.test.ts`) | `features/onboarding/ContactsFinder.tsx`, `features/onboarding/contactEmails.ts` |
| ACCT-06 | Sign out of this phone, and delete the account with everything it made once the fan types their password again | Sign out ends only this session (`TestLogoutEndsOnlyThisSession`); delete refuses a wrong password with 403 `bad-password`, then removes the fan's stubs, calls, follows, likes and replies, signs out every device, and other fans' screens still load (`TestDeleteAccount`, `TestDeleteUser`) | `features/auth/AccountSection.tsx`, `app/(app)/delete-account.tsx` |
| ACCT-07 | Keep the session token in the phone's secure storage and only its hash on the server; send a fan with a signed-out or unknown token back to sign-in | `TestSessionsAndPasswords`, `TestBootstrapResumes`, `TestUnauthorized`; walked: after the database was reset, the iPhone app went back to Welcome at launch | `src/data/session.ts`, `internal/auth/auth.go` |
| ACCT-08 | Look right on a brand-new account with no history | Walked on the Android emulator and an iPhone simulator: Home, Profile, the Bragging Board and Predictor Rank all have empty states, with "–" for averages | `features/home/BraggingBoard.tsx`, `features/profile/PredictorRank.tsx` |
| ACCT-10 | Reset a forgotten password with a 6-digit code emailed to the address; the new password signs this phone in and every other device out, and confirms an address that never was | `TestForgotPassword`, `TestForgotPasswordConfirms`; the reset token lasts 15 minutes and works once (`TestResetTokens`) | `app/(auth)/forgot.tsx`, `app/(auth)/reset-code.tsx`, `app/(auth)/new-password.tsx`, `internal/auth/flows.go` |
| ACCT-11 | Change the password in Settings, given the current one; every other device is signed out and this one stays in | `TestChangePassword`, `TestSessionsAndPasswords` | `app/(app)/change-password.tsx`, `features/auth/AccountSection.tsx` |
| ACCT-12 | Let a fan back up to five teams, main team first, picked in sign-up and Settings from the international sides and all 10 IPL sides; the main team stays the favourite team | `PUT /v1/me` takes `teamIds` and refuses unknown, repeated or too many teams (`TestTeams`); picking, dropping and choosing the main team (`teams.test.ts`) | `features/onboarding/TeamChips.tsx`, `features/onboarding/teams.ts`, `internal/httpapi/writes.go` |

### Real matches (v1.2)

Added 2026-10-29. The developer's plan asked for a cricket data API in week 1, and the app had shipped on simulated matches; the developer wants nothing paid and no sign-ups.

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| DATA-01 | Show today's real matches with live scores, the week's fixtures with their announced squads, the playing XIs once the toss is done, and results with the top scorers, top wicket-takers and player of the match; no account or key needed | Tests on saved ESPN responses from 29 October 2026 (`TestScoreboardMapping`, `TestPollFetchesSummariesSavesAndReloads`, `TestSquadsAndTossAreCheckedOftenEnough`, `TestAPIServesRealMatches`); walked on an iPhone simulator, where Home and Matches showed South Africa v Australia, Bangladesh v West Indies and England v Sri Lanka live | `internal/feed/`, `cmd/api/main.go` |
| DATA-02 | Find and log any match from every World Cup (men's and women's, ODI and T20), Champions Trophy, Asia Cup and IPL season since 1973, from an archive Innings builds itself with Wikipedia's tournament pages as the reference | `TestArchive` checks all 2,948: teams the app knows, a winner who played, a real date, no repeat of a Vault match, and five well-known games score for score; `TestArchiveMatches` finds India v Pakistan 2023 by "ind vs pak 2023", opens it, logs it and keeps it off the Matches tab; `TestSearchMatches` finds all 44 India–Australia games. Walked on an iPhone simulator: the 1996 Eden Gardens semi-final and the 2009 IPL final open with their day and no chat | `scripts/archive.py`, `internal/sim/archive.json`, `internal/sim/archive.go` |

### During the match (v1.2)

Added 2026-10-29, when the developer said the app is mainly about talking a match through and predicting how it goes.

| ID | The app must… | Check | Built in |
| --- | --- | --- | --- |
| CHAT-01 | Give every real match a chat room, open from a day before the first ball until a day after the match, with the score shown above it and new messages arriving within a few seconds | Rooms per match, oldest first, `after` reads, one message a second, 1–500 characters, closed rooms and classics refused, a deleted account's messages removed (`TestMatchChat`); walked on an iPhone simulator during South Africa v Australia, where a second fan's message appeared on its own | `app/(app)/match/chat/[id].tsx`, `features/match/ChatEntry.tsx`, `internal/httpapi/chat.go`, `internal/store/chat.go` |
| LIVE-02 | Let fans call how the match goes while it's on: where the side batting first will finish, whether the chase succeeds, and whether a batter reaches 50 or 100, each locking at a set point and settling on its own, 5 points each on the Bragging Board | A T20 followed reading by reading opens, locks and settles every kind; a stale reading never reopens a call; a call first seen past its lock starts locked (`TestLiveCallsFollowTheMatch`, `TestCallsFirstSeenPastTheirLockStartLocked`); picks change while open, are refused once locked, and score on the profile and board (`TestLiveCalls`); the batting card is read from a real ESPN summary (`TestBattingCardFromRealSummary`); walked on an iPhone simulator during Afghanistan v Zimbabwe | `features/match/LiveCalls.tsx`, `internal/feed/calls.go`, `internal/httpapi/calls.go` |
| LIVE-03 | Let a fan make their own call from the chat (a batter to make 30, 50 or 100, a team to reach a total, a team to win), refused once it's as good as decided, posted into the chat for everyone to answer, and settled like the app's calls | Refused for a batter within 10 of the target, a total past its lock, a winner deep into the chase or a finished match; one card per question; settled from the batting card and the result (`TestFanCallsAreFairAndSettle`); posted, joined and put in the chat once (`TestFanCalls`); walked on an iPhone simulator, calling Ryan Rickelton to make 50 during South Africa v Australia | `app/(app)/match/call/[id].tsx`, `internal/feed/calls.go`, `internal/httpapi/calls.go` |

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
| PLAT-03 | Give each fan a session the server can check on every request (v1.0 did this with guest accounts; v1.1 replaced them with phone sign-in, and v1.3 with email and password; see ACCT-01 to ACCT-07) | `TestBearerToken`, `TestUnauthorized` | `internal/auth` |
| PLAT-04 | Generate a deterministic cricket world (fixtures, XIs, live scores, friends, fans, crowd) with no API keys | The same anchor time gives the same world (`TestDeterminism`); live, upcoming and finished matches all exist (`TestStateCoverageAtAnchor`) | `internal/sim` |
| PLAT-05 | Keep one JSON contract in TypeScript, mirrored in Go, used through typed hooks | No endpoint returns `null` for a list (`TestNoNulls`, `TestEveryGET`) | `src/domain/`, `internal/domain/`, `src/data/index.ts` |
| PLAT-06 | Work on Android: icons, fonts, the API host and the system navigation bar | Home loads with data on the emulator; the emulator reaches the API at 10.0.2.2 | `src/data/client.ts` |
| PLAT-07 | Ship an app icon, an Android adaptive icon and a splash screen in the chosen look | A red ball on the concrete ground #ECEAE4 | `assets/`, `app.json` |
| PLAT-08 | Have a README and a GitHub repository | The README's run steps work from a fresh clone. The repository is public (it was private until 2026-11-16) | `README.md` |

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
| Email address | one address: a single @, a domain with a dot, no spaces, at most 254 characters | `bad-email` |
| Password | 8–128 characters; not the email address or a commonly used password | `weak-password` |
| Emailed code | 6 digits; 10 minutes; 5 tries | `bad-code`, `code-expired` |
| New code | 30 seconds apart; 5 an hour per address; 20 code emails an hour per IP | `too-soon`, `too-many` |
| Wrong passwords | 10 per account in 15 minutes; 50 per IP in an hour | `too-many` |
| Reset token | 15 minutes, once | `reset-expired` |
| Username | 3–20 characters: `a–z`, `0–9`, `_` and `.`, starting with a letter, unique | `bad-handle`, `handle-taken` |
| Contacts to match | up to 5,000 SHA-256 hashes | `bad-hashes` |
| Name | 1–40 characters | `bad-name` |
| Bio | up to 160 characters | `bad-bio` |
| Teams | 1–5 known teams, no repeats | `bad-team` |
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

- **Matches:** real, from ESPN's public scores feed, polled by the server and saved in the database. Simulated fans don't predict, rate or watch real matches. `INNINGS_MATCHES=sim` runs the deterministic simulation instead, where `INNINGS_ANCHOR` pins the clock so tests can assert exact numbers.
- **The Vault:** real matches only. An entry that can't be verified against a public scorecard is dropped.
- **Past tournaments:** Innings' own archive, `internal/sim/archive.json`, built by `scripts/archive.py` from Wikipedia's tournament pages (CC BY-SA 4.0). A tournament whose match count differs from its own page is left out, and so is a match whose winning margin doesn't fit its scores.
- **The contract:** lists and maps are never `null` in the JSON; nullable fields are explicit. `src/domain/types.ts` and `src/domain/api.ts` define it, and `internal/domain` mirrors it.

### Security and privacy

- **Signing in:** a fan signs up with an email address and a password and proves the address with a 6-digit code emailed to it; the account can't sign in until then. Codes come from `crypto/rand` and are stored and compared as hashes; each one works once, for 10 minutes, with 5 tries. Sign-in, resend and forgot password answer the same whether or not an address has an account, and an unknown address takes as long to refuse as a wrong password, so they can't be used to find out who's signed up. Sign-up is the exception: it has to say when a confirmed account already has the address.
- **Passwords:** 8 to 128 characters, checked against the email address and a list of the most-used passwords, and stored as PBKDF2-HMAC-SHA256 at 600,000 iterations with a random salt. Resetting or changing a password signs out every other device.
- **Sessions:** every `/v1` route except the seven sign-up and sign-in calls needs `Authorization: Bearer <token>`. Tokens are 32 random bytes, kept in the phone's keychain or keystore (expo-secure-store), and the server stores only their SHA-256 hash. Each phone has its own session; signing out ends only that one.
- **Codes are never handed out by accident:** the server returns a code only in dev mode, never when it has a mail server, and sends none when it has neither. Half an SMTP setup stops the server at startup.
- **Email addresses** appear only in the owner's own account details, never on a profile. The server log leaves them out, except dev mode's lines with the code, which are there so you can sign up without a mail account.
- **Contacts:** email addresses are trimmed, lower-cased and hashed on the phone, matched against the server's hashes and thrown away. Anyone who already knows an address can hash it and check, so this keeps the list out of casual view but isn't anonymity.
- **Server limits:**
  - Request bodies are capped at 1 MB.
  - The server times out slow request headers after 10 seconds.
  - Every write is validated (see the limits above).
- **Photos:** the profile picture stays on the device as a file path and is never uploaded. The log no longer asks for a photo.
- **Known limits, accepted for now:**
  - CORS allows any origin.
  - Sessions don't expire; a fan stays signed in until they sign out or delete their account.
  - Real email delivery hasn't been tried yet; the mailer is tested against a mail server on the same machine.
  - The per-IP and wrong-password limits live in memory, so a restart resets them.

### Accessibility

- The code sets accessibility roles in 91 places and labels in 73 (counted on 2026-11-16).
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

- `make api` runs with no setup on port 8088, with sign-up codes shown on screen in dev mode until `.env` holds a mail account's settings.
- `make app` starts Metro on port 8082.
- `make test` runs the Go tests and then the Jest tests; `npx tsc --noEmit` and `npx expo lint` check the app.

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
  | Red | #D23A2E | Live, stamps, serials and the Log ball, plus small text actions (See all, PREDICT, ghost buttons) |
  | Green | #2F6B3A | Only for wins |

- **Type:**
  - Big Shoulders Display for team codes and big numbers
  - Schibsted Grotesk for UI and reviews
  - IBM Plex Mono for ticket fine print

## Later (v2)

| ID | Requirement |
| --- | --- |
| SMS-01 | Replaced in phase 10: codes go by email instead (ACCT-01, ACCT-03), so no SMS provider is needed. It was real text messages, with an SMS provider account, a sender ID and, for Indian numbers, DLT registration |
| ACCT-09 | Sign in with Apple and Google |
| MEDIA-01 | Upload photos so friends can see them: the profile picture, and photos on stubs if the log asks for them again |
| STORE-01 | EAS builds and App Store / Play Store releases |
| LIVE-01 | Live mood reactions during a match |
| WRAP-01 | Season Wrapped, an end-of-season summary to share |
| CLUB-01 | Team fan clubs |
| AI-01 | AI summaries of fan reviews |

## Open items

- [ ] Test on a real iPhone and a real Android phone (each needs a new development build for sign-in)
- [ ] Put a Gmail app password in `.env` and try real email delivery before friends sign up (README → Real emails)
- [ ] Walk every screen in floodlights (the review captured four)
- [ ] Confirm the ticket stock colours for WPL, World Cup, BBL, SA20, The Hundred and Vault classics
- [x] Add a way to delete a log in the app: "Delete this stub" on your own stub (phase 10)
- [x] Unit tests for the Top 5 helper and the Log draft (phase 10)
- [x] Retake the sign-up screenshots in the README for email sign-in (2026-10-30)
