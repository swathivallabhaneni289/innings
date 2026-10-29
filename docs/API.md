# Innings API (Go)

`go run ./cmd/api` → `http://localhost:8088`. JSON everywhere. The JSON contract is the TypeScript in
`src/domain/types.ts` (entities) and `src/domain/api.ts` (responses). Go mirrors it in `internal/domain`.

## JSON rules (both sides rely on these)

- Field names are camelCase, exactly as in the TS types.
- A TS field marked `?` is a Go pointer (or string) with `omitempty`. A TS `X | null` field is always present and may be `null`.
- **Arrays and maps are never `null`.** Initialise every slice/map (`[]T{}`, `map[K]V{}`) before encoding.
- `teamIds` is a 2-element array. `xi` is `{teamId: [playerId × 11]}`. Timestamps are RFC 3339 UTC strings.
- Errors: `{"error": "<code>", "message": "<human text>"}` with a 4xx/5xx status.

## Auth

Fans sign in with a phone number and a 6-digit text code. The same two calls sign up a new number and sign in
an existing one, on any number of devices. Every route except `/v1/auth/start`, `/v1/auth/verify` and
`/healthz` needs `Authorization: Bearer <token>`; a missing, unknown or signed-out token gets 401.

| Method & path | Body | Response |
| --- | --- | --- |
| `POST /v1/auth/start` | `{phone}` (E.164, e.g. `+919876543210`) | `AuthStartResponse`. The answer is the same whether or not the number has an account. |
| `POST /v1/auth/verify` | `{phone, code}` | `AuthVerifyResponse`. Creates the account the first time a number verifies (`isNew: true`). |
| `POST /v1/auth/logout` | – | 204; ends this device's session only |
| `POST /v1/bootstrap` | – | `BootstrapResponse`: the fan, their `account` and the team and player catalog |
| `GET /v1/handles/{handle}` | – | `HandleCheckResponse` |
| `POST /v1/me/onboarded` | – | `{account: Account}`; 400 `profile-incomplete` until a name and @username are saved |
| `POST /v1/contacts/match` | `{hashes: string[]}` | `ContactsMatchResponse` |
| `DELETE /v1/me` | – | 204; deletes the account and everything it made |

**Codes.** Six digits, valid for 10 minutes, 5 tries each. A new code can be asked for after 30 seconds, at most
5 times an hour per number; one IP address can start at most 20 sign-ins an hour.

| Error | When |
| --- | --- |
| 400 `bad-phone` | not a full E.164 mobile number |
| 429 `too-soon` | a code was sent under 30 seconds ago (`Retry-After` header) |
| 429 `too-many` | 5 codes this hour for this number, or 20 starts this hour from this IP |
| 503 `sms-unavailable` | the server has no SMS provider and isn't in dev mode |
| 502 `sms-failed` | the SMS provider refused or failed |
| 400 `bad-code` | wrong code; the message says how many tries are left |
| 410 `code-expired` | no live code for this number: expired, used, or out of tries |

**SMS providers.** With `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN` and `TWILIO_VERIFY_SERVICE_SID` set, codes go
through Twilio Verify. With `INNINGS_DEV_OTP=1` and no Twilio settings (what `make api` does), the server makes
the code itself, writes it to its log and returns it as `devCode`, so the app can be tried without a phone
network. With neither, `/v1/auth/start` answers 503. A provider's code is never returned in a response.

**Usernames.** 3 to 20 characters: lower-case letters, digits, `_` and `.`, starting with a letter and not
ending with `.`. Unique across every fan (simulated fans included), and a few words such as `innings`,
`admin` and `support` are reserved.

**Privacy.** The phone number appears only in the owner's `Account`, never in `User`. Contact matching takes
SHA-256 hashes and keeps none of them; note that an unsalted hash of a phone number can be reversed by
trying every number, so this hides the list from casual view but is not anonymity. Session tokens are stored
as SHA-256 hashes.

A new account starts empty: no logs, predictions or follows. Its name and handle are blank until the fan sets
them, and it doesn't appear in search or friend lists until it is onboarded.

## Reads

| Method & path | Response (TS type) | Notes |
| --- | --- | --- |
| `GET /v1/home` | `HomeResponse` | |
| `GET /v1/matches?state=&competition=&team=` | `MatchesResponse` | `state` = live, upcoming or recent; `team` = team id (the "India" chip sends `team=ind,ind-w`) |
| `GET /v1/matches/{id}` | `MatchPageResponse` | fixtures and classics |
| `GET /v1/vault` | `VaultResponse` | |
| `GET /v1/vault/rankings?format=&era=` | `RankingsResponse` | Vault classics + completed fixtures, by average fan rating |
| `GET /v1/users/{id}` | `UserResponse` | `me` is accepted as an id |
| `GET /v1/friends` | `FriendsResponse` | simulated friends + whether I follow them |
| `GET /v1/logs/{id}` | `FeedItem` | one stub with likes and replies |
| `GET /v1/search?q=` | `SearchResponse` | teams, venue, city, stage, competition, year |
| `GET /v1/log/suggestions` | `LogSuggestionsResponse` | |
| `GET /v1/notifications` | `NotificationsResponse` | |
| `GET /v1/matches/{id}/chat?after=` | `ChatResponse` | the latest 200 messages, oldest first; `after` = a message's `createdAt` returns from then on (inclusive, so drop the ones you have by id). Open from 24 hours before the first ball until a day after the match's usual end; classics have no chat |

## Writes

| Method & path | Body | Response |
| --- | --- | --- |
| `POST /v1/logs` | `NewLog` = `MatchLog` minus `id`, `userId`, `createdAt`, `stubNo` | `FeedItem` (201). Re-logging the same match updates it and keeps the stub number. |
| `DELETE /v1/logs/{id}` | – | 204 |
| `PUT /v1/matches/{id}/prediction` | `{winnerId, topRunScorerId, topWicketTakerId}` | `{prediction: Prediction}`; **409 `locked`** once the first ball is bowled; 400 if a pick isn't in that match |
| `PUT /v1/matches/{id}/watching` | `{place: WatchPlace \| null}` | `{watching: WatchPlace \| null}` |
| `POST /v1/logs/{id}/like` | – | `{liked: boolean, likes: number}` (toggle) |
| `POST /v1/logs/{id}/there-too` | – | `{iWasThere: boolean, thereToo: number}` (toggle) |
| `POST /v1/logs/{id}/replies` | `{text}` | `Reply` (201) |
| `POST /v1/matches/{id}/chat` | `{text}` (1–500 characters) | `ChatMessage` (201); **409 `chat-closed`** outside the room's hours; **429 `slow-down`** (with `Retry-After`) for more than one message a second |
| `POST /v1/users/{id}/follow` | – | `{following: boolean}` (toggle) |
| `PUT /v1/me` | any of `{name, handle, bio, favouriteTeamId, avatarColour, photoUri}` | `User`; 400 `bad-handle` or 409 `handle-taken` for a handle that fails the rules above |
| `PUT /v1/me/top5` | `{matchIds: string[]}` (max 5, in order) | `{top5: Match[]}` |
| `GET /healthz` | – | `ok` |

## Scoring

Bragging points, streaks, upset bonus and badges follow `docs/PLAN.md` → "Build decisions"; implemented in
`internal/domain/scoring.go`.
