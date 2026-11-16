// What the sign-up screens can check before asking the server, which has the final word (internal/auth). Pure
// functions, so they're tested without a phone.

export const PASSWORD_MIN = 8;
export const PASSWORD_MAX = 128;

/** The address as the server keeps it: trimmed and lower-cased. */
export const normEmail = (raw: string) => raw.trim().toLowerCase();

/** Looks like an address mail can reach: something@domain.tld, no spaces, at most 254 characters. */
export function looksLikeEmail(raw: string): boolean {
  const e = normEmail(raw);
  return e.length <= 254 && /^[^\s@]+@[^\s@.]+(\.[^\s@.]+)+$/.test(e);
}

/** Characters as people count them, so an emoji or a Hindi letter counts once, as on the server. */
export const passwordLength = (p: string) => [...p].length;

/**
 * The length rule, as a fan sees it filling up: how many characters are still to go, or that it's long enough.
 * Other rules (not your email, not a common password) come back from the server in words.
 */
export function lengthStatus(p: string): { ok: boolean; toGo: number; tooLong: boolean } {
  const n = passwordLength(p);
  return { ok: n >= PASSWORD_MIN && n <= PASSWORD_MAX, toGo: Math.max(0, PASSWORD_MIN - n), tooLong: n > PASSWORD_MAX };
}
