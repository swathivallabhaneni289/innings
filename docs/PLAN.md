# Innings — The Fan's Cricket Diary

_This is the plan as I first wrote it, kept for the record. The app as built is described in the [PRD](PRD.md), and the decisions that filled the plan's gaps are in its Build decisions table. Innings was a working name; the alternatives were Over & Out, Second Innings and Stumped._

Log every match you watch, predict before the first ball, rate the players, remember where you watched it, and build your all-time list of greatest games.

## The Big Idea

Cricbuzz tells you **what happened**. CricHeroes helps you **score your own games**. Nobody captures **how fans felt**. Innings is a taste and memory app for cricket fans. Every match is a "release," and fans rush in to give their take, just like opening night on Letterboxd.

**Core features:**
- Predict, then review (with bragging points)
- Fan player ratings, out of 10
- "Where I watched it" memories
- Classic matches archive

## Design Direction (so it doesn't look AI-made)

Skip the generic look of dark mode, neon gradients, and rounded cards everywhere. Use a **heritage scorebook** feel instead, like an old paper scorecard mixed with a match ticket.

- **Colours:** off-white "paper" background (#F5F0E6), deep ink navy text (#1C2333), cricket-ball red accent (#B3261E), and pitch green only for wins and success (#2F6B3A). Add a dark "floodlights" mode for night matches.
- **Type:** a bold condensed font for scores and headings (Oswald or Bebas Neue), a clean serif for reviews (Source Serif), and a monospaced font for stats, so it feels like a scoreboard.
- **Signature element: the ticket stub.** Every logged match appears as a torn ticket with the teams, date, venue, your rating, and a perforated edge. Your profile becomes a collection of stubs.
- **Ratings:** use 5 small cricket balls instead of stars, including half-balls.
- **Motion:** a little ball spin when you submit a rating, and a stamp effect ("LOGGED") when you save a match.
- **Texture:** faint paper grain and scorecard lines in the background. Keep it subtle.

## App Structure: 5 Tabs

The bottom navigation bar, from left to right:
1. **Home**: what's happening now and what your friends are saying
2. **Matches**: browse live, upcoming, and recent matches
3. **➕ Log** (a big red ball button in the centre): quickly log a match
4. **Vault**: the classic matches archive and all-time rankings
5. **Profile**: your stubs, stats, and memories

## Tab 1 — Home

From top to bottom:
1. **Top bar:** the app logo on the left, with notifications and search on the right.
2. **"On Now" strip:** a horizontal row of live and today's match cards. Each shows the teams, the live score, and a **Predict** button if the match hasn't started (it turns into **Locked 🔒** at the first ball).
3. **"Your Call" prediction card:** a big card for the next major match. You pick the winner, the top run-scorer, and the top wicket-taker in three quick taps. It shows the % of friends who picked each team.
4. **Friends' Feed:** recent ticket-stub logs from friends, showing the rating, a short review, their fan Player of the Match, and where they watched ("📍 Hostel room, with 6 people"). You can like, reply, or tap "I was there too."
5. **Just Finished:** matches that ended in the last 24 hours, with the average fan rating and the top review. This is the "opening night" buzz section.
6. **Bragging Board (small widget):** this week's top predictors among your friends.

## Tab 2 — Matches

- **Filter chips at the top:** Live · Upcoming · Recent · plus competitions (India, IPL, WPL, World Cup, BBL, SA20, The Hundred…)
- **Match list:** cards grouped by day, showing the teams, format (T20, ODI, Test), venue, and fan rating once it's finished.

### The Match Page (the most important screen)

The page changes based on the match state:

**Before the match**
- Header with teams, venue, and start time
- Prediction panel (winner, top scorer, top wicket-taker)
- "Friends' predictions" split bar
- "I'm watching" button, to set where you'll watch from

**During the match**
- Live score (from a cricket data API)
- Friends who are watching right now
- Quick-reaction buttons (optional for later)

**After the match**
- **Fan Rating:** the big average score, with the distribution of 1 to 5 balls
- **Fan Player Ratings:** each player with an average rating out of 10. The highest is crowned **Fans' Player of the Match**, shown next to the official one, with a "Fans disagree!" badge when they differ.
- **Prediction results:** who got it right, with points awarded
- **Reviews:** popular reviews and friends' reviews
- **Where fans watched:** a small summary ("62% at home · 21% with friends · 9% at the stadium")
- **Log this match** button

## Tab 3 — Log (centre button)

A fast flow of 4 steps that takes under 60 seconds:
1. **Pick the match:** it suggests matches that are live now or just ended, with search for older ones, back to the first World Cups: every World Cup, Champions Trophy, Asia Cup and IPL match is in Innings' own archive.
2. **Rate it:** 1 to 5 cricket balls, plus an optional ❤️ "instant classic" tag.
3. **Rate the players:** a swipeable list of both XIs with a 1 to 10 slider each. You can skip players, and you pick your own Player of the Match.
4. **Where I watched it:** choose Stadium, Home, Friend's place, Hostel or PG, Office, Café or pub, or On the go. Add who you were with (tag friends), an optional photo, and an optional review.

On save, a **ticket stub** is printed with the "LOGGED" stamp animation, and you can share it as an Instagram story card.

## Tab 4 — Vault (Classic Matches Archive)

This keeps the app alive on days with no matches.
- **Featured collections:** "World Cup Finals," "Greatest Chases," "Test Match Thrillers," "Super Over Madness," "Women's Cricket Classics"
- **All-time fan rankings:** the highest-rated matches ever, which you can filter by format and era
- **"I watched this" logging:** log old matches with the memory of where you were ("2011 final: at my uncle's place in Hyderabad")
- **On This Day:** a match that happened on today's date in history, with a nudge to log it if you watched it
- **Your Top 5:** you can pin your personal greatest matches, from the Vault or any match's page, and reorder or trim them later

## Tab 5 — Profile

From top to bottom:
1. **Header:** your photo, name, favourite team badge, and a one-line fan bio ("Dhoni finishes, I cry")
2. **Top 5 Matches:** shown as large ticket stubs, like Letterboxd's top 4 films
3. **Stats row:** matches logged · average rating given · prediction accuracy % · bragging points
4. **Stub Collection:** a grid of every match you've logged
5. **Where I've Watched:** a breakdown of your watching spots, plus a stadium list ("Stadiums visited: 3")
6. **Predictor Rank:** your rank among friends, with badges like "Called the Upset" or "5-Match Streak"
7. **Season Wrapped (later):** an end-of-season shareable summary

## Bragging Points System

- Correct winner: **+10**
- Correct top run-scorer: **+15**
- Correct top wicket-taker: **+15**
- Calling an upset (picking the team fewer than 30% of fans chose): **+10 bonus**
- Streaks multiply your points

Points show on the Bragging Board, the Profile, and in prediction results.

## MVP Build Plan

**Week 1 — Foundations**: Expo + Supabase; a cricket data API for fixtures, squads, and scores (CricketData.org or SportMonks); auth and profiles.
**Week 2 — Core loop**: Matches tab and the match page (before and after states); the Log flow; ticket stub design.
**Week 3 — Social and predictions**: follow friends and the Home feed; the prediction system, which locks at the first ball, plus scoring; Fans' Player of the Match.
**Week 4 — Vault and polish**: seed the Vault with around 50 classic matches by hand; Profile stats and Top 5; a shareable story card.
**Later:** live mood reactions, Season Wrapped, team fan clubs, and AI summaries of fan reviews.

## Tips for Launch

- **Launch right before a big series** so there's instant release-day energy.
- **Seed it with your friends and your AI club** first, and get 20 people logging every match.
- **The share card is your growth engine.** Make the ticket stub beautiful enough that people post it on their own.

---

# Build decisions (this repo)

These settle gaps in the plan. Scoring and aggregates are implemented in Go: `internal/domain/scoring.go` and `internal/domain/aggregate.go`.

| Topic | Decision |
| --- | --- |
| Streak multiplier | Consecutive correct-winner calls: 1–2 → ×1, 3–4 → ×1.5, 5+ → ×2. Applied to that match's points (incl. upset bonus). A wrong winner resets the streak; a tie/no result leaves it unchanged. |
| Upset threshold | Measured on the split of all fan predictions **at lock time** (first ball). Bonus only when your pick wins. |
| Ties for top scorer / wicket-taker | Every tied player counts as correct. |
| Ratings | Half-balls everywhere: 0.5–5 in 0.5 steps. Distribution chart shows 10 half-ball bars labelled 1–5. |
| Fans' Player of the Match | Highest average rating with at least 3 ratings. "Fans disagree!" when it differs from the official pick. |
| Bragging Board window | Points from matches that finished in the last 7 days. |
| Backend | A Go API (`cmd/api`, `internal/...`, stdlib `net/http`) replaces Supabase — the same Expo + Go split as RNDMRoll. SQLite (pure-Go `modernc.org/sqlite`) with embedded migrations, so `make api` runs with zero setup; the SQL stays portable to Postgres for production. |
| Accounts | Email and a password (2026-10-30, phase 10). Sign-up asks only for an email and a password; a 6-digit code emailed to the address confirms it before the account opens. Forgot password works with another emailed code, and resetting or changing a password signs out the other devices. This is the third version: v1.0 had guest accounts (`POST /v1/bootstrap` issued a token stored on the device), and v1.1 signed in with a phone number and a texted code, like Beli. Email replaced the phone because real texts cost money for every sign-in, and without them nobody checked the number, so anyone could type in another fan's number and pose as them. An emailed code costs nothing through an ordinary Gmail account and proves the address. Contacts now match on SHA-256 hashes of lower-cased email addresses. |
| Teams a fan backs | Up to five, main team first (2026-10-30, phase 10). `User.teamIds` lists them and `favouriteTeamId` stays as the main team, so everything that shows one team keeps working. The picker in sign-up and Settings offers the international sides and all 10 IPL sides. |
| Data | No API keys needed. Matches are real (see Cricket API). A deterministic simulation in Go (`internal/sim`) supplies the simulated friends, public fans and the Vault's crowd, and runs the whole world, fixtures and live scores included, when `INNINGS_MATCHES=sim`. |
| Cricket API | Real matches come from ESPN's public scores feed (`internal/feed`), the JSON behind espn.in's scores pages. It covers internationals, World Cups, IPL, WPL, BBL and WBBL, SA20 and The Hundred; domestic, A and youth sides are left out. One poller reads the scoreboard every minute while a match is on (every five minutes otherwise) and looks a week ahead every six hours. After time off it asks for each day it missed, up to two weeks back (ESPN's dated scoreboard answers for recent days only). It reads each match's summary for the announced squads, every five minutes in the two hours before the start for the toss and playing XIs, and after the match for the top scorers, top wicket-takers and player of the match. Everything is saved in `feed_rows`. With the feed on, the simulation keeps only its fans, their Vault reviews and the Vault; its made-up fixtures, crowds and predictions never sit next to a real match. Decided 2026-10-29, after the owner asked for no paid services and no sign-ups. The alternatives were CricketData.org's free plan (100 requests a day, needs an account) and SportMonks (from €29 a month). Scraping Google or bcci.tv was ruled out on their terms. ESPN's feed is unofficial and may change without notice; that's fine for a personal app, not for a commercial one. |
| Vault | Real historical matches only; each date, venue and result was checked against public sources. Unverifiable entries were dropped. |

# Architecture

- `src/app/` — Expo Router routes. `(tabs)` holds the 5 tabs; the centre Log tab opens the `/log` modal flow.
- `src/theme/` — tokens (paper + floodlights), type scale, fonts.
- `src/components/` — the design system: `Paper`, `TicketStub`, `BallRating`, `Stamp`, `Text`, `Button`, `Chip`, etc.
- `src/domain/` — the JSON contract (`types.ts`, `api.ts`) shared with Go (`internal/domain`).
- `src/data/` — react-query hooks over the Go API (`@/data`) and on-device prefs (theme, token).
- `cmd/api`, `internal/httpapi`, `internal/store`, `internal/auth`, `internal/sim`, `internal/domain` — the Go API.
- `src/features/<area>/` — components owned by one screen area.
