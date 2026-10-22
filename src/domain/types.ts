// Core domain model for Innings. Every screen and data module builds on these types.

export type Format = 'T20' | 'ODI' | 'Test' | 'The Hundred';
export type Gender = 'men' | 'women';

/** Competition a match belongs to. The Matches tab's "India" chip is a team filter (ind / ind-w), not a competition. */
export type CompetitionId =
  | 'international'
  | 'ipl'
  | 'wpl'
  | 'world-cup'
  | 'bbl'
  | 'sa20'
  | 'the-hundred';

export interface Competition {
  id: CompetitionId;
  name: string; // "Indian Premier League"
  short: string; // "IPL"
}

export interface Team {
  id: string; // 'ind', 'aus-w', 'csk'
  name: string; // "India"
  short: string; // "IND"
  colour: string; // primary kit colour, used for small badges only
  gender: Gender;
  kind: 'international' | 'franchise';
}

export type PlayerRole = 'BAT' | 'BOWL' | 'AR' | 'WK';

export interface Player {
  id: string; // 'ind-kohli'
  name: string; // "Virat Kohli"
  teamId: string;
  role: PlayerRole;
}

export interface Venue {
  id: string;
  name: string; // "Wankhede Stadium"
  city: string; // "Mumbai"
  country: string;
}

export type MatchState = 'upcoming' | 'live' | 'completed';

export interface InningsScore {
  teamId: string;
  runs: number;
  wickets: number;
  overs: string; // "18.4", or balls for The Hundred e.g. "100 balls"
}

export interface LiveScore {
  innings: InningsScore[];
  battingTeamId: string;
  statusText: string; // "IND need 34 off 21 balls"
}

export interface MatchResult {
  winnerId: string | null; // null = tie / draw / no result
  text: string; // "India won by 6 wickets"
  innings: InningsScore[];
  topRunScorerIds: string[]; // several when tied
  topWicketTakerIds: string[]; // several when tied
  officialPotmId: string | null;
}

export type CollectionId =
  | 'world-cup-finals'
  | 'greatest-chases'
  | 'test-thrillers'
  | 'super-over-madness'
  | 'womens-classics';

export interface Collection {
  id: CollectionId;
  title: string; // "World Cup Finals"
  blurb: string;
}

/** Extra facts for archive matches in the Vault (real, historically verified matches). */
export interface ClassicInfo {
  date: string; // 'YYYY-MM-DD', the day the match was played (first day for Tests)
  competitionName: string; // "ICC Cricket World Cup 2011"
  headline: string; // one-line hook, e.g. "Dhoni finishes it with a six"
  scoreline: string; // "SL 274/6 (50) · IND 277/4 (48.2)"
  potmName?: string; // official Player of the Match, as a name
  collections: CollectionId[];
  era: string; // '1980s', '2010s'
}

export interface Match {
  id: string;
  kind: 'fixture' | 'classic';
  competitionId: CompetitionId;
  format: Format;
  gender: Gender;
  stage: string; // "3rd T20I", "Final", "Match 21"
  teamIds: [string, string];
  venue: Venue;
  startTime: string; // ISO timestamp; first ball. Predictions lock at this moment.
  state: MatchState; // derived from the clock for fixtures; always 'completed' for classics
  live?: LiveScore; // present while state === 'live'
  result?: MatchResult; // present once state === 'completed' (classics: winner + text + innings may be partial)
  xi?: Record<string, string[]>; // teamId → playing XI player ids (fixtures only)
  classic?: ClassicInfo; // archive matches only
}

export type WatchPlace = 'stadium' | 'home' | 'friends' | 'hostel' | 'office' | 'cafe' | 'on-the-go';

export interface User {
  id: string;
  name: string; // "Aditi Rao"
  handle: string; // "aditi"
  avatarColour: string;
  favouriteTeamId: string;
  bio: string; // "Dhoni finishes, I cry"
  photoUri?: string;
}

/** One logged match: the "ticket stub". */
export interface MatchLog {
  id: string;
  userId: string;
  matchId: string;
  rating: number; // 0.5–5 in half-ball steps
  instantClassic: boolean;
  playerRatings: Record<string, number>; // playerId → 1..10 (skipped players are absent)
  fanPotmId: string | null;
  watchedAt: WatchPlace;
  placeNote?: string; // "Hostel room", "my uncle's place in Hyderabad"
  watchedWith: string[]; // tagged friend user ids
  companions?: number; // "with 6 people"
  photoUri?: string;
  review?: string;
  createdAt: string; // ISO
  stubNo: number; // serial printed on the ticket
}

export interface Prediction {
  userId: string;
  matchId: string;
  winnerId: string;
  topRunScorerId: string;
  topWicketTakerId: string;
  createdAt: string; // must be before the match startTime
}

export interface PredictionOutcome {
  matchId: string;
  userId: string;
  winnerCorrect: boolean;
  topScorerCorrect: boolean;
  topWicketCorrect: boolean;
  upset: boolean; // picked the winner when fewer than 30% of fans did
  base: number; // points before the streak multiplier (includes upset bonus)
  multiplier: number;
  points: number; // base × multiplier, rounded
  streak: number; // correct-winner streak after this match
  settled: boolean; // false while the match has no result
}

export interface Reply {
  id: string;
  logId: string;
  userId: string;
  text: string;
  createdAt: string;
}

export type ThemePreference = 'auto' | 'paper' | 'floodlights';

export interface BadgeDef {
  id: 'called-the-upset' | 'five-match-streak' | 'sharpshooter' | 'first-call' | 'stadium-regular';
  title: string; // "Called the Upset"
  description: string;
}
