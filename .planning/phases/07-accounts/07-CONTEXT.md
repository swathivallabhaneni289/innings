# Phase 7: Phone sign-in & onboarding - Context

**Gathered:** 2026-10-25
**Status:** Complete 2026-10-25 (see `07-01-SUMMARY.md`)

<domain>
## Phase Boundary

Replace the invisible guest account with a real sign-in, modelled on Beli: a phone number and a 6-digit text code, a fan card with a unique @username, then finding friends from contacts or inviting them. Fans can sign out and delete their account. The same flow signs up a new number and signs in an existing one on another phone.

Not in this phase: email or social sign-in, photo upload, store builds, a production SMS sender ID or India DLT registration (these need the developer's accounts).

</domain>

<decisions>
## Implementation Decisions

### Why
- **D-01:** The developer asked for "a proper sign in like how Beli has" (2026-10-25) after hearing how guest accounts work: no recovery, no second device, nothing to show a real user.
- **D-02:** Beli's pattern, as far as public sources show: phone number with a text code, a username, contacts access to find friends, and invites. Sources: the App Store and Google Play listings and Beli's terms of service. A "rank places first" onboarding step and an invite-gated waitlist were not confirmed by those sources and are not built.

### Sign-in
- **D-03:** Phone and code only, no password. The same two calls (`/v1/auth/start`, `/v1/auth/verify`) sign up and sign in. Apple and Google sign-in are not added unless asked.
- **D-04:** Codes: 6 digits, 10 minutes, 5 tries, a resend after 30 seconds, at most 5 codes an hour per number, 20 starts an hour per IP. The start answer is the same whether or not the number has an account.
- **D-05:** SMS through Twilio Verify when its three settings exist. `INNINGS_DEV_OTP=1` (set by `make api`) makes and shows the code locally. With neither, sign-in is refused (503) so an unconfigured server never hands out codes. India needs DLT-compliant sending for any provider; whether Twilio Verify covers it is unconfirmed, so the sender sits behind an interface.
- **D-06:** Session tokens are random, kept in SecureStore on the phone and stored only as SHA-256 hashes on the server. Each device has its own session.

### Accounts
- **D-07:** New accounts start empty: no demo logs, predictions or follows. The old guests' demo history is gone for new fans, so every screen needs a proper zero state.
- **D-08:** Name and @username are blank until the fan sets them. An account is onboarded once both are saved and the friends step is finished. Accounts that aren't onboarded (including old guests) don't appear in search, friend lists or contact matches.
- **D-09:** Usernames are 3–20 characters (`a–z`, `0–9`, `_`, `.`), start with a letter, are unique against real and simulated fans, and skip a few reserved words.
- **D-10:** Old guest tokens stop working; those fans see the welcome screen. Their rows stay in the database, hidden.
- **D-11:** Delete account removes the user and everything they made, including other fans' likes and replies on their stubs, in one transaction (App Store guideline: apps that let you create an account must let you delete it).

### Friends
- **D-12:** Contacts are read on the phone, normalised to E.164 with libphonenumber-js (the default country is the fan's own number), hashed with SHA-256 and matched on the server. Nothing is stored. An unsalted hash of a phone number can be reversed by trying every number; the docs say so.
- **D-13:** Simulated fans have no phone numbers, so locally contacts never match. The no-match and permission-denied states are designed screens with "Invite friends" and "Open Settings". "Fans to follow" lists the simulated fans so a new feed fills up.

### App structure
- **D-14:** Signed-in screens move into an `(app)` route group. The root layout guards `(auth)`, `onboarding` and `(app)` with `Stack.Protected` (SDK 57; `redirectTo` is SDK 58+ and is not used). URLs don't change.
- **D-15:** New native modules (expo-secure-store, expo-contacts, expo-crypto) mean a new development build on each phone.

### Claude's Discretion
- The look of the welcome ticket, the code cells and the live fan-card preview, within Matchday Ticket.

</decisions>

<canonical_refs>
## Canonical References

- `docs/API.md` → "Auth": endpoints, limits, error codes, usernames, privacy
- `src/domain/types.ts` (`Account`), `src/domain/api.ts` (`AuthStartResponse`, `AuthVerifyResponse`, `HandleCheckResponse`, `ContactsMatchResponse`)
- Twilio Verify API: `https://www.twilio.com/docs/verify/api/verification`, `.../verification-check`
- Expo SDK 57: `https://docs.expo.dev/versions/v57.0.0/sdk/securestore/`, `.../contacts/`, `.../crypto/`; protected routes `https://docs.expo.dev/router/advanced/protected/`

</canonical_refs>

<deferred>
## Deferred Ideas

- Sign in with Apple and Google
- Twilio (or another provider) account, sender ID and India DLT registration
- Invite links that open the app (needs a domain and universal links)
- Moving an old guest's history into a new account

</deferred>

---

*Phase: 7-Phone sign-in & onboarding*
*Context gathered: 2026-10-25*
