# Phase 6: UI Review ("Matchday Ticket")

**Audited:** 2026-10-22
**Baseline:** the polish-common.md rules, the C-Home/C-Match mockups, and the Design Direction in docs/PLAN.md
**Screenshots:** 22 iPhone 17 Pro captures in scratchpad/qa/shots2/, each showing the top ~844pt of its screen. Content below the fold was judged from the source code.

---

## Pillar Scores

| Pillar | Score | Key finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | The voice is warm and cricket-literate. One placeholder ("coming soon") ships, and step counts disagree between log steps |
| 2. Visuals | 3/4 | The ticket metaphor is applied consistently. Duplicated headers and a wrong "On now" strip break the focal point on Home |
| 3. Color | 2/4 | Red is spread across team split bars, chart peaks, format codes and "Edit profile", so it no longer means "live/stamp" |
| 4. Typography | 3/4 | The three type roles are mostly correct. Mono leaks into prose such as "Locks in 19:10:54" and "fans" counts, and one rank number renders at about 4px |
| 5. Spacing | 3/4 | 20px gutters hold. The tab bar and Log ball cover the last content on every tab screen, and ticket text truncates |
| 6. Experience Design | 2/4 | Home "On now" opens scrolled and puts a finished match first. A refetch spinner sits over loaded rankings. Empty states are clipped |

**Overall: 16/24**

Verdict: most of the app reads as hand-made. Stocks, perforations, serials, the barcode stubs and the "From the Vault" stamp are convincing. The "AI-made" tells that remain are systemic, not per-screen: red used as a generic second colour, full-width black buttons on every card, centred empty-state boxes with the same three-line shape, and the same team-colour stripe used as a left accent bar in the scoreboards.

---

## Prioritized fix list

**P0: broken (fix before release)**

1. **Home, "On now" strip opens scrolled and shows the wrong match first.** In home.png a lavender ticket is cut off at x=0. The first card shown is a *finished* KKR v SRH marked "RESULT", while IND v AUS is live (see matches.png). This contradicts the comment at OnNowStrip.tsx:17 ("live matches first").
   - File: `src/features/home/OnNowStrip.tsx`. Also check the caller in `src/app/(tabs)/index.tsx`.
   - Change: sort `matches` with live first, then upcoming, then completed, before mapping. Add `contentOffset={{x:0,y:0}}` and `snapToAlignment="start"` to the ScrollView. Drop completed matches from "On now" (they belong in JustFinished). Replace `snapToInterval={ticketW + 12}` with a matching `gap: 12` in `stripContent` so the snap points line up.

2. **Fan rankings: rank "2" renders at about 4px.** vault-rankings.png at y≈1355 shows a speck instead of the number. The cause is `adjustsFontSizeToFit` with `width: 48` on a Big Shoulders glyph.
   - File: `src/features/vault/RankRow.tsx:29,69`
   - Change: remove `adjustsFontSizeToFit`. Set the rank style to `{ width: 56, fontSize: 44, lineHeight: 48 }` and add `minWidth: 56`.

3. **Fan rankings: a refresh spinner sits over loaded content.** The red spinner at y≈315 sits above the intro copy while the list is already rendered.
   - File: `src/app/vault/rankings.tsx:25`
   - Change: pass `refreshing={rankings.isRefetching && !rankings.isFetchedAfterMount}`, or use a local `pulling` state that is set only in `onRefresh`. Filter changes should not show the pull-to-refresh spinner.

4. **The tab bar and Log ball cover the end of content.** In vault.png and dark-vault.png the empty-state sentence is cut by the ball ("Fi\ \kets"). In home.png the "Friends' call" card is cut. In profile.png the stub grid is cut.
   - File: `src/components/Screen.tsx:68` (`insets.bottom + 110`)
   - Change: the bar plus the ball overhang is about 128pt above the safe area. Use `insets.bottom + 140` for tab screens.

5. **The fan-rating histogram's "BALLS" axis label is clipped.** In match-done.png and dark-match-done.png it shows at y≈1995 as half-cut "BALLS".
   - File: `src/features/match/FanRating.tsx:70-73`
   - Change: set `chart.height` to `BAR_MAX + 18 + 16`, or move "balls" into the "Out of 5…" line and delete the axis caption.

**P1: off-direction (the "AI-made" tells)**

6. **Red used as a team colour in the friends' split.** In home.png (y≈1712) and match-upcoming.png the second team's bar is overprint red. That reads as "losing/alert" and spends the screen's one accent.
   - File: `src/components/ui.tsx:177` (SplitBar). Fixing it here applies everywhere.
   - Change: render the left bar `t.ink` and the right bar `t.inkSoft`/`t.rule` at 50% (`t.ink` at opacity 0.35). Add the team-colour 3px underline under each code label, as TeamBadge already does.

7. **Red in the chart and on non-serial labels.** Examples: the histogram peak bar (FanRating.tsx:45), "T20" in red on the Home stub (OnNowStrip.tsx:140, `tone="red"`), and "Edit profile" in red (profile.png). Brief rule 4 limits red to live, stamps, serials and the Log ball.
   - Files and changes:
     - FanRating: `i === peak ? t.ink : t.inkSoft` (peak in ink, others soft). Do not use red.
     - OnNowStrip: change `{m.format}` to `tone="soft"`. Put the serial `No. ####` in red there instead, as MatchCard does.
     - ProfileHeader: make "Edit profile" an outline Button (`kind="outline"`) next to "5 following", not a red text link.

8. **Live/upcoming text truncates.** Examples: "Royal Challengers Beng…" (match-upcoming.png y≈1586), "Trent Br…" (stub.png y≈824), "ICC MEN'S T20…" (vault.png y≈518), and "Royal Challengers…" / "Chennai Super Kin…" (friends.png).
   - Files and changes:
     - `src/features/predict/PredictionPicker.tsx:218`: use `numberOfLines={2}` with `textAlign:'center'`, or show `team.short` plus the city.
     - `src/app/stub/[id].tsx` ticket footer: move the venue to its own line under the date.
     - `src/features/vault/OnThisDay.tsx:28`: use `numberOfLines={2}`, or drop "ADMIT ONE ·" here.
     - `src/features/profile/FriendRow.tsx`: show the team code only and drop the full name.

9. **Mono used in reading text.**
   - "Locks in 19:10:54" on the Home "Your call" card (home.png) mixes Schibsted and mono in one sentence.
   - "3,004 fans" (vault-rankings.png, vault-collection.png) is in mono.
   - "1 match" section counts (matches.png) are in mono.
   - Files: `src/features/home/YourCallCard.tsx`, `src/features/vault/RankRow.tsx`, `src/features/vault/ClassicCard.tsx`, `src/features/matches/MatchList.tsx`
   - Change: render the counts in Schibsted `bodySmall` with `tone="soft"` (`fonts.ui`). Keep mono only on the countdown digits, set as a separate right-aligned figure: "Locks in" in Schibsted, then `19:10:54` in mono, both at 15px.

10. **Duplicated or redundant headers.**
    - stub.png prints "TRT v LNS" as the page title *and* as the ticket.
    - match-live/done tickets print full names in caps ("KOLKATA KNIGHT RIDERS"), which read as ALL-CAPS headings. The mockup uses codes ("KKR vs SRH") with the full name small beneath.
    - Files: `src/app/stub/[id].tsx` header (title: "Sipho's stub", kicker "No. 0022" only). `src/features/match/Header.tsx:96`: show `team.short` in `display` at 64px, with `team.name` in `bodySmall`/`tone="soft"` below. That restores "vs" between the codes per rule 2.

11. **Left-stripe scoreboard rows (the cliché the brief bans).** LiveScoreboard (match-live.png y≈970-1065) and the result card (match-done.png y≈1048-1140) use a 6px vertical team-colour bar at the left of each row.
    - Files: `src/features/match/LiveScoreboard.tsx`, `src/features/match/Header.tsx` (result block)
    - Change: replace each bar with the `TeamBadge size="sm"` code plus a 3px underline (as on matches.png), then the full name.

12. **Copy and step consistency.**
    - The log flow says "Step 1 of 4 / Pick the match" (log-pick.png), then "Step 1 of 3 / Rate it" (log-rate.png). The count resets.
    - "Live reactions are coming soon." (match-live.png) is placeholder text left in production.
    - "No play today" is reused for an empty filter in rankings.tsx:55, where it reads as an error.
    - Files and changes:
      - `src/features/log/LogHeader.tsx:28`: compute against the full step list so "Rate it" shows "Step 2 of 4".
      - `src/features/match/LiveScoreboard.tsx:115`: delete the line.
      - `rankings.tsx:55`: set the title to "Nothing rated here yet" and the body to "No fan-rated {format} matches from the {era}. Try All."

---

## Detailed findings

**Copywriting (3/4).**
- Strong lines: "Every crown, every heartbreak.", "No review. The balls say it all.", "Two letters and we're off."
- Problems: the placeholder at LiveScoreboard.tsx:115; the step-count reset (LogHeader.tsx:28); "No play today" misused as a filter-empty title.
- The "your grave" phrasing appears twice (vault.png, profile.png). Vary one of them.

**Visuals (3/4).**
- The tickets on matches.png and log-pick.png match the mockup closely.
- Home's focal point is wrong (P0 #1).
- The match header with full caps names is heavier than the mockup.
- Full-width black buttons appear on match-live, match-done, match-classic and log-rate. That repeats a template; consider a half-width or stub-docked "Log this match".
- "Pin to Top 5" and "I watched this" render at equal weight.

**Color (2/4).**
- Concrete, white and stock are correct in both themes. Dark stocks (dark-home.png, dark-match-done.png) hold contrast.
- Red appears on more than 6 non-sanctioned roles: the split bar, the histogram peak, the "T20" format code, "Edit profile", the "IND need 160 off 98 balls" sentence on both matches.png and match-live.png (acceptable as a live marker, but bold red body text twice per screen), and the GF avatar default colour.
- Green is used correctly (only "WON").
- The `Panel accent` red band was not found in the screenshots. Good.

**Typography (3/4).**
- Big Shoulders is used for codes and numbers, Schibsted for prose, and mono for fine print. This is correct on most screens.
- Problems:
  - The rank glyph bug (P0 #2).
  - Mono in prose (P1 #9).
  - All-caps full team names in the match header (P1 #10).
  - Mono "LOGGED / AVG GIVEN" labels are fine as fine print.

**Spacing (3/4).**
- The 20px gutter is consistent.
- Gaps between sections run about 36-44px. "Live now" to "Tomorrow" on matches.png is about 48px, against 28px in the brief. Tighten SectionHeader `marginTop` to 28.
- Tab-bar overlap (P0 #4) and truncations (P1 #8).

**Experience design (2/4).**
- Loading, empty and error states exist (QueryState, Empty). Lock states are handled ("Locked 🔒").
- Problems:
  - The On-now ordering and scroll position (P0 #1).
  - The spinner over loaded content (P0 #3).
  - Empty states hidden under the FAB (P0 #4).
  - The disabled "Post" in stub.png is grey-on-grey. Use `t.ink` at opacity 0.35 so it does not read as broken.
  - The search screen leaves 60% of the screen blank. Add the "Recent searches" or "Classics" list below the chips.

Registry audit: there is no shadcn setup (React Native), so it was skipped.

## Files audited
- src/components/{ui.tsx,Screen.tsx,TicketStub.tsx}, src/theme/tokens.ts
- src/features/home/{OnNowStrip,YourCallCard}.tsx
- src/features/match/{FanRating,LiveScoreboard,Header}.tsx
- src/features/vault/{RankRow,OnThisDay}.tsx
- src/features/log/LogHeader.tsx
- src/features/predict/PredictionPicker.tsx
- src/app/vault/rankings.tsx, src/app/stub/[id].tsx
- The 22 PNGs in qa/shots2
