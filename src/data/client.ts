import Constants from 'expo-constants';
import { Platform } from 'react-native';

import type { Player, Team } from '@/domain/types';
import { endSession, useSession } from './session';

/**
 * The Go API. On a phone it is the Metro host's LAN address; on the iOS Simulator, localhost; on the Android
 * Emulator, 10.0.2.2 (the emulator's alias for the computer it runs on).
 */
function resolveApiUrl() {
  if (process.env.EXPO_PUBLIC_API_URL) return process.env.EXPO_PUBLIC_API_URL;
  const host = Constants.expoConfig?.hostUri?.split(':')[0];
  const local = !host || host === 'localhost' || host === '127.0.0.1';
  if (local) return `http://${Platform.OS === 'android' ? '10.0.2.2' : '127.0.0.1'}:8088`;
  return `http://${host}:8088`;
}

export const API_URL = resolveApiUrl();

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}

/**
 * One call to the Go API. Sends the session token (or `opts.token`, for a call that must outlive the session,
 * like logout). A 401 on a request that carried the current token ends the session, which sends the fan back
 * to sign-in; a 401 without a token (sign-in screens) is just an error.
 */
export async function api<T>(path: string, opts: { method?: string; body?: unknown; token?: string | null } = {}): Promise<T> {
  const token = opts.token !== undefined ? opts.token : useSession.getState().token;
  let res: Response;
  try {
    res = await fetch(API_URL + path, {
      method: opts.method ?? 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
    });
  } catch {
    throw new ApiError(0, 'offline', `Can't reach the Innings server at ${API_URL}. Is it running? (make api)`);
  }
  if (res.status === 204) return undefined as T;
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 401 && token && token === useSession.getState().token) endSession();
    throw new ApiError(res.status, json.error ?? 'error', json.message ?? `Request failed (${res.status})`);
  }
  return json as T;
}

/** Teams and players from bootstrap, kept in memory so lookups stay synchronous. */
export const catalog: { teams: Record<string, Team>; players: Record<string, Player> } = { teams: {}, players: {} };
