// Formatting helpers: dates, countdowns, scores. Pure functions.
import type { InningsScore, Match } from '@/domain/types';

const MONTHS = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];
const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const pad2 = (n: number) => String(n).padStart(2, '0');

/** "02 APR 2011" — the date printed on a ticket. Classics use their recorded date. */
export function ticketDate(m: Pick<Match, 'startTime' | 'classic'>): string {
  if (m.classic?.date) {
    const [y, mo, d] = m.classic.date.split('-').map(Number);
    return `${pad2(d)} ${MONTHS[mo - 1]} ${y}`;
  }
  const d = new Date(m.startTime);
  return `${pad2(d.getDate())} ${MONTHS[d.getMonth()]} ${d.getFullYear()}`;
}

/** "19:30" local. */
export const clock = (iso: string) => {
  const d = new Date(iso);
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}`;
};

const startOfDay = (t: number) => {
  const d = new Date(t);
  d.setHours(0, 0, 0, 0);
  return d.getTime();
};

/** "Today", "Tomorrow", "Yesterday", or "Sat 24 Oct". */
export function dayLabel(iso: string, now = Date.now()): string {
  const diff = Math.round((startOfDay(Date.parse(iso)) - startOfDay(now)) / 86_400_000);
  if (diff === 0) return 'Today';
  if (diff === 1) return 'Tomorrow';
  if (diff === -1) return 'Yesterday';
  const d = new Date(iso);
  return `${DAYS[d.getDay()]} ${d.getDate()} ${MONTHS[d.getMonth()].charAt(0)}${MONTHS[d.getMonth()].slice(1).toLowerCase()}`;
}

/** Stable key for grouping matches by local day. */
export const dayKey = (iso: string) => {
  const d = new Date(iso);
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;
};

/** "in 25 min", "in 3 h", "2 h ago", "just now". */
export function relative(iso: string, now = Date.now()): string {
  const ms = Date.parse(iso) - now;
  const abs = Math.abs(ms);
  const mins = Math.round(abs / 60_000);
  let span: string;
  if (mins < 1) return 'just now';
  if (mins < 60) span = `${mins} min`;
  else if (mins < 60 * 24) span = `${Math.round(mins / 60)} h`;
  else span = `${Math.round(mins / 1440)} d`;
  return ms > 0 ? `in ${span}` : `${span} ago`;
}

/** "1:04:09" or "24:09" countdown to a timestamp; "0:00" once passed. */
export function countdown(iso: string, now = Date.now()): string {
  const s = Math.max(0, Math.floor((Date.parse(iso) - now) / 1000));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return h > 0 ? `${h}:${pad2(m)}:${pad2(sec)}` : `${m}:${pad2(sec)}`;
}

/** "287/5 (48.2)"; an all-out innings is just the runs, as scorebooks print it: "412 (132.1)". */
export const inningsLine = (i: InningsScore) =>
  `${i.runs}${i.wickets >= 10 ? '' : `/${i.wickets}`}${i.overs ? ` (${i.overs})` : ''}`;

export const pct = (share: number) => `${Math.round(share * 100)}%`;

export const stubNo = (n: number) => `No. ${String(n).padStart(4, '0')}`;

/** Predictions lock at the first ball. */
export function isLocked(m: Pick<Match, 'startTime' | 'state'>, now = Date.now()): boolean {
  return m.state !== 'upcoming' || now >= Date.parse(m.startTime);
}
