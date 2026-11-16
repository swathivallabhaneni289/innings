"""End-to-end check of the API: email sign-up and sign-in, onboarding, the main reads and writes, changing and resetting
a password, sign out and delete account.

Run it against a fresh server in dev mode, on simulated matches (they always include a live one), a spare port and a
throwaway database (never `make api`, which reads .env and could send real emails):

    DB_PATH=/tmp/innings-smoke.db PORT=8096 INNINGS_DEV_CODES=1 INNINGS_MATCHES=sim go run ./cmd/api
    python3 scripts/smoke.py http://localhost:8096

It takes about 35 seconds, because it waits out the 30-second resend gap once.
"""
import json, sys, time, urllib.request, urllib.error

B = sys.argv[1] if len(sys.argv) > 1 else 'http://localhost:8096'


def req(method, path, token=None, body=None):
    r = urllib.request.Request(B + path, method=method, data=json.dumps(body).encode() if body is not None else None,
                               headers={'Content-Type': 'application/json', **({'Authorization': 'Bearer ' + token} if token else {})})
    try:
        with urllib.request.urlopen(r) as resp:
            raw = resp.read()
            return resp.status, (json.loads(raw) if raw else None), dict(resp.headers)
    except urllib.error.HTTPError as e:
        raw = e.read()
        return e.code, (json.loads(raw) if raw else {}), dict(e.headers)


ok = 0
fail = []


def check(name, cond):
    global ok
    if cond:
        ok += 1
    else:
        fail.append(name)


ALLOWED_NULL = {'winnerId', 'officialPotmId', 'fanPotmId', 'logId', 'matchId', 'yourCall', 'yourCallSplit', 'yourCallMine', 'topReview',
                'onThisDay', 'avgRatingGiven', 'accuracy', 'prediction', 'outcome', 'log', 'watching', 'fans', 'fansPotmId', 'reason', 'message'}


def bad_nulls(o, k=None):
    if o is None:
        return [] if k in ALLOWED_NULL else [k]
    if isinstance(o, dict):
        return [x for kk, v in o.items() for x in bad_nulls(v, kk)]
    if isinstance(o, list):
        return [x for v in o for x in bad_nulls(v, k)]
    return []


def session(path, body):
    s, v, _ = req('POST', path, body=body)
    assert s == 200, (path, s, v)
    return v


def sign_up(email, password):
    s, sent, _ = req('POST', '/v1/auth/signup', body={'email': email, 'password': password})
    assert s == 200, (s, sent)
    return session('/v1/auth/confirm', {'email': email, 'code': sent['devCode']})


def sign_in(email, password):
    return session('/v1/auth/signin', {'email': email, 'password': password})


EMAIL, PASSWORD = 'swathi@example.com', 'cover-drive-forever'

# ---- sign-up errors ----
check('bad email 400', req('POST', '/v1/auth/signup', body={'email': 'swathi@', 'password': PASSWORD})[1].get('error') == 'bad-email')
check('short password 400', req('POST', '/v1/auth/signup', body={'email': EMAIL, 'password': 'short'})[1].get('error') == 'weak-password')
check('common password 400', req('POST', '/v1/auth/signup', body={'email': EMAIL, 'password': 'password123'})[1].get('error') == 'weak-password')
s, st, _ = req('POST', '/v1/auth/signup', body={'email': ' Swathi@Example.COM ', 'password': PASSWORD})
check('sign up 200 with devCode', s == 200 and st['email'] == EMAIL and len(st.get('devCode', '')) == 6 and st['resendInSeconds'] == 30 and st['expiresInSeconds'] == 600)
s, b, h = req('POST', '/v1/auth/confirm/resend', body={'email': EMAIL})
check('resend too soon 429 + Retry-After', s == 429 and b.get('error') == 'too-soon' and 'Retry-After' in h)
s, b, _ = req('POST', '/v1/auth/signin', body={'email': EMAIL, 'password': PASSWORD})
check('sign in before confirming 403', s == 403 and b.get('error') == 'email-unconfirmed')
wrong = '000000' if st['devCode'] != '000000' else '111111'
s, b, _ = req('POST', '/v1/auth/confirm', body={'email': EMAIL, 'code': wrong})
check('wrong code 400 bad-code', s == 400 and b.get('error') == 'bad-code')
check('no code for another address 410', req('POST', '/v1/auth/confirm', body={'email': 'nobody@example.com', 'code': '123456'})[0] == 410)
check('401 without token', req('GET', '/v1/home')[0] == 401)
check('bootstrap needs a session', req('POST', '/v1/bootstrap')[0] == 401)

# ---- sign up ----
s, v, _ = req('POST', '/v1/auth/confirm', body={'email': EMAIL, 'code': st['devCode']})
check('confirm signs in a new account', s == 200 and v['isNew'] and not v['account']['onboarded'] and v['account']['email'] == EMAIL)
T = v['token']
check('code cannot be reused', req('POST', '/v1/auth/confirm', body={'email': EMAIL, 'code': st['devCode']})[0] == 410)
check('confirmed address is taken', req('POST', '/v1/auth/signup', body={'email': EMAIL, 'password': 'another-password'})[1].get('error') == 'email-taken')
s, b, _ = req('POST', '/v1/auth/signin', body={'email': EMAIL, 'password': 'not-my-password'})
s2, b2, _ = req('POST', '/v1/auth/signin', body={'email': 'nobody@example.com', 'password': PASSWORD})
check('wrong password and unknown address look the same', s == s2 == 401 and b == b2 and b.get('error') == 'bad-credentials')
s, boot, _ = req('POST', '/v1/bootstrap', T)
check('bootstrap with session', s == 200 and len(boot['teams']) > 50 and len(boot['players']) > 300 and 'account' in boot)
check('email not in public user', 'email' not in boot['me'])

# ---- empty account: every read works with no history ----
s, live, _ = req('GET', '/v1/matches?state=live', T)
s2, up, _ = req('GET', '/v1/matches?state=upcoming', T)
s3, rec, _ = req('GET', '/v1/matches?state=recent', T)
L = live['matches'][0]['id']; U = up['matches'][0]; R = rec['matches'][0]['id']
for p in ['/v1/home', '/v1/matches', '/v1/matches?competition=ipl', '/v1/matches?team=ind,ind-w', f'/v1/matches/{L}', f'/v1/matches/{U["id"]}', f'/v1/matches/{R}',
          '/v1/matches/wc-2011-final', '/v1/vault', '/v1/vault/rankings?format=all&era=all', '/v1/vault/rankings?format=Test&era=2000s', '/v1/users/me',
          '/v1/users/f-aditi', '/v1/friends', '/v1/search?q=india', '/v1/log/suggestions', '/v1/notifications']:
    s, b, _ = req('GET', p, T)
    nulls = bad_nulls(b)
    check(f'empty account GET {p} -> {s} nulls={nulls[:3]}', s == 200 and not nulls)
s, me, _ = req('GET', '/v1/users/me', T)
check('new account has no logs or follows', me['logs'] == [] and me['stats']['logged'] == 0)

# ---- onboarding ----
check('onboarded needs a profile', req('POST', '/v1/me/onboarded', T)[1].get('error') == 'profile-incomplete')
s, hc, _ = req('GET', '/v1/handles/Swathi', T)
check('handle check lower-cases and is free', s == 200 and hc['handle'] == 'swathi' and hc['available'])
check('reserved handle', req('GET', '/v1/handles/admin', T)[1]['reason'] == 'reserved')
check('invalid handle', req('GET', '/v1/handles/1abc', T)[1]['reason'] == 'invalid')
check('sim fan handle taken', req('GET', '/v1/handles/aditi', T)[1]['reason'] == 'taken')
s, me, _ = req('PUT', '/v1/me', T, {'name': 'Swathi', 'handle': 'swathi', 'favouriteTeamId': 'ind'})
check('save fan card', s == 200 and me['handle'] == 'swathi')
s, ob, _ = req('POST', '/v1/me/onboarded', T)
check('onboarded', s == 200 and ob['account']['onboarded'])
check('own handle counts as available', req('GET', '/v1/handles/swathi', T)[1]['available'])

# ---- a second fan ----
v2 = sign_up('ravi@example.com', 'ravi-bats-at-three')
T2 = v2['token']
check('second fan cannot take the handle', req('PUT', '/v1/me', T2, {'name': 'Other', 'handle': 'swathi'})[0] == 409)
req('PUT', '/v1/me', T2, {'name': 'Ravi', 'handle': 'ravi_k'})
check('not onboarded yet: hidden from search', not any(u['handle'] == 'ravi_k' for u in req('GET', '/v1/search?q=ravi', T)[1]['users']))
req('POST', '/v1/me/onboarded', T2)
check('onboarded: in search', any(u['handle'] == 'ravi_k' for u in req('GET', '/v1/search?q=ravi', T)[1]['users']))
import hashlib
h = hashlib.sha256(b'ravi@example.com').hexdigest()
s, cm, _ = req('POST', '/v1/contacts/match', T, {'hashes': [h, hashlib.sha256(b'stranger@example.com').hexdigest()]})
check('contacts match finds the second fan', s == 200 and [m['user']['handle'] for m in cm['matches']] == ['ravi_k'])
check('bad hashes 400', req('POST', '/v1/contacts/match', T, {'hashes': ['nope']})[0] == 400)

# ---- the v1.0 checks ----
xi = U['xi']; t0, t1 = U['teamIds']
s, _, _ = req('PUT', f'/v1/matches/{U["id"]}/prediction', T, {'winnerId': t0, 'topRunScorerId': xi[t0][0], 'topWicketTakerId': xi[t1][10]})
check('predict upcoming 200', s == 200)
s, b, _ = req('GET', f'/v1/matches/{L}', T); lm = b['match']; a, c = lm['teamIds']
s, _, _ = req('PUT', f'/v1/matches/{L}/prediction', T, {'winnerId': a, 'topRunScorerId': lm['xi'][a][0], 'topWicketTakerId': lm['xi'][c][10]})
check('predict live is 409 locked', s == 409)
log = {'matchId': R, 'rating': 4.5, 'instantClassic': True, 'playerRatings': {}, 'fanPotmId': None, 'watchedAt': 'hostel', 'placeNote': 'Hostel room',
       'watchedWith': [], 'companions': 6, 'review': 'What a night.'}
s, f1, _ = req('POST', '/v1/logs', T, log); check('log create 201', s == 201)
log['rating'] = 3.5
s, f2, _ = req('POST', '/v1/logs', T, log); check('re-log keeps stub number', s in (200, 201) and f2['log']['stubNo'] == f1['log']['stubNo'])
check('bad rating rejected', req('POST', '/v1/logs', T, {**log, 'rating': 4.3})[0] == 400)
LID = f1['log']['id']
s, a1, _ = req('POST', f'/v1/logs/{LID}/like', T); s, a2, _ = req('POST', f'/v1/logs/{LID}/like', T)
check('like toggles', a1['liked'] and not a2['liked'])
s, rp, _ = req('POST', f'/v1/logs/{LID}/replies', T, {'text': 'Brilliant'}); check('reply 201', s == 201)
s, f, _ = req('POST', '/v1/users/f-aditi/follow', T); s, g, _ = req('POST', '/v1/users/f-aditi/follow', T)
check('follow toggles', f['following'] != g['following'])
s, t5, _ = req('PUT', '/v1/me/top5', T, {'matchIds': ['wc-2011-final', 'wc-2019-final']}); check('top5 set', s == 200 and len(t5['top5']) == 2)
check('top5 rejects 6', req('PUT', '/v1/me/top5', T, {'matchIds': ['wc-2011-final'] * 6})[0] == 400)
s, me, _ = req('PUT', '/v1/me', T, {'bio': 'Dhoni finishes, I cry'}); check('update profile', s == 200 and me['bio'] == 'Dhoni finishes, I cry')
s, w, _ = req('PUT', f'/v1/matches/{L}/watching', T, {'place': 'friends'}); check('I am watching', s == 200 and w['watching'] == 'friends')

# ---- the second fan interacts with the first fan's stub, then the first fan deletes their account ----
req('POST', f'/v1/users/{me["id"]}/follow', T2)
s, _, _ = req('POST', f'/v1/logs/{LID}/like', T2); check('second fan likes the stub', s == 200)
s, _, _ = req('POST', f'/v1/logs/{LID}/replies', T2, {'text': 'Same, mate'}); check('second fan replies', s == 201)

# ---- second device: same account ----
v3 = sign_in(EMAIL, PASSWORD)
check('second device signs in to the same account', not v3['isNew'] and v3['me']['id'] == me['id'] and v3['account']['onboarded'])
T3 = v3['token']
check('logout 204', req('POST', '/v1/auth/logout', T3)[0] == 204)
check('logged-out token is refused', req('GET', '/v1/home', T3)[0] == 401)
check('first device still signed in', req('GET', '/v1/home', T)[0] == 200)

# ---- change password: other devices are signed out, this one stays ----
T3 = sign_in(EMAIL, PASSWORD)['token']
s, b, _ = req('PUT', '/v1/me/password', T, {'currentPassword': 'not-my-password', 'newPassword': 'a-better-password'})
check('wrong current password 403, still signed in', s == 403 and b.get('error') == 'bad-password' and req('GET', '/v1/home', T)[0] == 200)
check('change password 204', req('PUT', '/v1/me/password', T, {'currentPassword': PASSWORD, 'newPassword': 'a-better-password'})[0] == 204)
check('other device signed out by the change', req('GET', '/v1/home', T3)[0] == 401)
check('this device still signed in', req('GET', '/v1/home', T)[0] == 200)
check('old password refused', req('POST', '/v1/auth/signin', body={'email': EMAIL, 'password': PASSWORD})[0] == 401)
PASSWORD = 'a-better-password'

# ---- forgot password (wait out the 30-second resend gap from sign-up first) ----
time.sleep(31)
s, fp, _ = req('POST', '/v1/auth/password/forgot', body={'email': EMAIL})
check('forgot password 200 with devCode', s == 200 and len(fp.get('devCode', '')) == 6)
s, nb, _ = req('POST', '/v1/auth/password/forgot', body={'email': 'nobody@example.com'})
check('forgot for an unknown address: same answer, no code', s == 200 and 'devCode' not in nb and nb['expiresInSeconds'] == 600)
s, rc, _ = req('POST', '/v1/auth/password/code', body={'email': EMAIL, 'code': fp['devCode']})
check('reset code buys a token', s == 200 and len(rc['resetToken']) == 64 and rc['expiresInSeconds'] == 900)
check('weak new password 400', req('POST', '/v1/auth/password/reset', body={'resetToken': rc['resetToken'], 'password': 'short'})[1].get('error') == 'weak-password')
s, rs, _ = req('POST', '/v1/auth/password/reset', body={'resetToken': rc['resetToken'], 'password': 'reset-and-ready'})
check('reset signs in', s == 200 and rs['me']['id'] == me['id'])
check('every other device signed out by the reset', req('GET', '/v1/home', T)[0] == 401)
check('reset token is single use', req('POST', '/v1/auth/password/reset', body={'resetToken': rc['resetToken'], 'password': 'again-and-again'})[0] == 410)
T, PASSWORD = rs['token'], 'reset-and-ready'

# ---- delete account: the password is asked for again ----
s, b, _ = req('DELETE', '/v1/me', T, {'password': 'not-my-password'})
check('delete with a wrong password 403', s == 403 and b.get('error') == 'bad-password')
check('delete account 204', req('DELETE', '/v1/me', T, {'password': PASSWORD})[0] == 204)
check('deleted account token refused', req('GET', '/v1/home', T)[0] == 401)
for p in ['/v1/home', '/v1/notifications', '/v1/friends', '/v1/users/me', '/v1/search?q=swathi']:
    s, b, _ = req('GET', p, T2)
    check(f'after delete, other fan GET {p} -> {s}', s == 200 and not bad_nulls(b))
check('deleted stub is gone', req('GET', f'/v1/logs/{LID}', T2)[0] == 404)
check('deleted fan not in search', not any(u['handle'] == 'swathi' for u in req('GET', '/v1/search?q=swathi', T2)[1]['users']))
v4 = sign_up(EMAIL, 'a-fresh-start-1')
check('same address after delete starts a new account', v4['isNew'] and v4['me']['id'] != me['id'])

print(f'PASSED {ok} checks; FAILED {len(fail)}')
for x in fail:
    print(' FAIL', x)
