import type { Collection, Competition, CompetitionId, WatchPlace } from './types';

export const WATCH_PLACES: { id: WatchPlace; label: string; emoji: string }[] = [
  { id: 'stadium', label: 'Stadium', emoji: '🏟️' },
  { id: 'home', label: 'Home', emoji: '🏠' },
  { id: 'friends', label: "Friend's place", emoji: '🛋️' },
  { id: 'hostel', label: 'Hostel or PG', emoji: '🛏️' },
  { id: 'office', label: 'Office', emoji: '💼' },
  { id: 'cafe', label: 'Café or pub', emoji: '🍺' },
  { id: 'on-the-go', label: 'On the go', emoji: '🚆' },
];

export const watchPlaceLabel = (id: WatchPlace) => WATCH_PLACES.find((w) => w.id === id)?.label ?? id;

export const COMPETITIONS: Competition[] = [
  { id: 'international', name: 'International', short: 'INTL' },
  { id: 'ipl', name: 'Indian Premier League', short: 'IPL' },
  { id: 'wpl', name: "Women's Premier League", short: 'WPL' },
  { id: 'world-cup', name: 'World Cup', short: 'World Cup' },
  { id: 'bbl', name: 'Big Bash League', short: 'BBL' },
  { id: 'sa20', name: 'SA20', short: 'SA20' },
  { id: 'the-hundred', name: 'The Hundred', short: 'The Hundred' },
];

export const competitionShort = (id: CompetitionId) => COMPETITIONS.find((c) => c.id === id)?.short ?? id;

export const COLLECTIONS: Collection[] = [
  { id: 'world-cup-finals', title: 'World Cup Finals', blurb: 'Every crown, every heartbreak.' },
  { id: 'greatest-chases', title: 'Greatest Chases', blurb: 'Targets that should have been too many.' },
  { id: 'test-thrillers', title: 'Test Match Thrillers', blurb: 'Five days, decided in a session.' },
  { id: 'super-over-madness', title: 'Super Over Madness', blurb: 'Tied after the full distance.' },
  { id: 'womens-classics', title: "Women's Cricket Classics", blurb: 'The games that changed the women’s game.' },
];

/** Ratings are 0.5–5 in half-ball steps. */
export const RATING_STEPS = [0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5] as const;
export const MAX_TOP5 = 5;
