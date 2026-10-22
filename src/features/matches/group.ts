// Orders and groups the Matches tab list: live first, then today, then upcoming days (soonest first),
// then recent days (newest first). One rule for every filter: a filtered list simply has fewer kinds.
import type { Match } from '@/domain/types';
import { dayKey, dayLabel } from '@/lib/format';

export interface MatchGroup {
  key: string;
  title: string; // "Live now", "Today", "Tomorrow", "Sat 24 Oct"
  date?: string; // "Thu 22 Oct" printed beside relative titles
  live?: boolean;
  matches: Match[];
}

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

const shortDate = (iso: string) => {
  const d = new Date(iso);
  return `${DAYS[d.getDay()]} ${d.getDate()} ${MONTHS[d.getMonth()]}`;
};

const start = (m: Match) => Date.parse(m.startTime);

export function groupMatches(matches: Match[], now: number): MatchGroup[] {
  const live = matches.filter((m) => m.state === 'live').sort((a, b) => start(a) - start(b));
  const upcoming = matches.filter((m) => m.state === 'upcoming').sort((a, b) => start(a) - start(b));
  const recent = matches.filter((m) => m.state === 'completed').sort((a, b) => start(b) - start(a));
  const today = dayKey(new Date(now).toISOString());

  const groups: MatchGroup[] = [];
  if (live.length > 0) groups.push({ key: 'live', title: 'Live now', live: true, matches: live });

  const dayGroup = (m: Match, prefix: string): MatchGroup => {
    const title = dayLabel(m.startTime, now);
    const relative = title === 'Today' || title === 'Tomorrow' || title === 'Yesterday';
    return { key: `${prefix}-${dayKey(m.startTime)}`, title, date: relative ? shortDate(m.startTime) : undefined, matches: [] };
  };

  // Today: what's still to come, then what has finished.
  const todays = [...upcoming, ...recent].filter((m) => dayKey(m.startTime) === today);
  if (todays.length > 0) groups.push({ ...dayGroup(todays[0], 'today'), matches: todays });

  const addByDay = (list: Match[], prefix: string) => {
    let current: MatchGroup | null = null;
    let currentDay = '';
    for (const m of list) {
      const day = dayKey(m.startTime);
      if (day === today) continue;
      if (!current || currentDay !== day) {
        current = dayGroup(m, prefix);
        currentDay = day;
        groups.push(current);
      }
      current.matches.push(m);
    }
  };
  addByDay(upcoming, 'up');
  addByDay(recent, 'rec');
  return groups;
}
