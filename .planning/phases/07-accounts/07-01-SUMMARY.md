---
phase: 07-accounts
plan: 01
subsystem: auth
tags: [phone-sign-in, otp, sessions, twilio-verify, onboarding, contacts, secure-store, expo-router-protected]

requires:
  - phase: 06-design-release
    provides: the Matchday Ticket design system and every screen
provides:
  - Phone number and text-code sign-in, one flow for new and returning fans
  - Fan card onboarding with a unique @username, and finding friends from contacts
  - Sign out and delete account
affects: [v2 work, store builds]

tech-stack:
  added: [expo-secure-store 57.0.4, expo-contacts 57.0.8, expo-crypto 57.0.4, libphonenumber-js 1.13.15]
  patterns: ["Stack.Protected guards for (auth), onboarding and (app)", "a Verifier interface with dev and Twilio implementations", "KeyboardAware: KeyboardAvoidingView on iOS, keyboard-height padding on Android"]

key-files:
  created: [internal/store/migrations/0002_accounts.sql, internal/store/accounts.go, internal/auth/otp.go, internal/auth/twilio.go, internal/auth/ratelimit.go, internal/httpapi/account.go, internal/domain/handle.go, src/data/session.ts, "src/app/(auth)/", src/app/onboarding/, src/features/auth/, src/features/onboarding/, src/lib/phone.ts, src/components/KeyboardAware.tsx]
  modified: [internal/auth/auth.go, internal/httpapi/server.go, internal/httpapi/writes.go, cmd/api/main.go, Makefile, src/app/_layout.tsx, src/data/client.ts, src/data/index.ts, src/features/profile/SettingsForm.tsx, "src/app/(app)/settings.tsx", "src/app/(app)/friends.tsx", src/components/ui.tsx, app.json]

requirements-completed: [ACCT-01, ACCT-02, ACCT-03, ACCT-04, ACCT-05, ACCT-06, ACCT-07, ACCT-08]

completed: 2026-10-25
status: complete
---

# Phase 7 Plan 1: Phone Sign-in & Onboarding Summary

**Fans now sign in the way they do on Beli: a phone number, a 6-digit code, a fan card with a unique @username, then their people from contacts. The guest accounts are gone; sign out and delete account sit in Settings.**

## Performance

- **Completed:** 2026-10-25
- **Duration:** one afternoon. The contract was written first, then the server and the app were built in parallel against it, then the two were walked together on the Android emulator.
- **Commits:** recorded in the Git history alongside this file.

## Accomplishments

- **Contract** (`src/domain/`, `internal/domain/`, `docs/API.md`): `Account`, the four auth responses and `BootstrapResponse.account`. The Auth section of the API doc covers every limit, error code and privacy rule.
- **Server** (`internal/auth/`, `internal/store/`, `internal/httpapi/account.go`):
  - Migration 0002 adds the phone, its hash, `onboarded_at`, sessions and OTP state, and keeps old guest rows.
  - Session tokens are random and stored hashed.
  - The code flow enforces every limit in the database or in memory; a try is spent before any comparison, and a code works once.
  - A `Verifier` interface has a dev implementation (code on screen and in the log) and a plain-HTTP Twilio Verify client. With neither configured the server refuses to sign anyone in, and a half-set Twilio config stops it at startup.
  - Usernames are checked against real and simulated fans. Accounts that aren't onboarded stay out of search and friend lists. Contacts are matched by hash. Delete account runs as one transaction.
- **App** (`src/app/`, `src/features/auth/`, `src/features/onboarding/`, `src/data/session.ts`):
  - The signed-in screens moved into an `(app)` group; URLs are unchanged.
  - The root layout guards `(auth)`, `onboarding` and `(app)` with `Stack.Protected`.
  - Screens:
    - Welcome, led by a "Season Ticket"
    - The phone number with a country picker (cricket nations first), formatted as you type
    - Six code cells with one-time-code autofill and a resend countdown
    - The fan card with a live ticket preview and a username check as you type
    - "Find your people": contacts, invite, fans to follow
  - The token lives in SecureStore. A 401 sends the fan back to Welcome.
- **Settings:** a username field, plus an Account section with the masked number, Sign out and Delete account.
- **Friends screen:** the contacts finder and invite now sit there too, so "find friends later" is true.
- **Empty accounts:** the Bragging Board says "Just you on the board", Predictor Rank says "Unranked", and the friends' split shows an empty track instead of a fake 50/50.

## Found and fixed during the emulator walk

- **Keyboard on Android:** on Android 16 the app draws edge to edge, so the window no longer shrinks for the keyboard. KeyboardAvoidingView only lifted part of it, which left the username field, the stub reply bar and the Log review box under the keyboard. The new `KeyboardAware` pads by the keyboard's real height on Android and keeps KeyboardAvoidingView on iOS. It is used on the fan card, the stub page and the Log flow. The stub page and the Log flow had the same Android-only setup before this phase, so their text boxes were probably covered too.
- **"Next" on the name field** closed the keyboard instead of moving to the username.
- **Typing "+44…" a key at a time** dropped the "+", so the number was read as Indian. The field now keeps a typed international prefix until the country is known. Pasting already worked.
- **Writing contacts:** the contacts library also asked for Android's `WRITE_CONTACTS`, which Innings never uses. It's now blocked in `app.json`.

## Decisions Made

- New accounts start empty; there's no demo history. Every screen was checked on an empty account.
- Codes and sends are rate-limited per number in the database, so a restart doesn't reset them, and per IP in memory.
- The "send me a code" answer never says whether a number has an account.

## Deviations from Plan

- The Twilio Verify client is tested only against a mock built from Twilio's documented API, never the real service.
- Whether Twilio Verify handles India's DLT rules is unconfirmed. Sending to Indian numbers may need the developer's own registration (SMS-01).
- Sign in with Apple and Google were not built (ACCT-09).
- The iOS build was not re-run in this phase; it was walked afterwards, on 2026-10-28 (see `07-VALIDATION.md`). That walk found two keyboard bugs on iPhone, on the fan card and in the log's last step, and both were fixed. At the time of this phase: the developer was using the iOS Simulator, so all the visual checks were on Android. The new native modules need a fresh `npx expo run:ios`.

## Issues Encountered

- A dev-only warning, "Can't perform a React state update on a component that hasn't mounted yet", comes from Expo Router's own link handling when the dev client opens the app through its link. It doesn't appear in the app and isn't from Innings code.

## Next Phase Readiness

Sign-in is complete for testing. Real texts need an SMS provider account; store builds need the developer's Apple and Google accounts.

## Self-Check: PASSED

`go test ./...` (5 packages), `go vet`, `gofmt`, `npx tsc --noEmit` and 39 Jest tests pass. A 70-check end-to-end run against a fresh server passed. The whole flow was walked on the Android emulator. See `07-VALIDATION.md`.
