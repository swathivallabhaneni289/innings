// The signed-in session. The token lives in the device keychain/keystore (expo-secure-store), never in the
// AsyncStorage prefs. `status` drives the root layout's guards: loading → (splash), signedOut → (auth),
// signedIn → onboarding or (app).
import * as SecureStore from 'expo-secure-store';
import { create } from 'zustand';

const KEY = 'innings.session.token';
const OPTS: SecureStore.SecureStoreOptions = { keychainAccessible: SecureStore.AFTER_FIRST_UNLOCK };

export type SessionStatus = 'loading' | 'signedOut' | 'signedIn';

interface SessionState {
  status: SessionStatus;
  token: string | null;
  /** True once bootstrap has loaded the catalog, `me` and `account` for this token. */
  ready: boolean;
}

export const useSession = create<SessionState>(() => ({ status: 'loading', token: null, ready: false }));

/** Reads the stored token once at launch. A token found means signed in; bootstrap still has to run. */
export async function loadSession() {
  if (useSession.getState().status !== 'loading') return;
  let token: string | null = null;
  try {
    token = await SecureStore.getItemAsync(KEY, OPTS);
  } catch {
    token = null; // unreadable keychain entry (e.g. restored from another device): sign in again
  }
  if (useSession.getState().status !== 'loading') return;
  useSession.setState({ token, status: token ? 'signedIn' : 'signedOut', ready: false });
}

/**
 * Stores a fresh token from /v1/auth/verify. The status stays as it is until the caller has bootstrapped
 * (see `markReady`), so the guards flip once, with everything a screen needs already in the cache.
 */
export async function saveToken(token: string) {
  await SecureStore.setItemAsync(KEY, token, OPTS);
  useSession.setState({ token });
}

/** Bootstrap finished for the current token: open the signed-in routes. */
export function markReady() {
  useSession.setState({ status: 'signedIn', ready: true });
}

const endListeners = new Set<() => void>();

/** Runs after every sign-out, 401 and deleted account (the data layer clears its caches here). */
export function onSessionEnd(fn: () => void) {
  endListeners.add(fn);
  return () => endListeners.delete(fn);
}

/**
 * Ends the session on this device: flips to signed out first (the guards unmount every signed-in screen and
 * their polling stops), then forgets the token and lets listeners clear caches. Safe to call twice.
 */
export function endSession() {
  const { status, token } = useSession.getState();
  if (status === 'signedOut' && !token) return;
  useSession.setState({ status: 'signedOut', token: null, ready: false });
  SecureStore.deleteItemAsync(KEY, OPTS).catch(() => {});
  for (const fn of endListeners) fn();
}
