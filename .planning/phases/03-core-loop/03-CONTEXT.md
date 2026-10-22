# Phase 3: Core loop - Context

**Gathered:** 2026-10-22
**Status:** Complete

<domain>
## Phase Boundary

The loop a fan repeats every match day: find a match on the Matches tab, open its Match Page (which changes with the match state), and log it through the centre button as a ticket stub. Covers the plan's Week 2 ("Matches tab and the match page; the Log flow; ticket stub design").

Predictions, Fans' Player of the Match and prediction results on the Match Page belong to Phase 4. The share card belongs to Phase 5. Phase 3 leaves slots for them in `src/features/match/MatchPage.tsx`.

</domain>

<decisions>
## Implementation Decisions

### Matches tab
- **D-01:** State chips (Live, Upcoming, Recent) and scope chips (India, IPL, WPL, World Cup, BBL, SA20, The Hundred) combine; tapping a selected chip clears it. India sends `team=ind,ind-w` because it is a team, not a competition.
- **D-02:** The list is grouped by day ("Live now", "Today", "Tomorrow", "Yesterday", then dates). While a new filter loads, the previous list stays on screen dimmed instead of flashing a spinner.
- **D-03:** Every empty state says what is empty and offers "Clear filters" when a filter caused it.

### Match Page
- **D-04:** One route, `/match/[id]`, for fixtures and Vault classics. Layout by state: BEFORE (upcoming), DURING (live), AFTER (completed and every classic).
- **D-05:** The page refreshes itself every 20 seconds so live scores move without a pull.
- **D-06:** The plan's "quick-reaction buttons" during a match are marked optional for later. The DURING layout shows a disabled reactions row labelled as coming soon (v2, LIVE-01).

### Log flow
- **D-07:** A full-screen modal with four steps called overs ("Over 2 of 4"): Pick the match, Rate it, Rate the players, Where I watched it. Opening it from a Match Page or the Vault leaves out Pick the match; classics have no XI, so Rate the players is left out too.
- **D-08:** Only live or finished matches can be logged; an upcoming one shows "Not started yet".
- **D-09:** Re-logging a match pre-fills the flow from the old stub, and the save keeps the stub number (the API upserts on user and match).
- **D-10:** Untouched players count as skipped; the Next button reads "Skip players" when nobody is rated and no Player of the Match is crowned.
- **D-11:** Tagging a friend raises the head count to at least the number tagged.
- **D-12:** The photo comes from the photo library (`expo-image-picker`), cropped 4:3. It is stored as a device file URI (see STATE concerns).

### Ticket stub and motion
- **D-13:** The stub is the signature element: a torn ticket with a notched perforation and a tear-off side stub carrying the serial number. One component (`src/components/TicketStub.tsx`) serves the Printed screen, the Profile and the share card.
- **D-14:** Motion is small and specific: the ball spins when a rating changes, the ticket feeds out after Save, then the LOGGED stamp slams down and settles at an angle, with a haptic tap.

### Claude's Discretion
- Cricket-flavoured empty and loading copy ("Fetching the scorecard…", "No play today").

</decisions>

<canonical_refs>
## Canonical References

- `docs/PLAN.md`: Tab 2 (Matches, the Match Page) and Tab 3 (Log)
- `src/domain/api.ts`: `MatchesResponse`, `MatchPageResponse`, `LogSuggestionsResponse`, `FeedItem`
- `.planning/phases/03-core-loop/03-UI-SPEC.md`

</canonical_refs>

<code_context>
## Existing Code Insights

From Phase 1: `Screen`, `QueryState`, `Empty`, `Chip`, `Panel`, `SectionHeader`, `SplitBar`, `BallRating`, `CricketBall`. From Phase 2: `GET /v1/matches`, `GET /v1/matches/{id}`, `GET /v1/log/suggestions`, `GET /v1/search`, `POST /v1/logs`, `PUT /v1/matches/{id}/watching`.

</code_context>

<deferred>
## Deferred Ideas

- Live quick reactions (v2, LIVE-01)
- Deleting a stub from the app: the API supports it; no screen offers it yet
- Camera capture for the photo (library only for now)

</deferred>

---

*Phase: 3-Core loop*
*Context gathered: 2026-10-22*
