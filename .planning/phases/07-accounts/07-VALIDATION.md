---
phase: 7
slug: accounts
status: complete
nyquist_compliant: false
wave_0_complete: true
created: 2026-10-25
---

# Phase 7: Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` with `httptest` (server, Twilio mock); Jest with `jest-expo` (phone number handling); TypeScript strict; an end-to-end Python script against a running server; the Android emulator for the screens |
| **Quick run command** | `go test ./internal/auth ./internal/httpapi` and `npx tsc --noEmit` |
| **Full suite command** | `go vet ./... && go test ./... && npx tsc --noEmit && npx jest` |
| **Estimated runtime** | ~30 seconds, plus about 70 seconds for the end-to-end run (it waits out the 30-second resend gap twice) |

---

## Per-Task Verification Map

| Plan | Requirement | Behavior | Test Type | Automated Command | Status |
|------|-------------|----------|-----------|-------------------|--------|
| 07-01 | ACCT-01 | Sign up, then the same number on a second device opens the same account | integration | `go test ./internal/httpapi -run TestSignInAndSecondDevice` | green |
| 07-01 | ACCT-02 | Wrong code counts down, expiry, single use, resend gap, hourly cap, IP cap | integration | `go test ./internal/httpapi -run 'TestWrongCode|TestCodeExpires|TestCodeCantBeReused|TestResendTooSoon|TestHourlyCap|TestIPCap'` | green |
| 07-01 | ACCT-03 | Dev mode returns the code; Twilio mode never does; no provider means 503 | integration | `go test ./internal/httpapi -run 'TestNoSMSProvider|TestTwilioMode'`, `go test ./internal/auth -run 'TestTwilio|TestVerifierFromEnv'` | green (Twilio against a mock) |
| 07-01 | ACCT-04 | Username rules, reserved words, clashes with simulated fans, taken, your own | unit + integration | `go test ./internal/domain -run TestHandleRules`, `go test ./internal/httpapi -run 'TestHandles|TestOnboarding'` | green |
| 07-01 | ACCT-05 | Contacts matched by hash; Indian and UK numbers normalised from every common format | integration + unit | `go test ./internal/httpapi -run TestContactsMatch`, `npx jest phone` | green |
| 07-01 | ACCT-06 | Logout ends one session; delete removes everything and other fans' screens still load | integration | `go test ./internal/httpapi -run 'TestLogoutEndsOnlyThisSession|TestDeleteAccount'`, `go test ./internal/store -run TestDeleteUser` | green |
| 07-01 | ACCT-07 | Sessions stored hashed; unknown or missing token is 401; old guest database migrates | integration | `go test ./internal/store -run 'TestSessions|TestMigrateGuestDB'`, `go test ./internal/httpapi -run 'TestBootstrapResumes|TestUnauthorized'` | green |
| 07-01 | ACCT-08 | Every screen on an empty account | manual (emulator) | see below | green |
| 07-01 | all | The whole API flow end to end on a fresh database | end to end | `smoke2.py` against `INNINGS_DEV_OTP=1` on port 8096: 70 of 70 checks | green |

*Status: pending · green · red · flaky*

---

## Results (2026-10-25)

```
$ gofmt -l cmd internal          (nothing)
$ go vet ./...                   ok
$ go test -count=1 ./...
ok  	.../internal/auth
ok  	.../internal/domain
ok  	.../internal/httpapi
ok  	.../internal/sim
ok  	.../internal/store
$ npx tsc --noEmit               exit 0
$ npx jest                       Tests: 39 passed, 39 total
$ python3 smoke2.py              PASSED 70 checks; FAILED 0
```

Migration 0002 was run on a copy of an existing `innings.db`. Its 3 old guests were kept, hidden and not onboarded, and their 15 logs and 15 follows were intact. An old guest token gets 401.

---

## Emulator walk (Pixel 8, Android 16, new development build)

Each step was checked with a screenshot:

1. Welcome.
2. Phone number formatted as you type.
3. The code screen with the dev code.
4. A wrong code: "4 tries left".
5. The right code opens the fan card.
6. Closing and reopening the app resumes the fan card.
7. Name, then "next" into the username: "@swathi is yours".
8. Find your people:
   - Contacts denied gives "Open Settings" and "Ask again".
   - Allowed with no matches gives "Invite friends".
9. Start watching with no follows, on an empty account:
   - Home: the friends' feed and the Bragging Board have their empty states.
   - Profile: "–" for the averages, and Predictor Rank shows "Unranked".
10. Settings, Account. Sign out, then sign in again with the same number: straight to Home.
11. Following three fans fills the feed.
12. A reply on a stub with the keyboard up.
13. A full log with a typed review; the first stub prints as No. 0001.
14. Delete account. The server returns 204, and the database has no user, logs, replies or sessions for that number.
15. A second sign-up with a UK number typed as "+44…".

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| A real text arrives | ACCT-03 | Needs a Twilio account (and India DLT for +91) | Set the three `TWILIO_*` settings, sign in with your own number |
| One-time-code autofill from a real SMS | ACCT-01 | Needs a real text | As above, on iPhone and Android |
| The iPhone build with the new native modules | ACCT-07 | The iOS Simulator was in use by the developer | `npx expo run:ios`, then walk sign-up |
| Contacts that do match | ACCT-05 | Simulated fans have no phone numbers | Two real accounts whose numbers are in each other's contacts |

---

## Validation Sign-Off

- [x] Every requirement has an automated check or a recorded emulator walk
- [x] No watch-mode flags
- [ ] `nyquist_compliant: true`: not set; the screens have no component tests and the four manual checks above need accounts or devices

**Approval:** automated gates and the Android walk passed 2026-10-25
