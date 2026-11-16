# Innings API (Go)

`go run ./cmd/api` → `http://localhost:8088`. JSON everywhere. The JSON contract is the TypeScript in
`src/domain/types.ts` (entities) and `src/domain/api.ts` (responses). Go mirrors it in `internal/domain`.

## JSON rules (both sides rely on these)

- Field names are camelCase, exactly as in the TS types.
- A TS field marked `?` is a Go pointer (or string) with `omitempty`. A TS `X | null` field is always present and may be `null`.
- **Arrays and maps are never `null`.** Initialise every slice/map (`[]T{}`, `map[K]V{}`) before encoding.
- A match's `teamIds` is a 2-element array; a fan's `teamIds` lists 1 to 5 teams, main team first, and their
  `favouriteTeamId` is the first of them. `xi` is `{teamId: [playerId × 11]}`. Timestamps are RFC 3339 UTC strings.
- Errors: `{"error": "<code>", "message": "<human text>"}` with a 4xx/5xx status.

## Auth

Fans sign up with an email address and a password. Sign-up emails a 6-digit code to that address, and the account
opens only when the code comes back, so every address on Innings has been proven by whoever reads its inbox. After
that, the email and password sign in on any number of phones, and a forgotten password is reset with another emailed
code. Every route except `/healthz` and the seven calls from `signup` to `password/reset` below needs
`Authorization: Bearer <token>`; a missing, unknown or signed-out token gets 401.

| Method & path | Body | Response |
| --- | --- | --- |
| `POST /v1/auth/signup` | `{email, password}` | `AuthCodeResponse`, and a code goes to the address. Signing up again before confirming replaces the password and sends a new code; 409 `email-taken` once a confirmed account has the address |
| `POST /v1/auth/confirm` | `{email, code}` | `AuthSessionResponse` with `isNew: true`: the account's first sign-in, on this device |
| `POST /v1/auth/confirm/resend` | `{email}` | `AuthCodeResponse`. The answer is the same whether or not an unconfirmed account has the address. |
| `POST /v1/auth/signin` | `{email, password}` | `AuthSessionResponse`; 403 `email-unconfirmed` (only with the right password) if the address was never confirmed |
| `POST /v1/auth/password/forgot` | `{email}` | `AuthCodeResponse`. The answer is the same whether or not the address has an account. |
| `POST /v1/auth/password/code` | `{email, code}` | `ResetCodeResponse`: a `resetToken` that can set a new password for the next 15 minutes |
| `POST /v1/auth/password/reset` | `{resetToken, password}` | `AuthSessionResponse`. Signs this device in and every other device out, and counts the address as confirmed, since the code proved the inbox. A weak password doesn't use the token up. |
| `POST /v1/auth/logout` | – | 204; ends this device's session only |
| `POST /v1/bootstrap` | – | `BootstrapResponse`: the fan, their `account` and the team and player catalog |
| `GET /v1/handles/{handle}` | – | `HandleCheckResponse` |
| `POST /v1/me/onboarded` | – | `{account: Account}`; 400 `profile-incomplete` until a name and @username are saved |
| `POST /v1/contacts/match` | `{hashes: string[]}` | `ContactsMatchResponse`. Each hash is the SHA-256, in lower-case hex, of a contact's email address trimmed and lower-cased; at most 5,000, or 400 `bad-hashes` |
| `PUT /v1/me/password` | `{currentPassword, newPassword}` | 204; every other device is signed out and this one stays in |
| `DELETE /v1/me` | `{password}` | 204; deletes the account and everything it made and signs out every device; 403 `bad-password` for a wrong password |

`AuthCodeResponse` is `{email, resendInSeconds, expiresInSeconds}`, plus `devCode` in dev mode. `AuthSessionResponse`
is `{token, me, account, isNew}`.

**Codes.** Six digits from `crypto/rand`, stored only as a hash. Each address has at most one live code for confirming
it and one for resetting its password. A code works for 10 minutes and allows 5 tries; every guess spends a try, even
a malformed one, and the right one uses the code up. Another code can go to the same address after 30 seconds, and at
most 5 an hour. That count is kept in the database, so a restart doesn't clear it, and it counts sends to addresses
with no account as well, so the answers give nothing away. Inside the 30 seconds, sign-up and forgot password don't
fail: they answer 200 with the wait in `resendInSeconds`, and the earlier code still works. Only `confirm/resend`
answers 429 `too-soon`. One IP address can ask for at most 20 code emails an hour across sign-up, resend and forgot
(IPv6 addresses count by their /64 block, and `X-Forwarded-For` is ignored). A send the mail server refuses still
counts toward these caps, and its code is dropped.

**Passwords.** 8 to 128 characters, with no rules about mixing letters, digits and symbols (after NIST SP 800-63B), so
a passphrase works. A new password can't be the email address, or the part before the @ when that has 4 or more
characters, one of the most-used passwords from public breach lists, or one character over and over or a short run
of the keyboard; the `weak-password` message says which. Changing to the same password is refused the same way.
Passwords are stored as PBKDF2-HMAC-SHA256 at 600,000 iterations, with a 16-byte random salt and a 32-byte key, as
`pbkdf2-sha256$600000$<salt>$<key>`. The count travels with each hash, so it can be raised later: a sign-in that finds
an older count stores a fresh hash.

**Wrong passwords.** After 10 wrong passwords for one account in 15 minutes, or 50 from one IP address in an hour,
sign-in answers 429 `too-many` with `Retry-After`, even for the right password, until the oldest failure ages out.
Checking the current password before a change or a deletion counts the same way. The right password clears the
account's count, and so does a reset. These counts, and the 20-an-hour cap per IP, live in memory, so a restart
forgets them.

| Error | When |
| --- | --- |
| 400 `bad-email` | not a single address: one @, a domain with a dot, no spaces, at most 254 characters |
| 400 `weak-password` | a new password that breaks the rules above; the message says why |
| 409 `email-taken` | sign-up with an address that a confirmed account already has |
| 401 `bad-credentials` | sign-in with a wrong password, or with an address that has no account: the same answer, in the same time |
| 403 `email-unconfirmed` | the right password for an account whose address was never confirmed; the app asks for a new code |
| 403 `bad-password` | a wrong current password when changing it or deleting the account (403, not 401, so the device stays signed in) |
| 429 `too-soon` | `confirm/resend` under 30 seconds after the last code (`Retry-After` header) |
| 429 `too-many` | 5 codes this hour for this address, 20 code emails this hour from this IP, or too many wrong passwords (`Retry-After` header) |
| 400 `bad-code` | wrong code; the message says how many tries are left |
| 410 `code-expired` | no live code for this address: expired, used, or out of tries |
| 410 `reset-expired` | the reset token is unknown, already used or older than 15 minutes |
| 503 `email-unavailable` | the server has no way to send email and isn't in dev mode |
| 502 `email-failed` | the mail server refused the email or couldn't be reached |

**Sending the emails.** With `SMTP_HOST`, `SMTP_USERNAME` and `SMTP_PASSWORD` all set, codes are emailed through that
server; Gmail with an app password works (README → Real emails). `SMTP_PORT` defaults to 587, where the server must
offer STARTTLS; on 465 the connection is TLS from the start, so the password never crosses the network in the clear.
`MAIL_FROM` defaults to `Innings <SMTP_USERNAME>`. Setting only some of the three is a startup error rather than a
quiet fall back to dev mode, which would hand codes to whoever asks. With `INNINGS_DEV_CODES=1` and no SMTP settings,
nothing is emailed: the server logs each email's address and subject, which starts with the code, and returns the
code as `devCode`, so the app can be tried with no mail account. With neither, every call that would email a code
answers 503 `email-unavailable`. A code that was really emailed is never returned in a response. `INNINGS_ENV_FILE`
names a file of `KEY=VALUE` lines to read at startup; `make api` passes `.env` along with `INNINGS_DEV_CODES=1`, so it
runs in dev mode until `.env` holds the SMTP settings and emails codes from then on. The environment wins over the
file, and the server logs the names of the keys it read, never their values.

**Usernames.** 3 to 20 characters: lower-case letters, digits, `_` and `.`, starting with a letter and not
ending with `.`. Unique across every fan (simulated fans included), and a few words such as `innings`,
`admin` and `support` are reserved.

**Privacy.** The email address appears only in the owner's `Account`, never in `User`. Sign-in, resend and forgot
password answer the same whether or not an address has an account, and a sign-in for an unknown address is checked
against a dummy hash, so it takes as long as a wrong password. Sign-up is the exception: it has to answer
`email-taken` for an address a confirmed account already has. Codes, session tokens and reset tokens are stored as
SHA-256 hashes, and passwords as above. Contact matching takes SHA-256 hashes of email addresses and keeps none of
them; anyone who already knows an address can hash it and check, so this hides the list from casual view but is not
anonymity. With SMTP the server log leaves fans' addresses out; dev mode's log lines name them. Accounts whose address is
never confirmed are cleared after a week. Accounts from the phone-number sign-in that came before have no email or
password, so they can't sign in; their rows stay until `make reset-db` empties the database.

A new account starts empty: no logs, predictions or follows. Its name and handle are blank until the fan sets
them, and it doesn't appear in search or friend lists until it is onboarded.

## Reads

| Method & path | Response (TS type) | Notes |
| --- | --- | --- |
| `GET /v1/home` | `HomeResponse` | |
| `GET /v1/matches?state=&competition=&team=` | `MatchesResponse` | `state` = live, upcoming or recent; `team` = team id (the "India" chip sends `team=ind,ind-w`) |
| `GET /v1/matches/{id}` | `MatchPageResponse` | fixtures and classics; `calls` = predictions during the match (real matches only), newest first |
| `GET /v1/vault` | `VaultResponse` | |
| `GET /v1/vault/rankings?format=&era=` | `RankingsResponse` | Vault classics + completed fixtures, by average fan rating |
| `GET /v1/users/{id}` | `UserResponse` | `me` is accepted as an id |
| `GET /v1/friends` | `FriendsResponse` | simulated friends + whether I follow them |
| `GET /v1/logs/{id}` | `FeedItem` | one stub with likes and replies |
| `GET /v1/search?q=` | `SearchResponse` | teams, venue, city, stage, competition, year. "vs" is skipped, and a team's code or one-word name ("ind", "india") matches that team and its women's side only. Current matches come first, then the past newest first: the Vault and the archive of past tournaments, whose matches carry `archived: true` (a day but no hour, and no chat). At most 100 matches |
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
| `POST /v1/matches/{id}/calls` | `CallSpec`: `{kind: "runs", playerId, target: 30 \| 50 \| 100}`, `{kind: "teamTotal", teamId, target}` or `{kind: "win", teamId}`, plus `side: "yes" \| "no"` | `CallView` (201): a fan's own call, posted into the match's chat with the asker's side picked; the same call asked again joins it. **409 `call-too-late`** once it's as good as decided (a batter within 10 of the target or out, a total past 10 overs of the innings, a winner 5 overs into the chase), 400 `bad-call`, 429 `too-many-calls` (five a match each) |
| `PUT /v1/matches/{id}/calls/{callId}` | `{optionId}` | `CallView`; change it as often as you like while the call is open; **409 `locked`** once it isn't; 400 for an option the call doesn't have |
| `POST /v1/matches/{id}/chat` | `{text}` (1–500 characters) | `ChatMessage` (201); **409 `chat-closed`** outside the room's hours; **429 `slow-down`** (with `Retry-After`) for more than one message a second |
| `POST /v1/users/{id}/follow` | – | `{following: boolean}` (toggle) |
| `PUT /v1/me` | any of `{name, handle, bio, favouriteTeamId, teamIds, avatarColour, photoUri}` | `User`; 400 `bad-handle` or 409 `handle-taken` for a handle that fails the rules above. `teamIds` sets every team the fan backs, 1 to 5 known teams with the main team first, and makes the first one `favouriteTeamId`; `favouriteTeamId` alone makes that team the main one and keeps the others. 400 `bad-team` for an unknown or repeated team, or a list outside 1–5 |
| `PUT /v1/me/top5` | `{matchIds: string[]}` (max 5, in order) | `{top5: Match[]}` |
| `GET /healthz` | – | `ok` |

## Scoring

Bragging points, streaks, upset bonus and badges follow `docs/PLAN.md` → "Build decisions"; implemented in
`internal/domain/scoring.go`.

Live calls (predictions during a match, `internal/feed/calls.go`) are worth 5 points each when right, added to the Bragging Board and the profile total. "Where will {team} finish?" opens with the first innings of a limited-overs match and locks at 10 overs (25 in an ODI); "Will {team} chase {target}?" opens with the second and locks 5 overs in (10 in an ODI); "Will {batter} make 50?" (or 100) opens when a batter on ESPN's batting card reaches 25 (75) and locks at 40 (90). A call first seen past its lock is created locked, and a call only ever moves forward: open, locked, then settled or void.
