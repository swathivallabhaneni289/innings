// Finding friends in the address book. Every number is normalised to E.164 (src/lib/phone.ts), deduped, and
// sent only as a SHA-256 hash; the server matches the hashes and keeps none of them.
import { Contact, ContactField, getPermissionsAsync, requestPermissionsAsync } from 'expo-contacts';
import * as Crypto from 'expo-crypto';

import { matchContacts } from '@/data';
import type { User } from '@/domain/types';
import { normalizeContactNumber, type CountryCode } from '@/lib/phone';

export type ContactsResult =
  | { kind: 'denied'; canAskAgain: boolean }
  | { kind: 'done'; matches: { user: User; following: boolean }[]; checked: number };

/** Hashes in small parallel batches, so a big address book doesn't flood the native bridge at once. */
async function sha256All(numbers: string[]): Promise<string[]> {
  const out: string[] = [];
  for (let i = 0; i < numbers.length; i += 250) {
    const batch = numbers.slice(i, i + 250).map((n) => Crypto.digestStringAsync(Crypto.CryptoDigestAlgorithm.SHA256, n));
    out.push(...(await Promise.all(batch)));
  }
  return out;
}

/**
 * Asks for contacts (once), reads every number, and returns the fans on Innings among them. Numbers saved
 * without a country code are read as `defaultCountry` (the fan's own); the fan's own number is left out.
 */
export async function findFriendsInContacts(defaultCountry: CountryCode, ownPhone?: string): Promise<ContactsResult> {
  let perm = await getPermissionsAsync();
  if (perm.status !== 'granted' && perm.canAskAgain) perm = await requestPermissionsAsync();
  if (perm.status !== 'granted') return { kind: 'denied', canAskAgain: perm.canAskAgain };

  const contacts = await Contact.getAllDetails([ContactField.FULL_NAME, ContactField.PHONES]);
  const numbers = new Set<string>();
  for (const c of contacts) {
    for (const p of c.phones ?? []) {
      const e164 = p.number ? normalizeContactNumber(p.number, defaultCountry) : null;
      if (e164 && e164 !== ownPhone) numbers.add(e164);
    }
  }
  if (numbers.size === 0) return { kind: 'done', matches: [], checked: 0 };
  const matches = await matchContacts(await sha256All([...numbers]));
  return { kind: 'done', matches, checked: numbers.size };
}
