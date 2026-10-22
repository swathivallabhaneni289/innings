---
phase: 5
slug: vault-profile
status: as-built
created: 2026-10-22
revised: 2026-10-22
revision: 1
revision_reason: "Written after the build to record the screens as they shipped in the Phase 1 heritage scorebook look. Phase 6 re-skins them in direction C (Matchday Ticket), where Vault classics print on an old-card ticket stock."
---

# Phase 5 - UI Design Contract

## Design System

Phase 1 tokens and components. Large ticket stubs (`MatchTicket` size `lg`) carry the Top 5 and the share card; badges are drawn as red rubber stamps, slightly askew; stats use the mono stat role.

## Screens

### Vault tab (`src/app/(tabs)/vault.tsx`)

| Section | Content |
|---------|---------|
| Title | "The Vault" with the line "Every great game you were part of." |
| On this day | Year in big type, headline, teams, scoreline, "{n} years ago today" (or "{n} years ago on {date}"), button "I watched this — log it" |
| Your Top 5 | Pager of large numbered tickets with the crowd's average rating and the headline printed on each; tap opens the match; "Unpin" above each ticket, or long-press for "Unpin from Top 5?" |
| Collections (x5) | Section title, blurb in italic, horizontal strip of `ClassicCard` (year in big mono, teams, hook, fan rating), action "See all" |
| All-time fan rankings | Top five `RankRow`s, action "Full rankings" |

### Collection screen (`src/app/vault/collection/[id].tsx`)

Kicker "The Vault", title from the collection. Each `ClassicRow`: date, teams with badges, headline, scoreline, venue, fan rating, buttons "I watched this" and "Pin to Top 5" / "Pinned", and any pin error inline under that row.

### Fan Rankings (`src/app/vault/rankings.tsx`)

Kicker "The Vault", title "Fan Rankings". Format chips (All and each format) and era chips (All plus the eras from the API). Rows: rank in big stadium type, match, year, headline, balls, average to two decimals and "({n} fans)".

### Profile (`src/app/(tabs)/profile.tsx` and `src/app/user/[id].tsx`)

| Section | Content |
|---------|---------|
| Bar | Kicker "Your scorebook" (mine, with a settings gear) or "Fan profile" (others, with Back) |
| Header | Avatar (photo, or initials on the chosen colour), name, @handle, favourite team badge, bio in italic quotes; mine: "Friends · {n} following" and "Edit profile"; others: Follow / Following |
| Top 5 Matches | Snapping strip of large numbered tickets with the owner's own rating where they logged it, read-only |
| Stats row | Logged · Avg given · Accuracy · Points, between rules |
| Stub Collection · {n} | Small vertical stubs two to a row, newest first; each opens `/stub/[id]` |
| Where I've Watched | One inked bar per place with its count; "Stadiums visited: {n}" and the grounds |
| Predictor Rank | "#{rank}" in red, "of {n} among friends" (or "in {Name}'s circle"), bragging points, current streak, badges as stamps (tap shows what each means) |
| Season Wrapped | Mine only: "Later this season", "Season Wrapped", "Your end-of-season card arrives when the season ends." |

### Settings (`src/app/settings.tsx`)

Kicker "The dressing room", title "Settings". "Your fan card": avatar preview, "Choose photo" / "Change photo" / "Remove photo"; Name with a counter (40); "Fan bio · one line" with a counter (60), placeholder "Dhoni finishes, I cry"; "Favourite team" chips; "Avatar colour" swatches (Ball red, Ink navy, Slate, Leather, Heather); "Save changes". "Appearance": Paper, Floodlights, Auto. "About": "Innings — The Fan's Cricket Diary", version, API address.

### Share card (`src/features/share/StoryCard.tsx`)

9:16, captured at device pixel ratio (1080x1920 on a 3x phone): wordmark, the stamped ticket, balls, where watched, a line of the review, "Your Player of the Match", footer "innings.app".

## Copywriting Contract

| Element | Copy |
|---------|------|
| Vault, no On This Day | "No play today" / "Nothing in the archive for this date. The scorebooks are resting." |
| Vault, no Top 5 | "No Top 5 yet" / "Open a collection below and tap “Pin to Top 5” on the games you'd take to your grave. Five tickets, your call." |
| Top 5 full | "Your Top 5 is full. Unpin one on the Vault tab first." |
| Rankings empty | "No play today" / "No rated classics for this format and era. Try a wider net." |
| Profile, no Top 5 (mine) | "No Top 5 yet" / "Pin your greatest matches from the Vault" / "Open the Vault" |
| Profile, no Top 5 (others) | "No Top 5 pinned" / "{Name} hasn't picked their all-time greats yet." |
| Profile, no stubs (mine) | "No stubs yet" / "Log a match and its ticket gets printed here." / "Log a match" |
| Share unavailable | "Sharing isn't available on this device." |

## Interaction Contracts

- Pins save immediately; the button shows a busy state on that row only.
- Settings sends only the fields that changed; Save is disabled until something changes and the name is valid.
- Appearance applies without a restart and is stored on the device.

## Accessibility

- The settings gear is labelled "Settings"; badges announce their title and open their description on tap.

## Checker Sign-Off

- [x] Every screen in this phase is described as built
- [ ] Visual review in both themes: Phase 6 screenshot QA
