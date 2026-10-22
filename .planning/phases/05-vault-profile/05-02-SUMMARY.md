---
phase: 05-vault-profile
plan: 02
subsystem: ui
tags: [vault, profile, settings, share-card, react-native-view-shot, expo-sharing, image-picker]

requires:
  - phase: 02-api-simulated-world
    provides: vault, rankings, users, top5 and me routes; stats and badges
  - phase: 03-core-loop
    provides: MatchTicket, the Log flow and its Printed step
  - phase: 04-social-predictions
    provides: FeedCard, FollowButton
provides:
  - Vault tab, collection and rankings screens
  - Profile tab and other fans' pages from one ProfileView
  - Settings (fan card, theme, about)
  - The 9:16 story card and share flow
affects: [phase-06 re-skin, README screenshots]

tech-stack:
  added: [react-native-view-shot 5.1.0, expo-sharing]
  patterns: ["one view for 'me' and any user id", "off-screen render then capture for share images", "pin and unpin helpers that report errors per row"]

key-files:
  created: ["src/app/(tabs)/vault.tsx", "src/app/vault/collection/[id].tsx", src/app/vault/rankings.tsx, src/features/profile/ProfileView.tsx, src/features/vault/top5.ts, src/features/share/StoryCard.tsx, src/features/share/useShareStory.ts, src/app/settings.tsx]
  modified: [src/features/log/LogFlow.tsx]

requirements-completed: [VAULT-01, VAULT-02, VAULT-03, VAULT-04, VAULT-05, PROF-01, PROF-02, PROF-03, PROF-04, PROF-05, PROF-06, PROF-07, SOC-05, LOG-07]

completed: 2026-10-22
status: complete
---

# Phase 5 Plan 2: Vault and Profile Screens, Settings and Share Card Summary

**The Vault gives fans something to open on a blank day (On This Day, Top 5, five collections, rankings), the Profile turns logs and calls into a collection worth showing, and a saved stub can be shared as a story image.**

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented
- **Commits:** none yet (uncommitted)

## Accomplishments

- **Vault tab** (`src/app/(tabs)/vault.tsx`): "The Vault", "Every great game you were part of."; On This Day with the year, how many years ago and "I watched this — log it"; Your Top 5 as a numbered pager of large tickets with unpin (confirmed with an alert); one horizontal strip per collection with its blurb and "See all"; the top five of the all-time fan rankings with "Full rankings".
- **Collection and rankings screens** (`src/app/vault/`): collection rows with date, teams, headline, scoreline, venue, fan rating, "I watched this" (opens the Log flow on that match) and "Pin to Top 5" / "Pinned". Rankings: format chips and era chips (eras come from the API), rank in big type, match, year, average and count.
- **Top 5** (`src/features/vault/top5.ts`): keeps order, refuses a sixth pin with "Your Top 5 is full. Unpin one on the Vault tab first." beside that row.
- **Profile** (`src/features/profile/ProfileView.tsx`, used by the Profile tab and `/user/[id]`): kicker "Your scorebook" or "Fan profile"; header with photo or coloured initials, name, @handle, favourite team badge and bio (with an "Add a one-line fan bio…" prompt on my own), Friends and Edit on mine, Follow on others; Top 5 strip (read-only); stats row (Logged, Avg given, Accuracy, Points); Stub Collection two to a row, each opening the stub page; Where I've Watched with counts by place and "Stadiums visited" with the grounds; Predictor Rank with rank of circle, streak and badges drawn as rubber stamps (tap for the description); a "Later this season / Season Wrapped" placeholder on my own profile.
- **Settings** (`src/app/settings.tsx`): kicker "The dressing room"; "Your fan card" with photo (library, square crop), name and bio with counters, favourite team chips (ten international sides and four IPL franchises), avatar colour swatches and "Save changes" (only changed fields sent); Appearance (Paper, Floodlights, Auto, applied at once); About with the version and the API address.
- **Share card** (`src/features/share/`): a 9:16 card (1080x1920 on a 3x phone) with the wordmark, the stamped ticket, balls, where watched, a review line, your Player of the Match and an "innings.app" footer; mounted off-screen after Save, captured with `captureRef`, shared through the system sheet titled "Share to your story".

## Decisions Made

- Pinning lives in the Vault only; the Profile's Top 5 is read-only and its empty state links to the Vault.
- The card is rendered by React, not drawn on a canvas, so it reuses `MatchTicket` and the theme.

## Deviations from Plan

- Season Wrapped is a placeholder panel, not a feature (the plan lists it as later; v2, WRAP-01).
- Profile photos, like log photos, are device file URIs saved as strings by `PUT /v1/me`; they do not upload.

## Issues Encountered

None recorded.

## Next Phase Readiness

Every tab and secondary screen exists. Phase 6 re-skins them and checks each one in both themes.

## Self-Check: PASSED

`npx tsc --noEmit` exits 0; `TestTop5AndMe` and `TestEveryGET` pass (see `05-VALIDATION.md`).
