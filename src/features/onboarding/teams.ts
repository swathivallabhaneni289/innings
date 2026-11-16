// The teams a fan backs, as an ordered list with the main team first. Pure, so it's tested without a screen.
/** How many teams one fan can back (the server's domain.MaxTeams). */
export const MAX_TEAMS = 5;

/**
 * Tapping a team: a picked one is dropped, a new one joins the end. At the limit nothing changes and `full` says why.
 * The main team can be dropped too; the next one along takes its place.
 */
export function toggleTeam(ids: string[], id: string, max = MAX_TEAMS): { ids: string[]; full: boolean } {
  if (ids.includes(id)) return { ids: ids.filter((x) => x !== id), full: false };
  if (ids.length >= max) return { ids, full: true };
  return { ids: [...ids, id], full: false };
}

/** Moves a picked team to the front, making it the main team. */
export const makeMain = (ids: string[], id: string) => (ids.includes(id) ? [id, ...ids.filter((x) => x !== id)] : ids);

/** "India", "India and RCB", "India, RCB and MI": for accessibility labels. */
export function listLine(names: string[]): string {
  if (names.length <= 1) return names[0] ?? '';
  return `${names.slice(0, -1).join(', ')} and ${names[names.length - 1]}`;
}

/** Two lists hold the same teams in the same order. */
export const sameTeams = (a: string[], b: string[]) => a.length === b.length && a.every((x, i) => x === b[i]);
