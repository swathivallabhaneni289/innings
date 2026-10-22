---
phase: 3
slug: core-loop
status: as-built
created: 2026-10-22
revised: 2026-10-22
revision: 1
revision_reason: "Written after the build to record the screens as they shipped in the Phase 1 heritage scorebook look. Phase 6 re-skins them in direction C (Matchday Ticket) without changing their structure; copy changes made there are recorded in Phase 6."
---

# Phase 3 - UI Design Contract

## Design System

Tokens, fonts and components come from Phase 1 (`src/theme/tokens.ts`, `src/components/`). As built: paper ground with grain and ruled lines, ink navy text, cricket-ball red for primary actions and stamps, pitch green only for wins; Oswald for headings and scores, Source Serif 4 for reviews, IBM Plex Mono for stats. Sharp corners on paper and tickets; only balls and avatars are round. The floodlights palette applies when the system is dark or the fan forces it.

## Screens

### Matches tab (`src/app/(tabs)/matches.tsx`)

| Region | Content |
|--------|---------|
| Header | "Matches" (h1), then a row of state chips (Live, Upcoming, Recent) and a row of scope chips (India, IPL, WPL, World Cup, BBL, SA20, The Hundred), then a rule |
| List | Day groups ("Live now", "Today", "Tomorrow", "Yesterday", dated); each `MatchCard` shows "IPL · T20 · Match 21", both teams with badges, the venue, then the live score with a LIVE tag, the first-ball time, or the result with balls, average and rating count |
| Loading | Previous list stays, dimmed to 45%, while a new filter loads; first load reads "Reading the fixture list…" |

### Match Page (`src/app/match/[id].tsx`)

| State | Sections, top to bottom |
|-------|-------------------------|
| Header (all) | Competition, format and stage; both teams; venue and start time. Classics add "From the Vault · {era}" and the headline |
| BEFORE | Prediction panel (Phase 4); "Where you're watching" with "I'm watching" opening the seven places inline, then Change and Clear; "Friends tuning in" |
| DURING | Live scoreboard (batting side marked, status line such as "IND need 34 off 21 balls"); reactions row shown faded as "Reactions — coming soon"; "Friends watching now"; "Your picks" read-only |
| AFTER | Result card; "Log this match" or "Edit your log"; Fan rating (big average, five balls, ten half-ball bars labelled 1 to 5); Fan player ratings and Prediction results (Phase 4); Reviews with Popular and Friends tabs; Where fans watched as a share line ("62% at home · 21% with friends") over a stacked bar with a key |

### Log flow (`src/app/log/index.tsx`, full-screen modal from the bottom)

| Over | Title | Content |
|------|-------|---------|
| Header | "Over 2 of 4" kicker, step title, an over strip with one cell per step | Close with a discard prompt ("Tear up this log?") once the draft is touched |
| 1 | Pick the match | Search field ("Older match or a classic: “2011 final”, “Wankhede”"), "Live now / just ended", "Recent" |
| 2 | Rate it | Match line; five large balls with half steps and a spinning ball; "Instant classic" check row |
| 3 | Rate the players | Both XIs; per player a crown (your Player of the Match), a 1 to 10 slider and Skip; hint "Untouched players are skipped. Slide to 0 to skip again." |
| 4 | Where I watched it | "Where were you?" seven place chips; "Where exactly?" note; head count stepper ("Just me", "With 6 people"); "Who was with you?" friend tags; "Photo & review" with "Add a photo" and "Your review (optional). What will you remember?" |
| Footer | Back, then Next, "Skip players" or "Save & print" | |
| Printed | Kicker "Close of play"; the ticket feeds out; LOGGED stamp; your Player of the Match; the review line; "Share story"; Done | |

## Copywriting Contract

| Element | Copy |
|---------|------|
| Matches empty, live | "No play right now" / "Nothing live{ in the IPL} at the moment. The covers are on." |
| Matches empty, upcoming | "Nothing on the fixture list" |
| Matches empty, recent | "No recent results" |
| Fan rating empty | "No ratings yet" / "Nobody has rated this one. Log it and set the tone." |
| Reviews empty | "No reviews yet" / "The scorebook is blank. Log it and write the first one." |
| Friends' reviews empty | "Quiet in the dressing room" |
| Log, nothing to suggest | "No play today" / "Nothing live or freshly finished. Pick a recent one below, or search the Vault." |
| Log, upcoming match | "Not started yet" / "You can log this one once the first ball is bowled." |
| Log, no friends to tag | "Batting alone" / "Follow some fans and you can tag them here." |
| Log, Next without a rating | "Give it at least half a ball." |
| Log, Save without a place | "Pick where you watched it." |
| Log, discard | "Tear up this log?" / "Your rating and notes won't be saved." |

Watch-place chips print an emoji before each label (from `WATCH_PLACES` in `src/domain/constants.ts`); these are content, matching the plan.

## Interaction Contracts

- Chips toggle: tapping the selected chip clears it.
- The Match Page refetches every 20 seconds and on pull to refresh.
- Haptics: selection tick on rating, crown and place; warning when Next is blocked; success when the stamp lands.
- Rating, crown and place errors show inline under the step, not as alerts.

## Accessibility

- Over strip cells are labelled "Over {n}: {title}"; the centre tab is "Log a match".
- Crown and Skip buttons name the player ("Make Virat Kohli your Player of the Match", "Skip Virat Kohli").
- "Reactions coming soon" is announced as one element.

## Checker Sign-Off

- [x] Every screen in this phase is described as built
- [ ] Visual review in both themes: Phase 6 screenshot QA
