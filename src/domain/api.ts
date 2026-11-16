// Response shapes of the Go API (docs/API.md). Go mirrors these exactly; see internal/domain/api.go.
import type {
  Account,
  BadgeDef,
  ChatMessage,
  Collection,
  LiveCall,
  Match,
  MatchLog,
  Player,
  Prediction,
  PredictionOutcome,
  Reply,
  Team,
  User,
  Venue,
  WatchPlace,
} from './types';

export interface PlayerAverage {
  playerId: string;
  avg: number; // out of 10
  count: number;
}

/** A log as it appears in a feed or review list. */
export interface FeedItem {
  log: MatchLog;
  user: User;
  match: Match;
  likes: number; // includes mine
  liked: boolean; // by me
  thereToo: number; // "I was there too" taps, includes mine
  iWasThere: boolean;
  replies: Reply[]; // oldest first
}

export interface FanSummary {
  ratingAvg: number; // 0–5
  ratingCount: number;
  buckets: number[]; // always 10 counts: 0.5, 1, … 5 balls
  players: PlayerAverage[]; // both XIs, best first
  fansPotmId: string | null;
  officialPotmId: string | null;
  fansDisagree: boolean;
  watch: { place: WatchPlace; share: number }[]; // sorted desc, shares sum to 1
  reviews: FeedItem[]; // popular reviews from all fans, most liked first (max 10)
  friendReviews: FeedItem[]; // friends I follow (and me)
}

export interface PredictionSplit {
  total: number; // all fans who predicted
  shares: Record<string, number>; // teamId → 0..1 (both teams always present)
  friendsTotal: number;
  friendShares: Record<string, number>; // teamId → 0..1 among friends I follow (both teams present)
  friendPicks: { user: User; teamId: string }[];
}

export interface PredictionResultRow {
  user: User;
  prediction: Prediction;
  outcome: PredictionOutcome;
}

export interface FanStats {
  logged: number;
  avgRatingGiven: number | null;
  accuracy: number | null; // 0..1
  points: number;
  streak: number;
  rank: number; // among the user + people they follow
  rankOf: number;
  badges: BadgeDef[];
  watch: { place: WatchPlace; count: number }[];
  stadiums: Venue[];
}

export interface RatedMatch {
  match: Match;
  ratingAvg: number;
  ratingCount: number;
}

export interface BoardRow {
  user: User;
  points: number;
  isMe: boolean;
}

// ---- Endpoint responses ----

/**
 * POST /v1/auth/signup {email, password}, /v1/auth/confirm/resend {email} and /v1/auth/password/forgot {email} —
 * each emails a 6-digit code. Resend and forgot answer the same whether or not the address has an account.
 */
export interface AuthCodeResponse {
  email: string; // the address, trimmed and lower-cased
  resendInSeconds: number; // wait this long before asking for another code
  expiresInSeconds: number; // how long the code works
  devCode?: string; // only when the server runs with INNINGS_DEV_CODES=1 and no mail server
}

/**
 * POST /v1/auth/confirm {email, code}, /v1/auth/signin {email, password} and /v1/auth/password/reset
 * {resetToken, password} — each signs this device in. `isNew` is true the first time (confirming the address).
 */
export interface AuthSessionResponse {
  token: string;
  me: User;
  account: Account;
  isNew: boolean;
}

/** POST /v1/auth/password/code {email, code} — the right reset code buys a token that may set a new password. */
export interface ResetCodeResponse {
  resetToken: string;
  expiresInSeconds: number;
}

/** GET /v1/handles/{handle} — is this @username free? Your own current handle counts as available. */
export interface HandleCheckResponse {
  handle: string; // lower-cased, as it would be saved
  available: boolean;
  reason: 'invalid' | 'reserved' | 'taken' | null;
  message: string | null; // human text when not available
}

/** POST /v1/contacts/match {hashes} — SHA-256 hex of contacts' E.164 numbers (max 5000). Nothing is stored. */
export interface ContactsMatchResponse {
  matches: { user: User; following: boolean }[];
}

/** POST /v1/bootstrap (needs a session) — the fan, their account and the catalog. */
export interface BootstrapResponse {
  token: string;
  me: User;
  account: Account;
  teams: Record<string, Team>;
  players: Record<string, Player>;
  serverTime: string;
}

/** GET /v1/home */
export interface HomeResponse {
  onNow: Match[]; // live first, then the rest of today's fixtures
  yourCall: Match | null;
  yourCallSplit: PredictionSplit | null;
  yourCallMine: Prediction | null;
  feed: FeedItem[];
  justFinished: (RatedMatch & { topReview: FeedItem | null })[];
  board: BoardRow[];
  unread: number; // notifications badge
}

/** GET /v1/matches?state=live|upcoming|recent&competition=ipl&team=ind */
export interface MatchesResponse {
  matches: Match[];
  fanRatings: Record<string, { avg: number; count: number }>; // completed matches only
}

/** A live call as one fan sees it. */
export interface CallView {
  call: LiveCall;
  mine: string | null; // my option, if I picked one
  counts: Record<string, number>; // option id → fans who picked it
  points: number; // mine, once it's settled
}

/** GET /v1/matches/{id} — everything the Match Page needs, in every state. */
export interface MatchPageResponse {
  calls: CallView[]; // predictions during the match, newest first
  match: Match;
  locked: boolean;
  split: PredictionSplit;
  mine: {
    prediction: Prediction | null;
    outcome: PredictionOutcome | null;
    log: MatchLog | null;
    watching: WatchPlace | null;
  };
  friendsWatching: User[]; // live (and upcoming: friends who said "I'm watching")
  fans: FanSummary | null; // completed only
  results: PredictionResultRow[]; // completed only: me + friends I follow, best first
}

/** GET /v1/vault */
export interface VaultResponse {
  collections: { collection: Collection; matches: RatedMatch[] }[];
  onThisDay: { match: Match; yearsAgo: number; exact: boolean } | null;
  top5: Match[];
  fanRatings: Record<string, { avg: number; count: number }>; // for every top5 match
  eras: string[];
}

/** GET /v1/vault/rankings?format=T20|ODI|Test|all&era=2010s|all */
export interface RankingsResponse {
  matches: RatedMatch[];
}

/** GET /v1/users/{id} ("me" works as an id) */
export interface UserResponse {
  user: User;
  isMe: boolean;
  following: boolean;
  stats: FanStats;
  logs: FeedItem[];
  top5: Match[];
}

/** GET /v1/friends */
export interface FriendsResponse {
  friends: { user: User; following: boolean }[];
}

/** GET /v1/search?q= */
export interface SearchResponse {
  matches: Match[];
  users: User[];
}

/** GET /v1/log/suggestions */
export interface LogSuggestionsResponse {
  suggested: Match[]; // live now or ended in the last 24 h
  recent: Match[]; // other completed fixtures, newest first
}

/** GET /v1/notifications */
export interface NotificationsResponse {
  items: { id: string; text: string; matchId: string | null; logId: string | null; createdAt: string }[];
}

/** GET /v1/matches/{id}/chat — the room's messages and who wrote them. */
export interface ChatResponse {
  messages: ChatMessage[]; // oldest first; with ?after=, only those from then on
  users: Record<string, User>; // every author in messages
  open: boolean; // fans can post now
  opensAt: string; // when a closed room opens ('' once it has closed, and for classics)
}
