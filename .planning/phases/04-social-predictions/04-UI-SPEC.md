---
phase: 4
slug: social-predictions
status: as-built
created: 2026-10-22
revised: 2026-10-22
revision: 1
revision_reason: "Written after the build to record the screens as they shipped in the Phase 1 heritage scorebook look. Phase 6 re-skins them in direction C (Matchday Ticket); in C, a feed stub's rating also appears as a red rubber-stamp circle."
---

# Phase 4 - UI Design Contract

## Design System

Phase 1 tokens and components, unchanged. Red marks primary actions, live state, the unread badge, my own row and stamps; green marks correct picks, saved calls and points earned; nothing else is coloured.

## Screens

### Home (`src/app/(tabs)/index.tsx`)

| Section | Content |
|---------|---------|
| Header | Small ball and "INNINGS" wordmark left; search and bell right, bell with a red count badge (9+ cap); double rule; date line |
| On Now | Horizontal strip of match cards: competition and format, both teams, live score with a LIVE tag or the start countdown; Predict button, or "Locked" with a padlock glyph once the first ball is bowled |
| Your Call | One large card: competition, format and stage, "TEAM v TEAM", lock countdown ("Locks in 2h 14m"), the picker in compact form, the friends' split bar ("Friends' call · 3 picked" or "nobody's called it yet") |
| Friends' Feed | Section action "Friends"; `FeedCard` list, 8 at a time, then "More from the stands · {n}" |
| Just Finished | Matches ended in the last 24 hours: kicker, teams, balls with the average and count, top review |
| Bragging Board | Up to the top rows plus my row if lower: two-digit rank (leader in red), avatar, name ("You" for me, highlighted), points |

### Prediction picker (Home and Match Page)

| State | Shows |
|-------|-------|
| Picking | Step tabs (Winner, Most runs, Most wickets) with a question per step ("Who wins it?", "Who makes the most runs?", "Who takes the most wickets?"); two team buttons; player chips from both XIs |
| Saved | Green tick, "Call saved" or "Your call is in", "Tap a pick above to change it before the first ball." |
| Error | The server message and "Try again" |
| Locked | "Locked" with a padlock glyph; the three picks read-only; "Predictions lock at the first ball." on the Match Page |

### Match Page additions

- BEFORE: "Your call" (picker), "Friends' predictions" (split bar plus who picked what), "All fans · {n} calls".
- DURING: "Your picks" read-only.
- AFTER: "Fan player ratings": two panels, official Player of the Match and Fans' Player of the Match, with a "Fans disagree!" stamp when they differ, then every rated player with a ball icon and an average out of 10. "Prediction results": me first, then friends; each row has avatar, name, "You" tag, "Upset +10" tag, three picks with ticks, points in green and the streak multiplier.

### Stub page (`src/app/stub/[id].tsx`)

Author row; the match ticket; balls and rating; review in italic serif; facts (Fan POTM, Watched, With); player ratings; replies with relative time; reply field ("Reply to {first name}…") and Post.

### Friends, Notifications, Search

- Friends: "Your circle" kicker, list of fans with avatar, name, bio, favourite team and a Follow or Following button.
- Notifications: "From the scorer's box" kicker; rows with text and relative time, tapping opens the match or the stub.
- Search: field "Teams, grounds, cities, years…"; results under Matches and Fans.

## Copywriting Contract

| Element | Copy |
|---------|------|
| On Now empty | "No play today" / "Nothing live or on today's fixture list. The Vault is always open." |
| Feed empty | "Quiet in the stands" / "When the friends you follow log a match, their stubs land here." / "Find friends" |
| Just Finished empty | "Nothing at the close yet" / "No matches have finished in the last 24 hours." |
| Board empty | "No points on the board" / "Make a call before the first ball. Points land here once matches finish." |
| Results empty | "No calls on the books" / "Nobody in your circle predicted this one." |
| Player ratings empty | "No player ratings yet" / "Rate the XIs when you log it and the table fills up here." |
| Notifications empty | "No new deliveries" |
| Search empty | "Nothing in the scorebook" |
| Friends empty | "Empty pavilion" |

## Interaction Contracts

- Home refetches every 30 seconds; notifications every 60; the Match Page every 20.
- Like and "I was there too" update at once and settle on the server's numbers; a failure restores the previous state.
- Opening Notifications marks everything seen on this device.

## Accessibility

- Bell: "Notifications, {n} new"; search: "Search"; like: "Like" or "Unlike"; "I was there too" exposes its selected state.

## Checker Sign-Off

- [x] Every screen in this phase is described as built
- [ ] Visual review in both themes: Phase 6 screenshot QA
