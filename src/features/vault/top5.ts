// Vault helpers: pin / unpin a match in the fan's Top 5, and small formatting bits for archive matches.
import { useCallback, useState } from 'react';

import { useActions } from '@/data';
import { MAX_TOP5 } from '@/domain/constants';
import type { Match } from '@/domain/types';

/** "2011" — the year a classic was played (fixtures fall back to their start time). */
export const matchYear = (m: Pick<Match, 'startTime' | 'classic'>) =>
  m.classic?.date ? m.classic.date.slice(0, 4) : String(new Date(m.startTime).getFullYear());

/**
 * "ICC Men's T20 World Cup" from "ICC Men's T20 World Cup 2022" (also drops seasons like "2020-21"), for fine print
 * that already prints the year big beside it.
 */
export const competitionNoYear = (name: string) => name.replace(/\s+\d{4}(-\d{2,4})?$/, '');

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/** "2 Apr" — the anniversary day of a classic. */
export function anniversary(m: Pick<Match, 'startTime' | 'classic'>): string {
  if (m.classic?.date) {
    const [, mo, d] = m.classic.date.split('-').map(Number);
    return `${d} ${MONTHS[mo - 1]}`;
  }
  const d = new Date(m.startTime);
  return `${d.getDate()} ${MONTHS[d.getMonth()]}`;
}

/**
 * Pin / unpin against the fan's current Top 5 (ids in order). Keeps the order, refuses a sixth pin,
 * and reports errors per match so screens can show them inline next to the row that failed.
 */
export function useTop5(currentIds: string[]) {
  const { setTop5 } = useActions();
  const [busyId, setBusyId] = useState<string | null>(null);
  const [message, setMessage] = useState<{ matchId: string; text: string } | null>(null);

  const save = useCallback(
    async (matchId: string, next: string[]) => {
      setBusyId(matchId);
      setMessage(null);
      try {
        await setTop5(next);
      } catch (e) {
        setMessage({ matchId, text: e instanceof Error ? e.message : 'Could not update your Top 5.' });
      } finally {
        setBusyId(null);
      }
    },
    [setTop5],
  );

  const isPinned = (matchId: string) => currentIds.includes(matchId);

  const unpin = (matchId: string) => save(matchId, currentIds.filter((id) => id !== matchId));

  const toggle = (matchId: string) => {
    if (isPinned(matchId)) return unpin(matchId);
    if (currentIds.length >= MAX_TOP5) {
      setMessage({ matchId, text: `Your Top ${MAX_TOP5} is full. Tap Edit on it in the Vault to take one out.` });
      return;
    }
    return save(matchId, [...currentIds, matchId]);
  };

  return { isPinned, toggle, unpin, busyId, message, full: currentIds.length >= MAX_TOP5 };
}
