---
phase: 05-vault-profile
plan: 01
subsystem: data
tags: [vault, fact-check, seed-data, sources]

requires:
  - phase: 02-api-simulated-world
    provides: ClassicSeed format and the sim team list
provides:
  - 58 real classic matches in internal/sim/classics.json
  - One source per match in docs/VAULT_SOURCES.md
affects: [Vault tab, rankings, On This Day, demo history (two demo logs are classics)]

tech-stack:
  added: []
  patterns: ["seed data embedded in the binary", "independent multi-agent verification before data ships"]

key-files:
  created: [internal/sim/classics.json, docs/VAULT_SOURCES.md]
  modified: []

requirements-completed: [VAULT-06]

completed: 2026-10-22
status: complete
---

# Phase 5 Plan 1: Vault Seed and Fact-Check Summary

**58 real classic matches from the 1975 World Cup final to the 2025 Women's World Cup final, each checked by three independent agents against public sources; all 58 confirmed, none corrected, none dropped.**

## Performance

- **Completed:** 2026-10-22
- **Duration:** not instrumented
- **Commits:** none yet (uncommitted)

## Accomplishments

- **The seed** (`internal/sim/classics.json`): 58 matches, oldest `wc-1975-final` (1975-06-21), newest the 2025 Women's World Cup final (2025-11-02). By format: 28 ODI, 18 T20, 12 Test. By gender: 47 men's, 11 women's. By era: 1970s 1, 1980s 6, 1990s 7, 2000s 7, 2010s 16, 2020s 21. Two have no winner (a tie or a draw).
- **Collections** (a match can be in several): World Cup Finals 23, Greatest Chases 17, Test Match Thrillers 12, Women's Cricket Classics 11, Super Over Madness 6.
- **Fact-check:** three checker agents verified every entry independently against web sources. Tally: 58 confirmed, 0 corrected, 0 dropped.
- **Sources** (`docs/VAULT_SOURCES.md`): one row per match with the date and a link: Wikipedia 32, ESPNcricinfo and its archives 18, Wisden 3, and one each from cricket.com.au, Lord's, The Irish Times, the Daily FT and newsonair.gov.in. 58 rows, matching the 58 entries.

## Decisions Made

- The plan asked for "around 50"; 58 made it through, so all 58 shipped.
- Tests are dated by their first day; the Vault's On This Day uses that date.

## Deviations from Plan

None.

## Issues Encountered

None recorded. No entry needed a correction.

## Next Phase Readiness

The Vault screens (05-02) and the rankings route serve these matches; `TestClassics` checks they load, are ordered and are findable by id.

## Self-Check: PASSED

`go test ./internal/sim -run TestClassics` passes; 58 entries and 58 source rows (counted with `python3` and `grep -c`).
