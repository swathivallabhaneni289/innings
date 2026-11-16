# Innings: Product Requirements Document

**Owner:** Swathi Vallabhaneni · **Status:** Built and tested, not launched · **Last updated:** 2026-11-16

## 1. Overview

Cricket fans have apps for scores and apps for scoring their own games, but nothing records how they felt about a match: who they backed before the toss, how good it was, which player they'd crown, and where they were when it finished.

Innings is a matchday diary for cricket fans on iPhone and Android that treats every match the way Letterboxd treats a film on opening night. Fans predict before the first ball and follow the match with a chat room and calls. Afterwards they rate it in cricket balls, rate the players out of 10, pick a Fans' Player of the Match, say where they watched and get a ticket stub. Friends, a Vault of classics, a Top 5, an archive of past tournaments and a profile round it out.

Fans already have a scores app, so Innings doesn't compete on scores. A log has to be quicker than a group-chat message, and the stub good enough to post.

I made Innings by prompting Claude Code. The idea, the plan and the product calls are mine. See [the prompts that shaped the build](PROMPTS.md) and the original [product plan](PLAN.md).

## 2. Goals and non-goals

Goals:

- Make the whole loop work for a fan and their friends, from the first call to a collection of stubs.
- Log a match in under 60 seconds, the plan's target.
- Stay worth opening on days with no cricket.
- Use real matches and cost nothing to run.
- Get every fact right.

Non-goals: competing on scores, scoring your own games, and domestic, A and youth cricket.

## 3. Target users

These personas are assumptions drawn from the plan, not research findings.

**The group watcher** sees most matches with other people and argues about them afterwards. They want a quick log and a stub worth posting.

**The predictor** calls every big match and wants proof when they're right, so they watch the Bragging Board and chase streaks and upsets.

**The memory keeper** remembers where they were for the great games ("2011 final: at my uncle's place in Hyderabad") and logs them between series to build a Top 5.

## 4. The experience

**Before the match**, the fan calls the winner, the top run-scorer and the top wicket-taker, on the match's page or on Home, until the first ball locks the picks.

**During the match**, they follow the live score, see which friends are watching, talk in the match's chat room and answer calls on how it's going.

**After the match**, they read the fans' verdict, rate the match with a tap or log it in four steps, and get a ticket stub to share.

**Between series**, they browse the Vault's 58 classics in five collections, or search 2,948 past tournament matches for the games they remember.

## 5. Requirements by area

All 80 requirements are built: 79 over ten phases, and DATA-02, added later as a quick task. [REQUIREMENTS.md](REQUIREMENTS.md) has each one's acceptance check.

**Accounts and onboarding** (ACCT-01 to ACCT-08, ACCT-10 to ACCT-12, PLAT-03): sign-up with an email and a password, confirmed by an emailed 6-digit code; a fan card with a unique @username and up to five teams; friends from contacts; and password reset by email.

**Home** (HOME-01 to HOME-06): what's on now, Your Call, friends' stubs, matches finished in the last 24 hours, and a 7-day Bragging Board.

**Matches and the Match Page** (MATCH-01 to MATCH-10, DATA-01): real matches by day, filtered by state and competition, on a page that changes before, during and after play.

**Predictions and calls** (PRED-01 to PRED-07, LIVE-02, LIVE-03): picks lock at the first ball and earn bragging points, with a bonus for calling an upset and a multiplier for a streak of correct winners (the exact rules are in [REQUIREMENTS.md](REQUIREMENTS.md#bragging-points)). A correct call during play earns 5 points, and a fan can post five calls of their own per match.

**Chat** (CHAT-01): a room for every real match, open from a day before the first ball until a day after.

**Logging and stubs** (LOG-01 to LOG-08): four steps from the red ball, a stub that shares as a 9:16 story image, and the same stub number when a match is logged again.

**Ratings and reviews** (LOG-03, LOG-04, MATCH-05, MATCH-06, MATCH-08): match ratings from 0.5 to 5 in half-ball steps and player ratings from 1 to 10. The Fans' Player of the Match needs at least 3 ratings, with "Fans disagree!" when it isn't the official pick. Reviews are listed as popular and friends'.

**Social** (SOC-01 to SOC-03, SOC-05): follows, likes, replies, "I was there too", notifications and fan profiles.

**The Vault** (VAULT-01 to VAULT-04, VAULT-06): five collections, rankings by format and era, On This Day, and "I watched this" logging.

**Past matches and search** (DATA-02, SOC-04): search across teams, grounds, cities, stages, competitions, years and fans, including archived matches, which stay off the Matches tab and Home.

**Top 5** (VAULT-05, PROF-02): up to five matches, pinned from any match page, reordered in an editor and shown on the profile as large stubs.

**Profile** (PROF-01, PROF-03 to PROF-06): team badges, stats, every stub, where the fan watched, and their rank among friends.

**Settings and privacy** (PROF-07, DESIGN-01, ACCT-06, ACCT-11): edit the fan card, choose the theme, change the password, sign out, or delete the account.

**Design** (DESIGN-01 to DESIGN-07, PLAT-07): the Matchday Ticket look prints every match as a ticket on its competition's card stock, from the screens to the app icon.

## 6. Non-functional requirements

**Both platforms from one codebase** (PLAT-01, PLAT-06). One Expo codebase in strict TypeScript runs on iPhone and Android as a development build.

**Free to run.** This is a firm constraint: no paid services, no API keys and no provider sign-ups. ESPN's feed needs no key, SQLite needs no setup, and emailed codes cost nothing.

**Privacy and security.** An emailed code proves the address before an account opens. Passwords run from 8 to 128 characters, can't be the email address or a common password, and are stored as PBKDF2-SHA256 at 600,000 iterations. The server keeps only hashes of session tokens. Contacts are matched on SHA-256 hashes of lower-cased email addresses and never stored, which keeps them from casual view but isn't anonymity. No phone numbers are collected, and email addresses never appear on profiles.

**Data accuracy.** Each Vault match was checked three times against public scorecards, with sources in [VAULT_SOURCES.md](VAULT_SOURCES.md). The archive script checks match counts, winning margins and grounds.

**Reliability** (PLAT-05). The JSON contract is written in TypeScript and mirrored in Go, lists are never `null`, and tests check every endpoint for it. Request handlers never call ESPN; one poller saves what it learns in the database.

**Accessibility and performance.** Buttons and controls carry accessibility roles and labels, filter chips are at least 40 pt tall, and ratings and split bars have spoken labels. A full VoiceOver and TalkBack pass is still to do, and load times haven't been measured.

## 7. Data sources

**ESPN's public scores feed** gives fixtures, live scores, squads, playing XIs and results for internationals, World Cups, the IPL, WPL, BBL and WBBL, SA20 and The Hundred. Only the server reads it, every minute while a match is on and every five minutes otherwise. It is unofficial, so ESPN can change it without notice.

**Innings' own archive** holds 2,948 matches from every World Cup (men's and women's, ODI and T20), Champions Trophy, Asia Cup and IPL season since 1973. A script builds it from Wikipedia's tournament pages, which are licensed CC BY-SA 4.0. Four games where Wikipedia's margin and scores disagree are left out, as are four walkovers.

**The simulation** (`INNINGS_MATCHES=sim`, PLAT-04) runs a deterministic cricket world in place of the feed, for tests and for running offline. With the feed on, it still supplies the simulated fans and their Vault reviews.

## 8. Technical approach

The app is React Native with Expo SDK 57, Expo Router and TypeScript. It talks JSON to a Go 1.26 API built on the standard library (PLAT-02), which stores data in SQLite through the pure-Go `modernc.org/sqlite`, with SQL kept portable to Postgres. The [README](../README.md) has the details, and [API.md](API.md) lists the endpoints.

## 9. Status and release history

Innings is built and tested, but it has not launched and has no real users. It has run on an iPhone 17 Pro simulator and a Pixel 8 emulator, not yet on a physical phone. The Go tests, typecheck, lint and 69 Jest tests pass, and an 88-check smoke run passed on 2026-11-16. The friends, fans and crowd ratings in the app and its screenshots are simulated. Next come an Android pass on the newest match-page panels, a first real sign-up email, and tests on real phones.

- **v1.0, 2026-10-22** (Phases 1 to 6, 64 requirements): the full loop on a simulated cricket world, with guest accounts.
- **v1.1, 2026-10-25** (Phase 7, 8 requirements): sign-in by phone number and texted code, fan cards and finding friends.
- **v1.2, 2026-10-29** (Phases 8 and 9, 4 requirements): real matches from ESPN, match chat and calls during play.
- **v1.3, 2026-10-30** (Phase 10, 3 requirements): email and password replaced the phone number, since texts cost money and, without them, anyone could pose as another fan. It also added forgot and change password and up to five teams.
- **2026-11-13**: the past-matches archive (DATA-02), then the match page's tap-to-rate, Top 5 pin and review box, and a Top 5 editor.
- **2026-11-16**: the GitHub repository went public (PLAT-08).

## 10. Success metrics

Innings has no users yet, so these are targets for the first big series after launch, not results.

1. **A seed group.** The plan's target is 20 people logging every match. Measured by logs per match in the series.
2. **Fast logging.** Under 60 seconds from the red ball to the printed stub. It has never been timed, so I'd time the first fans.
3. **Calls first.** I'd count predicting as a habit if the group calls most of the matches it later logs, measured by matching predictions to logs.
4. **Opening-night buzz.** Each match the group watches gets a top review and a Fans' Player of the Match within the 24 hours it spends in Just Finished, measured from log times.
5. **Between series.** Fans log Vault and archive matches on days with no cricket, measured from those logs.
6. **Shared stubs.** Fans post stubs as stories unprompted. Shares aren't counted yet, so this needs a counter.

Guardrails: running costs stay at nothing, and live scores refresh within a minute or two.

## 11. Out of scope and future work

Store releases (STORE-01) need Apple and Google developer accounts. Photo uploads (MEDIA-01) need server storage, so the profile picture stays on the phone. Also deferred: Sign in with Apple or Google (ACCT-09), live mood reactions (LIVE-01), Season Wrapped (WRAP-01), team fan clubs (CLUB-01), AI summaries of fan reviews (AI-01) and a hosted database. Today the data lives in SQLite on one machine, and phones reach the API over the same Wi-Fi.

Decided against: Supabase, replaced by the Go API and SQLite; phone numbers and texted codes (SMS-01), replaced by email and password because every text costs money and nothing proved who owned a number; and launch tactics such as timing and seeding, which aren't build tasks.

## 12. Risks and mitigations

**ESPN's feed changes.** It's unofficial, and it suits a personal app, not a commercial one. The code reads it defensively, the poller keeps what it learns in the database, and its tests run on saved responses that get refreshed when the format changes.

**A wrong fact.** Classic matches are the app's credibility on days with no cricket. The archive is built by a script rather than checked like the Vault, so tests check all 2,948 matches, and games whose margin and scores disagree are left out.

**An empty app at launch.** Simulated fans don't predict, rate or watch real matches, so those sections start empty. The plan's answer is to launch before a big series with a seeded group of friends.

**Email delivery.** Real delivery hasn't been tried; the mailer was only tested against a mail server on the same machine. A real sign-up through Gmail comes before friends join.

**Security limits.** Sessions don't expire, CORS allows any origin, and some rate limits reset when the server restarts. They're documented and accepted for now.

## 13. Open questions

- Where will the API and database run for real fans, and can that stay free?
- Is an ordinary Gmail account enough to send sign-up codes at launch?
- What replaces ESPN's feed if Innings ever goes commercial?
- Should the simulated fans stay once real fans arrive?
- Should sessions expire, and CORS narrow, before launch?
- How should the app credit Wikipedia under CC BY-SA 4.0?
