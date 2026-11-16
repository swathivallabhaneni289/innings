// Finding friends in the address book. Every email address is trimmed, lower-cased and deduped, then sent only as a
// SHA-256 hash; the server matches the hashes against fans' confirmed addresses and keeps none of them.
import { Contact, ContactField, getPermissionsAsync, requestPermissionsAsync } from 'expo-contacts';
import * as Crypto from 'expo-crypto';

import { matchContacts } from '@/data';
import type { User } from '@/domain/types';
import { contactEmails } from './contactEmails';

export type ContactsResult =
  | { kind: 'denied'; canAskAgain: boolean }
  | { kind: 'done'; matches: { user: User; following: boolean }[]; checked: number };

/** Hashes in small parallel batches, so a big address book doesn't flood the native bridge at once. */
async function sha256All(values: string[]): Promise<string[]> {
  const out: string[] = [];
  for (let i = 0; i < values.length; i += 250) {
    const batch = values.slice(i, i + 250).map((v) => Crypto.digestStringAsync(Crypto.CryptoDigestAlgorithm.SHA256, v));
    out.push(...(await Promise.all(batch)));
  }
  return out;
}

/** Asks for contacts (once), reads every email address, and returns the fans on Innings among them. */
export async function findFriendsInContacts(ownEmail?: string): Promise<ContactsResult> {
  let perm = await getPermissionsAsync();
  if (perm.status !== 'granted' && perm.canAskAgain) perm = await requestPermissionsAsync();
  if (perm.status !== 'granted') return { kind: 'denied', canAskAgain: perm.canAskAgain };

  const emails = contactEmails(await Contact.getAllDetails([ContactField.FULL_NAME, ContactField.EMAILS]), ownEmail);
  if (emails.length === 0) return { kind: 'done', matches: [], checked: 0 };
  const matches = await matchContacts(await sha256All(emails));
  return { kind: 'done', matches, checked: emails.length };
}
