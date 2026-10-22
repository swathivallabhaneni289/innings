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

Guest accounts. `POST /v1/bootstrap` creates a guest (or resumes one when a valid `Authorization: Bearer <token>`
is sent) and returns the token. Every other route needs `Authorization: Bearer <token>`; missing/unknown → 401.
A new guest gets demo history: a few logs of Vault classics and of fixtures that finished 2+ days ago, settled
predictions on those fixtures, and follows 5 of the 8 simulated friends.

## Reads

| Method & path | Response (TS type) | Notes |
| --- | --- | --- |
| `POST /v1/bootstrap` | `BootstrapResponse` | catalog of every team and player the app can show |
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
| `POST /v1/users/{id}/follow` | – | `{following: boolean}` (toggle) |
| `PUT /v1/me` | any of `{name, bio, favouriteTeamId, avatarColour, photoUri}` | `User` |
| `PUT /v1/me/top5` | `{matchIds: string[]}` (max 5, in order) | `{top5: Match[]}` |
| `GET /healthz` | – | `ok` |

## Scoring

Bragging points, streaks, upset bonus and badges follow `docs/PLAN.md` → "Build decisions"; implemented in
`internal/domain/scoring.go`.
