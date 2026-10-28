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
| 07-01 | ACCT-08 | Every screen on an empty account | manual (emulator, simulator) | see below | green |
| 07-01 | all | The whole API flow end to end on a fresh database | end to end | `python3 scripts/smoke.py` against a fresh `INNINGS_DEV_OTP=1` server: 70 of 70 checks | green |

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
$ python3 scripts/smoke.py http://localhost:8096
PASSED 70 checks; FAILED 0
```

The same checks were run again on 2026-10-28 after the keyboard fixes, with the same results. `npx expo lint` reports 13 errors and 1 warning, all in code from earlier phases (React's rules about refs and setting state during render or in effects). None are in the lines changed for the fixes.

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

## iPhone walk and two-phone checks (2026-10-28)

The iPhone was walked on a separate simulator ("Innings QA", iPhone 17, iOS 26.5) with a new development build, so the developer's own simulator wasn't disturbed. The app was driven by a small XCUITest command loop that taps, types and takes screenshots. Maestro was the first choice, but installing it was blocked as outside code. The Android emulator ran alongside it against the same throwaway server, so the two phones could find each other in their contacts.

On the iPhone:

1. The build has the contacts permission text, and Welcome loads, so the new native modules start.
2. The number formats as it's typed, a wrong code gives "4 tries left" with the keyboard still up, and the dev code opens the fan card.
3. Killing the app mid-sign-up and launching it again comes back to the fan card. Signed in, it comes back to Home.
4. Contacts denied gives "Open Settings" and no "Ask again", since iOS asks only once.
5. iOS 18's limited access works: sharing one contact found that fan.
6. A reply posts with the keyboard up, and a full log with a typed review prints stub No. 0001.
7. Sign out goes to Welcome. The same number signs straight back in to Home.
8. Delete account returns 204 and leaves no user, logs, replies or sessions for that number.
9. With the database reset while the phone still held a session, the app got a 401 at launch and went back to Welcome.

Across the two phones:

1. The iPhone found an Android fan from a contact saved as "098450 11111".
2. Android found the iPhone fan from a contact saved as "98450 22222".
3. The same number signed in on both phones opened one account, with two sessions.
4. Deleting the account on Android signed the iPhone out on its next request, and the other fan's follow of it was removed.

Two bugs were found on the iPhone and fixed:

- On the fan card, the keyboard covered the Name and Username boxes. iOS doesn't scroll a focused field into view the way Android does. The ScrollView now sets `automaticallyAdjustKeyboardInsets` (with `KeyboardAware scrollInsets`), and Settings got the same treatment.
- In the log, the keyboard hid Back and "Save & print" while a review was typed. KeyboardAvoidingView measures itself against its parent, so below the step header it lifted too little. `KeyboardAware` now wraps the whole screen, as the reply screen already did.

Both fixes were checked again on the iPhone and on Android.

Things that are not bugs, noted so they aren't chased again:

- "Open Settings" opened the top of the Settings app on the simulator, not Innings' own page. This is unverified on a real iPhone.
- React Native's dev-only warning toast, "Open debugger to view warnings", can sit over a screen's bottom button until it's closed.
- The driver types faster than a person. Typing while a screen was still sliding in lost digits, and a tap during the slide was missed. With a short pause, both worked every time.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| A real text arrives | ACCT-03 | Needs a Twilio account (and India DLT for +91) | Set the three `TWILIO_*` settings, sign in with your own number |
| One-time-code autofill from a real SMS | ACCT-01 | Needs a real text | As above, on iPhone and Android |
| Sign-in on real phones | ACCT-01, ACCT-05 | Only simulators and an emulator have been used | `npx expo run:ios --device` and `npx expo run:android --device`, then walk sign-up |
| "Open Settings" lands on Innings' page | ACCT-05 | The simulator opened the top of Settings | Deny contacts on a real iPhone, tap Open Settings |

---

## Validation Sign-Off

- [x] Every requirement has an automated check or a recorded emulator walk
- [x] No watch-mode flags
- [ ] `nyquist_compliant: true`: not set; the screens have no component tests and the four manual checks above need accounts or real phones

**Approval:** automated gates and the Android walk passed 2026-10-25; the iPhone walk and the two-phone checks passed 2026-10-28
