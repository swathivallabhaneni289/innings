// Data API for screens: react-query hooks over the Go API (docs/API.md). Screens import from '@/data' only.
// Every read hook returns react-query's result: { data, isLoading, error, refetch, isRefetching }.
// Render <QueryState q={...}> (src/components/States.tsx) for loading/error, never treat "no data yet" as "empty".
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useState } from 'react';

import type {
  AuthStartResponse,
  AuthVerifyResponse,
  BootstrapResponse,
  ContactsMatchResponse,
  FeedItem,
  FriendsResponse,
  HandleCheckResponse,
  HomeResponse,
  LogSuggestionsResponse,
  MatchesResponse,
  MatchPageResponse,
  NotificationsResponse,
  RankingsResponse,
  SearchResponse,
  UserResponse,
  VaultResponse,
} from '@/domain/api';
import type { Account, Format, Match, MatchLog, Player, Prediction, Reply, Team, User, WatchPlace } from '@/domain/types';
import { api, ApiError, catalog } from './client';
import { queryClient } from './queryClient';
import { endSession, markReady, onSessionEnd, saveToken, useSession } from './session';
import { usePrefs } from './store';

export { ApiError } from './client';
export { queryClient } from './queryClient';
export { loadSession, useSession, type SessionStatus } from './session';
export { usePrefs } from './store';
export type * from '@/domain/api';

// Whoever signs in next starts clean: no cached screens, no catalog, no "seen" mark from the last fan.
onSessionEnd(() => {
  queryClient.clear();
  catalog.teams = {};
  catalog.players = {};
  usePrefs.setState({ notificationsSeenAt: null });
});

// ---------- catalog (synchronous after bootstrap) ----------

export function getTeam(id: string): Team {
  return catalog.teams[id] ?? { id, name: id.toUpperCase(), short: id.toUpperCase(), colour: '#1C2333', gender: 'men', kind: 'international' };
}

export function getPlayer(id: string | null | undefined): Player | undefined {
  return id ? catalog.players[id] : undefined;
}

export const playerName = (id: string | null | undefined) => getPlayer(id)?.name ?? '—';

/** "IND v AUS" */
export const matchTitle = (m: Pick<Match, 'teamIds'>) => `${getTeam(m.teamIds[0]).short} v ${getTeam(m.teamIds[1]).short}`;

/** Both XIs as Player objects, team order preserved. */
export function squads(m: Match): { team: Team; players: Player[] }[] {
  return m.teamIds.map((id) => ({
    team: getTeam(id),
    players: (m.xi?.[id] ?? []).map((pid) => catalog.players[pid]).filter(Boolean) as Player[],
  }));
}

// ---------- clock ----------

/** Wall clock that re-renders on an interval (default 15 s) so countdowns, locks and live states move on their own. */
export function useNow(intervalMs = 15_000): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);
  return now;
}

// ---------- session ----------

/** Loads the catalog, the fan and their account into the cache. Needs a session token. */
async function bootstrap(): Promise<BootstrapResponse> {
  // The response still carries `token`; it is ignored. The only token is the one from /v1/auth/verify.
  const res = await api<BootstrapResponse>('/v1/bootstrap', { method: 'POST' });
  catalog.teams = res.teams;
  catalog.players = res.players;
  queryClient.setQueryData(['me'], res.me);
  queryClient.setQueryData(['account'], res.account);
  return res;
}

/**
 * Root layout only: bootstraps a session found in SecureStore at launch. Sign-in bootstraps on its own
 * (see `signIn`), so this only runs while `status` is signedIn and `ready` is false.
 */
export function useBootstrap() {
  const [error, setError] = useState<Error | null>(null);
  const [running, setRunning] = useState(false);
  const run = useCallback(async () => {
    setError(null);
    setRunning(true);
    try {
      await bootstrap();
      markReady();
    } catch (e) {
      // A 401 has already ended the session (back to sign-in); anything else is shown with a retry.
      if (!(e instanceof ApiError && e.status === 401)) setError(e as Error);
    } finally {
      setRunning(false);
    }
  }, []);
  return { error, running, run };
}

/** Texts a sign-in code to an E.164 number. */
export const startSignIn = (phone: string) => api<AuthStartResponse>('/v1/auth/start', { method: 'POST', body: { phone }, token: null });

/**
 * Checks the code, stores the session and bootstraps it. Resolves once everything is in the cache, then the
 * root guards flip to onboarding or the app. Returns the account so the caller can route by `onboarded`.
 */
export async function signIn(phone: string, code: string): Promise<Account> {
  const res = await api<AuthVerifyResponse>('/v1/auth/verify', { method: 'POST', body: { phone, code }, token: null });
  await saveToken(res.token);
  try {
    const boot = await bootstrap();
    markReady();
    return boot.account;
  } catch (e) {
    endSession(); // half signed in helps nobody: forget the token and let them try again
    throw e;
  }
}

/** Signs out of this device: back to Welcome at once; the server is told best-effort, in the background. */
export function signOut() {
  const token = useSession.getState().token;
  endSession();
  if (token) api<void>('/v1/auth/logout', { method: 'POST', token }).catch(() => {});
}

/** Deletes the account and everything it made, then ends the session. */
export async function deleteAccount() {
  await api<void>('/v1/me', { method: 'DELETE' });
  endSession();
}

/** The signed-in fan's private account (phone, onboarded). Seeded by bootstrap; never refetched on its own. */
export function useAccount() {
  const signedIn = useSession((s) => s.status === 'signedIn' && s.ready);
  return useQuery({
    queryKey: ['account'],
    queryFn: () => bootstrap().then((r) => r.account),
    enabled: signedIn,
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

/** The signed-in fan (set by bootstrap, refreshed by updateMe). */
export function useMe() {
  return useQuery({ queryKey: ['me'], queryFn: () => api<UserResponse>('/v1/users/me').then((r) => r.user), staleTime: Infinity });
}

/** Is this @username free? Lower-cased by the server; your own current handle counts as available. */
export const checkHandle = (handle: string) => api<HandleCheckResponse>(`/v1/handles/${encodeURIComponent(handle)}`);

/** SHA-256 hashes of contacts' E.164 numbers → fans on Innings. The server takes at most 5000 per call. */
export async function matchContacts(hashes: string[]): Promise<ContactsMatchResponse['matches']> {
  const found = new Map<string, ContactsMatchResponse['matches'][number]>();
  for (let i = 0; i < hashes.length; i += 5000) {
    const res = await api<ContactsMatchResponse>('/v1/contacts/match', { method: 'POST', body: { hashes: hashes.slice(i, i + 5000) } });
    for (const m of res.matches) found.set(m.user.id, m);
  }
  return [...found.values()];
}

/** Finishes sign-up. The new account lands in the cache, and the root guard opens the app. */
export async function finishOnboarding(): Promise<Account> {
  const res = await api<{ account: Account }>('/v1/me/onboarded', { method: 'POST' });
  queryClient.setQueryData(['account'], res.account);
  return res.account;
}

// ---------- reads ----------

export const useHome = () => useQuery({ queryKey: ['home'], queryFn: () => api<HomeResponse>('/v1/home'), refetchInterval: 30_000 });

export interface MatchFilter {
  state?: 'live' | 'upcoming' | 'recent';
  competition?: string;
  team?: string; // comma-separated team ids
}

export function useMatches(f: MatchFilter = {}) {
  const qs = new URLSearchParams(Object.entries(f).filter(([, v]) => v) as [string, string][]).toString();
  return useQuery({
    queryKey: ['matches', qs],
    queryFn: () => api<MatchesResponse>(`/v1/matches${qs ? `?${qs}` : ''}`),
    refetchInterval: 30_000,
    placeholderData: keepPreviousData,
  });
}

/** Everything the Match Page needs. Refreshes every 20 s for live scores. */
export const useMatchPage = (id: string | undefined) =>
  useQuery({ queryKey: ['match', id], queryFn: () => api<MatchPageResponse>(`/v1/matches/${id}`), enabled: !!id, refetchInterval: 20_000 });

export const useVault = () => useQuery({ queryKey: ['vault'], queryFn: () => api<VaultResponse>('/v1/vault') });

export const useRankings = (f: { format: Format | 'all'; era: string | 'all' }) =>
  useQuery({
    queryKey: ['rankings', f.format, f.era],
    queryFn: () => api<RankingsResponse>(`/v1/vault/rankings?format=${encodeURIComponent(f.format)}&era=${encodeURIComponent(f.era)}`),
  });

/** A fan's profile ("me" for the signed-in fan): stats, stubs, top 5. */
export const useUserProfile = (id: string | undefined) =>
  useQuery({ queryKey: ['user', id], queryFn: () => api<UserResponse>(`/v1/users/${id}`), enabled: !!id });

export const useFriends = () => useQuery({ queryKey: ['friends'], queryFn: () => api<FriendsResponse>('/v1/friends') });

export const useLogItem = (id: string | undefined) =>
  useQuery({ queryKey: ['log', id], queryFn: () => api<FeedItem>(`/v1/logs/${id}`), enabled: !!id });

export const useSearch = (q: string) =>
  useQuery({
    queryKey: ['search', q.trim()],
    queryFn: () => api<SearchResponse>(`/v1/search?q=${encodeURIComponent(q.trim())}`),
    enabled: q.trim().length >= 2,
    placeholderData: keepPreviousData,
  });

export const useLogSuggestions = () => useQuery({ queryKey: ['log-suggestions'], queryFn: () => api<LogSuggestionsResponse>('/v1/log/suggestions') });

export const useNotifications = () =>
  useQuery({ queryKey: ['notifications'], queryFn: () => api<NotificationsResponse>('/v1/notifications'), refetchInterval: 60_000 });

/** Notifications newer than the last visit to the notifications screen. */
export function useUnreadCount(): number {
  const seenAt = usePrefs((s) => s.notificationsSeenAt);
  const q = useNotifications();
  const since = seenAt ? Date.parse(seenAt) : 0;
  return (q.data?.items ?? []).filter((n) => Date.parse(n.createdAt) > since).length;
}

// ---------- writes ----------

export type NewLog = Omit<MatchLog, 'id' | 'userId' | 'createdAt' | 'stubNo'>;
export type Picks = Pick<Prediction, 'winnerId' | 'topRunScorerId' | 'topWicketTakerId'>;

/**
 * Mutations. Each returns a promise and refreshes affected screens afterwards.
 * `predict` rejects with ApiError status 409 (code "locked") once the first ball is bowled.
 */
export function useActions() {
  const qc = useQueryClient();
  // 'me' and 'account' are set by hand (updateMe, finishOnboarding); everything else refetches after a write.
  const refresh = useCallback(() => qc.invalidateQueries({ predicate: (q) => q.queryKey[0] !== 'me' && q.queryKey[0] !== 'account' }), [qc]);
  const run = useMutation({ mutationFn: (fn: () => Promise<unknown>) => fn(), onSuccess: refresh });
  const exec = useCallback(<T,>(fn: () => Promise<T>) => run.mutateAsync(fn) as Promise<T>, [run]);
  return {
    pending: run.isPending,
    saveLog: (body: NewLog) => exec(() => api<FeedItem>('/v1/logs', { method: 'POST', body })),
    deleteLog: (id: string) => exec(() => api<void>(`/v1/logs/${id}`, { method: 'DELETE' })),
    predict: (matchId: string, picks: Picks) =>
      exec(() => api<{ prediction: Prediction }>(`/v1/matches/${matchId}/prediction`, { method: 'PUT', body: picks })),
    setWatching: (matchId: string, place: WatchPlace | null) =>
      exec(() => api<{ watching: WatchPlace | null }>(`/v1/matches/${matchId}/watching`, { method: 'PUT', body: { place } })),
    toggleLike: (logId: string) => exec(() => api<{ liked: boolean; likes: number }>(`/v1/logs/${logId}/like`, { method: 'POST' })),
    toggleThereToo: (logId: string) =>
      exec(() => api<{ iWasThere: boolean; thereToo: number }>(`/v1/logs/${logId}/there-too`, { method: 'POST' })),
    addReply: (logId: string, text: string) => exec(() => api<Reply>(`/v1/logs/${logId}/replies`, { method: 'POST', body: { text } })),
    toggleFollow: (userId: string) => exec(() => api<{ following: boolean }>(`/v1/users/${userId}/follow`, { method: 'POST' })),
    /** Follows each fan once. The endpoint toggles, so pass only fans not followed yet. */
    followAll: (userIds: string[]) =>
      exec(() => Promise.all(userIds.map((id) => api<{ following: boolean }>(`/v1/users/${id}/follow`, { method: 'POST' })))),
    /** Rejects with ApiError 409 "handle-taken" or 400 "bad-handle" for a handle that can't be had. */
    updateMe: (patch: Partial<Omit<User, 'id'>>) =>
      exec(async () => {
        const me = await api<User>('/v1/me', { method: 'PUT', body: patch });
        qc.setQueryData(['me'], me);
        return me;
      }),
    setTop5: (matchIds: string[]) => exec(() => api<{ top5: Match[] }>('/v1/me/top5', { method: 'PUT', body: { matchIds } })),
  };
}

export const isLockedError = (e: unknown) => e instanceof ApiError && e.status === 409;
