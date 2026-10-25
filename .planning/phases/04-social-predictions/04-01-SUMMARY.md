---
phase: 04-social-predictions
plan: 01
subsystem: ui
tags: [predictions, social, feed, notifications, search, react-query, optimistic-ui]

requires:
  - phase: 02-api-simulated-world
    provides: scoring, prediction lock (409), home, notifications, search, likes, there-too, replies, follows
  - phase: 03-core-loop
    provides: MatchPage slots, MatchTicket, BallRating usage
provides:
  - Home tab with five sections
  - PredictionPicker used on Home and the Match Page
  - Fans' Player of the Match, player ratings and prediction results on the Match Page
  - Stub page with replies; like and "I was there too" toggles
  - Friends, notifications and search screens
affects: [phase-05 Profile (FeedCard, FollowButton), phase-06 re-skin]

tech-stack:
  added: []
  patterns: ["optimistic toggles that settle on server counts and roll back on failure", "client clock plus server 409 for the lock", "unread count kept on the device, not the server"]

key-files:
  created: ["src/app/(tabs)/index.tsx", src/features/predict/PredictionPicker.tsx, src/features/feed/FeedCard.tsx, src/features/feed/StubActions.tsx, "src/app/stub/[id].tsx", src/app/notifications.tsx, src/app/search.tsx, src/app/friends.tsx]
  modified: [src/features/match/MatchPage.tsx]

requirements-completed: [HOME-01, HOME-02, HOME-03, HOME-04, HOME-05, HOME-06, PRED-01, PRED-02, PRED-07, MATCH-06, MATCH-07, SOC-01, SOC-02, SOC-03, SOC-04]

completed: 2026-10-22
status: complete
---

# Phase 4 Plan 1: Home, Predictions and the Social Layer Summary

**Home now answers "what's on, what's my call, what did my friends think", predictions lock at the first ball on both the client and the server, and stubs can be liked, replied to and claimed with "I was there too".**

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented
- **Commits:** `0622243` (committed together on 2026-10-22, after Phase 6)

## Accomplishments

- **Home** (`src/app/(tabs)/index.tsx`, `src/features/home/`): "INNINGS" wordmark with a small ball, search and a bell with a red unread badge capped at 9+, a date line under a double rule. On Now: live matches first, then the rest of today, each card with the live score or a countdown and a Predict button that becomes Locked. Your Call: the picker in compact form, a lock countdown and the friends' split ("Friends' call · 3 picked"). Friends' Feed: 8 stubs at a time with "More from the stands". Just Finished: the last 24 hours with balls, average and the top review. Bragging Board: rank, avatar, name and points for the last 7 days, with my row marked.
- **Predictions** (`src/features/predict/PredictionPicker.tsx`): winner, then top run-scorer, then top wicket-taker; saves after the third pick ("Call saved", "Tap a pick above to change it before the first ball."); a failed save offers "Try again"; a 409 flips it to the locked view.
- **Match Page additions** (`src/features/match/`): "Your call" panel, "Friends' predictions" split with names, "All fans" split. AFTER: Fans' Player of the Match beside the official one with a "Fans disagree!" stamp, every rated player with an average out of 10; prediction results for me and friends with ticks per pick, points, an "Upset +10" tag and the streak multiplier.
- **Feed and stub page** (`src/features/feed/`, `src/app/stub/[id].tsx`): `FeedCard` shows author, a compact ticket, balls, a three-line review, the fan Player of the Match and the place line; `StubActions` toggles like and "I was there too" optimistically. The stub page shows the full log (Fan POTM, Watched, With, player ratings), replies with relative times, and a reply box.
- **Friends, notifications, search**: follow toggles on every simulated friend; notifications from the server with relative times and links to the match or stub; search for matches and fans from two characters.

## Decisions Made

- The unread badge counts notifications newer than the last time the screen was opened (`notificationsSeenAt` in the device store); `HomeResponse.unread` from the server is not used for the badge.
- `FeedCard` and `people.tsx` are marked as fixed contracts in their headers because the Match Page and the Profile import them.

## Deviations from Plan

- `FollowButton` and `FriendRow` live in `src/features/profile/` because the Profile header and the Friends screen both use them.

## Issues Encountered

None recorded.

## Next Phase Readiness

Profile (Phase 5) can reuse `FeedCard`, `FollowButton` and the stats from `GET /v1/users/{id}`.

## Self-Check: PASSED

`npx tsc --noEmit` exits 0; prediction, like, follow and merge tests pass (see `04-VALIDATION.md`).
