import { looksLikeEmail, normEmail } from '@/features/auth/validate';

/**
 * The distinct, well-formed email addresses in a list of contacts, trimmed and lower-cased as the server keeps them,
 * leaving out the fan's own. Pure, so it's tested without an address book.
 */
export function contactEmails(contacts: { emails?: { address?: string }[] }[], ownEmail?: string): string[] {
  const own = ownEmail ? normEmail(ownEmail) : '';
  const found = new Set<string>();
  for (const c of contacts) {
    for (const e of c.emails ?? []) {
      const address = e.address ? normEmail(e.address) : '';
      if (address && address !== own && looksLikeEmail(address)) found.add(address);
    }
  }
  return [...found];
}
