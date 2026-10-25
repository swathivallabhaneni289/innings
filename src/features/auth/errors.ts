import { ApiError } from '@/data';

/** What to say when an auth call fails: the server's own words (they explain limits and wait times), else ours. */
export function errorText(e: unknown, fallback: string): string {
  if (e instanceof ApiError) return e.message || fallback;
  return e instanceof Error && e.message ? e.message : fallback;
}
