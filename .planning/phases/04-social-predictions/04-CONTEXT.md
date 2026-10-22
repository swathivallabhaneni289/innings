# Phase 4: Social & predictions - Context

**Gathered:** 2026-10-22
**Status:** Complete

<domain>
## Phase Boundary

The social layer and the prediction game on top of the core loop: the Home tab (On Now, Your Call, Friends' Feed, Just Finished, Bragging Board), predictions with the first-ball lock, Fans' Player of the Match and prediction results on the Match Page, likes, replies and "I was there too" on stubs, following friends, notifications and search. Covers the plan's Week 3 ("follow friends and the Home feed; the prediction system, which locks at the first ball, plus scoring; Fans' Player of the Match").

Scoring itself was built in Phase 2 (`internal/domain/scoring.go`); this phase surfaces it.

</domain>

<decisions>
## Implementation Decisions

### Predictions
- **D-01:** One component, `src/features/predict/PredictionPicker.tsx`, serves the Your Call card on Home and the Match Page. Three taps in order (winner, top run-scorer, top wicket-taker); the call saves itself after the third pick; tapping any pick changes it until the first ball.
- **D-02:** The lock is enforced twice: the app hides the picker once the start time passes (the clock re-renders every 1 to 5 seconds on these cards), and the API answers 409 `locked` on the server clock. A 409 switches the picker to its locked state instead of showing an error.
- **D-03:** Locked picks show read-only under "Locked" with a padlock glyph, on Home, on the Match Page BEFORE once the time passes, and as "Your picks" DURING.
- **D-04:** The Match Page shows two splits: friends' predictions (with who picked what) and all fans.

### Fans' Player of the Match
- **D-05:** The fans' pick is the highest average out of 10 with at least 3 ratings, shown beside the official Player of the Match; a "Fans disagree!" stamp appears when they differ.

### Home
- **D-06:** Your Call is the next open match in the next three days, ranked India first, then World Cup, then other internationals, then leagues.
- **D-07:** Friends' Feed shows the people you follow, newest first, 8 at a time with "More from the stands".
- **D-08:** The Bragging Board counts points from matches finished in the last 7 days, for you and everyone you follow.
- **D-09:** Home refetches every 30 seconds; pull to refresh shows a spinner only for the pull.

### Social
- **D-10:** Follows are one-way toggles; any simulated friend or public fan can be followed. A new guest already follows 5 of the 8 friends.
- **D-11:** Like and "I was there too" are optimistic toggles that settle on the server's counts. Replies are plain text on the stub page.
- **D-12:** Notifications are built on the server from current state (locks within 3 hours, friends' logs from the last 3 days, points scored, replies and likes on my stubs). The bell counts items newer than the last visit to the notifications screen, stored on the device.

### Claude's Discretion
- Search matches teams, venue, city, stage, competition and year, and fans by name or handle; it starts at two characters.

</decisions>

<canonical_refs>
## Canonical References

- `docs/PLAN.md`: Tab 1 (Home), Bragging Points System, Build decisions
- `src/domain/api.ts`: `HomeResponse`, `PredictionSplit`, `PredictionResultRow`, `FanSummary`, `FeedItem`, `NotificationsResponse`, `SearchResponse`
- `.planning/phases/04-social-predictions/04-UI-SPEC.md`

</canonical_refs>

<code_context>
## Existing Code Insights

`MatchPage.tsx` (Phase 3) has slots for the prediction panel, player ratings and prediction results. `FeedCard` is shared with the Match Page reviews and the Profile; its header marks it as a fixed contract owned by the Home workstream.

</code_context>

<deferred>
## Deferred Ideas

- Push notifications: the list is in-app only
- Other real guests: search and follow include them, but only guests on the same server exist to be found

</deferred>

---

*Phase: 4-Social & predictions*
*Context gathered: 2026-10-22*
