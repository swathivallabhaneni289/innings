# Phase 5: Vault & Profile - Context

**Gathered:** 2026-10-22
**Status:** Complete

<domain>
## Phase Boundary

The two tabs that make the app worth opening on a day with no cricket: the Vault (a fact-checked archive of real classic matches, fan rankings, On This Day, Your Top 5) and the Profile (header, Top 5 stubs, stats, stub collection, where watched, predictor rank and badges), plus Settings and the shareable story card. Covers the plan's Week 4 ("seed the Vault with around 50 classic matches by hand; Profile stats and Top 5; a shareable story card").

Two workstreams ran in parallel: the Vault seed and its fact-check (05-01, data only) and the Vault and Profile screens (05-02).

</domain>

<decisions>
## Implementation Decisions

### Vault data
- **D-01:** Real historical matches only. Each entry's date, venue, teams, result and scoreline must be checked against a public source; an entry that cannot be verified is dropped rather than guessed.
- **D-02:** Three independent checker agents each verified every entry. Result: 58 confirmed, 0 corrected, 0 dropped. One source per match is recorded in `docs/VAULT_SOURCES.md`.
- **D-03:** Each seed carries a `hype` score from 1 to 5 that shapes its simulated fan ratings, so legendary matches rank high without hand-written numbers.
- **D-04:** The five collections come from the plan: World Cup Finals, Greatest Chases, Test Match Thrillers, Super Over Madness, Women's Cricket Classics. A match can sit in more than one.
- **D-05:** Classics have no playing XI in the data, so logging one skips Rate the players.

### Vault screens
- **D-06:** Rankings mix Vault classics and completed fixtures, ordered by average fan rating, filterable by format and era.
- **D-07:** On This Day prefers a classic played on today's date; failing that, the next anniversary coming up, with "years ago" shown either way.
- **D-08:** Top 5 is pinned from the collection screens ("Pin to Top 5"), unpinned from the Vault tab (with a confirm), and shown read-only on the Profile. Five is the hard cap, enforced in the app and by the API.

### Profile
- **D-09:** One view serves the Profile tab ("Your scorebook") and any fan's page ("Fan profile"); only my own shows Settings, Edit and Friends.
- **D-10:** Stats come from `GET /v1/users/{id}`: matches logged, average rating given, prediction accuracy (ties excluded), points, current streak, rank among the fan and the people they follow, badges, watch places and stadiums visited.
- **D-11:** Season Wrapped is in the plan as "later": the Profile shows a "Later this season" placeholder panel only (v2, WRAP-01).
- **D-12:** Settings edits the fan card (photo from the library, name up to 40 characters, a one-line bio (60 characters in the app; the API accepts up to 160), favourite team, avatar colour) and the theme (Paper, Floodlights, Auto).

### Share card
- **D-13:** The story card is a 9:16 view rendered off-screen and captured as a PNG with `react-native-view-shot`, then handed to the system share sheet with `expo-sharing`. It carries the wordmark, the stamped stub, the rating, where watched, a review line, the fan Player of the Match and an "innings.app" footer.

### Claude's Discretion
- The choice of the 58 matches across eras, formats and the women's game.

</decisions>

<canonical_refs>
## Canonical References

- `docs/PLAN.md`: Tab 4 (Vault), Tab 5 (Profile), Tips for Launch (the share card)
- `docs/VAULT_SOURCES.md`
- `internal/sim/classics.go` (`ClassicSeed` fields)

</canonical_refs>

<code_context>
## Existing Code Insights

`internal/sim/classics.go` embeds `classics.json` and is read-only for sim code; the seed file is owned by 05-01. `FeedCard`, `FollowButton`, `MatchTicket` and the Log flow's `Printed` step exist from Phases 3 and 4.

</code_context>

<deferred>
## Deferred Ideas

- Season Wrapped (v2, WRAP-01)
- Reordering the Top 5 by drag: pins append in order; unpin and re-pin to reorder

</deferred>

---

*Phase: 5-Vault & Profile*
*Context gathered: 2026-10-22*
