# Phase 6: Design pass & release prep - Context

**Gathered:** 2026-10-22
**Status:** Complete 2026-10-22 (see `06-01-SUMMARY.md`)

<domain>
## Phase Boundary

Give the finished app a look that reads as deliberate rather than generated, then get the repo ready to show: re-skin every screen in the chosen direction, check every screen in both themes from screenshots, replace the icons and splash, run the app on Android, write the README and push to GitHub.

Not in this phase: new features, store builds (EAS, App Store, Play Store need the developer's accounts), a real cricket API.

</domain>

<decisions>
## Implementation Decisions

### Why a design pass
- **D-01:** The first pass followed the plan's "heritage scorebook" brief (paper texture, ruled lines, Oswald, Source Serif, Plex Mono). The developer's verdict: it felt AI-made.
- **D-02:** Three directions were sketched on a design canvas: A Almanack (editorial, Wisden-like), B Pavilion Scoreboard (a manual scoreboard), C Matchday Ticket (collectible tickets). The developer chose **C, Matchday Ticket** on 2026-10-22. A and B are not pursued.

### Direction C: Matchday Ticket (developer's spec)
- **D-03 Ground and surfaces:** concrete grey ground #ECEAE4; flat white cards with no borders, no shadows, no paper texture and no ruled lines.
- **D-04 Ticket stock:** every match is printed on ticket stock coloured by competition: international pale blue #D7E4EE, IPL salmon #F4D6C8, Test cream #F3E7C9, WPL lilac, World Cup pale gold, BBL pale teal, SA20 pink, The Hundred periwinkle, Vault classics old card. Each ticket has a dotted perforation, notches, and a barcode stub with a red serial number.
- **D-05 Type:** Big Shoulders Display for team codes and big numerals; Schibsted Grotesk for UI and reviews; IBM Plex Mono for ticket fine print and stats.
- **D-06 Colour roles:** ink #16171A; overprint red #D23A2E for stamps, live and serials; pitch green #2F6B3A only for wins. Primary buttons are solid ink (red is no longer the primary action colour).
- **D-07 Ratings:** cricket balls, as before, and on feed stubs also a red rubber-stamp circle.
- **D-08 Floodlights:** the dark theme keeps the ticket stocks, darkened.
- **D-09:** Only the hex values above were specified. Stock colours for WPL, World Cup, BBL, SA20, The Hundred and Vault classics, and the darkened floodlights stocks, are picked during implementation and confirmed with the developer.

### Scope of the re-skin
- **D-10:** Screens, flows and section order stay as built in Phases 3 to 5 (see their UI-SPEC files). The re-skin changes tokens, fonts, surfaces and the ticket; copy is tightened only where the new look needs it (for example Match Page section titles), and those changes are recorded in the Phase 6 summary.
- **D-11:** The token role names stay where possible so screens don't all need edits; a `stock` map keyed by competition (plus Test and classic) is added to the palette.

### QA and release
- **D-12:** Screenshot QA covers every screen in both themes, driven from the command line with the existing QA driver (`src/lib/qaDriver.ts`: Metro started with `EXPO_PUBLIC_QA_DRIVER=<url>`, the app polls the URL for a route or `theme:paper` / `theme:floodlights`).
- **D-13:** Distribution stays as tonight: native iOS dev build on the simulator, Expo Go on phones. Android is checked on an emulator or a device. *(Superseded during the phase: Expo Go from the stores only opens the newest SDK, so phones use a development build. See the summary.)*

### Claude's Discretion
- The barcode drawing, notch geometry and how the red stamp circle sits on a feed stub.

</decisions>

<canonical_refs>
## Canonical References

- `.planning/phases/03-core-loop/03-UI-SPEC.md`, `04-UI-SPEC.md`, `05-UI-SPEC.md`: the screens as built
- `docs/PLAN.md`: Design Direction (the original brief this phase replaces)
- `AGENTS.md`: add font packages with `npx expo install`; don't hand-edit `ios/` or `android/`

</canonical_refs>

<code_context>
## Existing Code Insights

The look is concentrated in a few files: `src/theme/tokens.ts`, the font list in `src/app/_layout.tsx`, `src/components/Paper.tsx`, `src/components/TicketStub.tsx`, `src/components/ui.tsx` (Button kinds), `src/components/Txt.tsx`, plus `app.json` (splash and adaptive-icon background #F5F0E6) and `assets/`. Feature code mostly reads tokens through `useTheme()`, so most screens change without edits.

</code_context>

<deferred>
## Deferred Ideas

- EAS builds and store submission (v2, STORE-01)
- Directions A Almanack and B Pavilion Scoreboard

</deferred>

---

*Phase: 6-Design pass & release prep*
*Context gathered: 2026-10-22*
