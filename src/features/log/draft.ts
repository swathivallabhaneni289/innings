// Log-flow draft: everything the fan fills in across the four steps, plus helpers to prefill and save it.
import type { NewLog } from '@/data';
import { watchPlaceLabel } from '@/domain/constants';
import type { MatchLog, WatchPlace } from '@/domain/types';

export type StepId = 'match' | 'rating' | 'players' | 'place';

export const STEP_TITLES: Record<StepId, string> = {
  match: 'Pick the match',
  rating: 'Rate it',
  players: 'Rate the players',
  place: 'Where I watched it',
};

export const STEP_SHORT: Record<StepId, string> = {
  match: 'Match',
  rating: 'Rating',
  players: 'Players',
  place: 'Place',
};

export interface Draft {
  matchId: string | null;
  rating: number; // 0 = not rated yet
  instantClassic: boolean;
  playerRatings: Record<string, number>; // 0 / absent = skipped
  fanPotmId: string | null;
  watchedAt: WatchPlace | null; // null = use the default (last place, else Home)
  placeNote: string;
  watchedWith: string[];
  companions: number;
  photoUri?: string;
  review: string;
}

export const emptyDraft = (matchId: string | null): Draft => ({
  matchId,
  rating: 0,
  instantClassic: false,
  playerRatings: {},
  fanPotmId: null,
  watchedAt: null,
  placeNote: '',
  watchedWith: [],
  companions: 0,
  photoUri: undefined,
  review: '',
});

/** Re-logging a match starts from the existing stub. */
export const draftFromLog = (log: MatchLog): Draft => ({
  matchId: log.matchId,
  rating: log.rating,
  instantClassic: log.instantClassic,
  playerRatings: { ...log.playerRatings },
  fanPotmId: log.fanPotmId,
  watchedAt: log.watchedAt,
  placeNote: log.placeNote ?? '',
  watchedWith: [...log.watchedWith],
  companions: log.companions ?? 0,
  photoUri: log.photoUri,
  review: log.review ?? '',
});

export function toNewLog(d: Draft, matchId: string, place: WatchPlace, xi: Set<string>): NewLog {
  const playerRatings: Record<string, number> = {};
  for (const [pid, v] of Object.entries(d.playerRatings)) if (v >= 1 && xi.has(pid)) playerRatings[pid] = Math.round(v);
  const placeNote = d.placeNote.trim();
  const review = d.review.trim();
  return {
    matchId,
    rating: d.rating,
    instantClassic: d.instantClassic,
    playerRatings,
    fanPotmId: d.fanPotmId && xi.has(d.fanPotmId) ? d.fanPotmId : null,
    watchedAt: place,
    placeNote: placeNote || undefined,
    watchedWith: d.watchedWith,
    companions: d.companions > 0 ? d.companions : undefined,
    photoUri: d.photoUri,
    review: review || undefined,
  };
}

/** Where the fan logged from last (their newest log's place), or null before their first log. */
export function lastPlaceOf(logs: Pick<MatchLog, 'watchedAt' | 'createdAt'>[]): WatchPlace | null {
  let last: Pick<MatchLog, 'watchedAt' | 'createdAt'> | null = null;
  for (const l of logs) if (!last || l.createdAt > last.createdAt) last = l;
  return last?.watchedAt ?? null;
}

/**
 * A log saved from the Match Page: a rating tapped on its balls or a review typed under Reviews. An existing log keeps
 * everything else as it was; a first one starts like a new log in the Log flow, watched at `place`.
 */
export function relog(log: MatchLog | null, matchId: string, place: WatchPlace, patch: { rating?: number; review?: string }): NewLog {
  if (!log) return toNewLog({ ...emptyDraft(matchId), rating: patch.rating ?? 0, review: patch.review ?? '' }, matchId, place, new Set());
  const { id: _id, userId: _userId, createdAt: _createdAt, stubNo: _stubNo, ...rest } = log;
  const review = patch.review === undefined ? rest.review : patch.review.trim() || undefined;
  return { ...rest, rating: patch.rating ?? rest.rating, review };
}

/** One or two words for a rating, as a fan would say it. */
export function verdict(rating: number): string {
  if (rating <= 0) return 'Not rated yet';
  const words = ['Dead rubber', 'Damp squib', 'Snoozer', 'Below par', 'Par score', 'Decent contest', 'Good game', 'Cracker', 'Thriller', 'All-timer'];
  return words[Math.min(9, Math.max(0, Math.round(rating * 2) - 1))];
}

/** "📍 Hostel room, with 6 people" */
export function placeLine(log: Pick<MatchLog, 'watchedAt' | 'placeNote' | 'companions'>): string {
  const where = log.placeNote?.trim() || watchPlaceLabel(log.watchedAt);
  const n = log.companions ?? 0;
  return `📍 ${where}${n > 0 ? `, with ${n} ${n === 1 ? 'person' : 'people'}` : ''}`;
}
