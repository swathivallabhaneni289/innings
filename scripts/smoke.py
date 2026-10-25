"""End-to-end check of the API: phone sign-in, onboarding, the main reads and writes, sign out and delete account.

Run it against a fresh server in dev mode, on a spare port and a throwaway database:

    DB_PATH=/tmp/innings-smoke.db PORT=8096 INNINGS_DEV_OTP=1 go run ./cmd/api
    python3 scripts/smoke.py http://localhost:8096

It takes about 70 seconds, because it waits out the 30-second resend gap twice.
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


def sign_in(phone):
    s, st, _ = req('POST', '/v1/auth/start', body={'phone': phone})
    assert s == 200, (s, st)
    s, v, _ = req('POST', '/v1/auth/verify', body={'phone': phone, 'code': st['devCode']})
    assert s == 200, (s, v)
    return v


# ---- sign-in errors ----
check('bad phone 400', req('POST', '/v1/auth/start', body={'phone': '98765'})[0] == 400)
s, st, _ = req('POST', '/v1/auth/start', body={'phone': '+919876543210'})
check('start 200 with devCode', s == 200 and len(st.get('devCode', '')) == 6 and st['resendInSeconds'] == 30 and st['expiresInSeconds'] == 600)
s, b, h = req('POST', '/v1/auth/start', body={'phone': '+919876543210'})
check('resend too soon 429 + Retry-After', s == 429 and b.get('error') == 'too-soon' and 'Retry-After' in h)
wrong = '000000' if st['devCode'] != '000000' else '111111'
s, b, _ = req('POST', '/v1/auth/verify', body={'phone': '+919876543210', 'code': wrong})
check('wrong code 400 bad-code', s == 400 and b.get('error') == 'bad-code')
check('no code for other number 410', req('POST', '/v1/auth/verify', body={'phone': '+919999999999', 'code': '123456'})[0] == 410)
check('401 without token', req('GET', '/v1/home')[0] == 401)
check('bootstrap needs a session', req('POST', '/v1/bootstrap')[0] == 401)

# ---- sign up ----
s, v, _ = req('POST', '/v1/auth/verify', body={'phone': '+919876543210', 'code': st['devCode']})
check('verify creates account', s == 200 and v['isNew'] and not v['account']['onboarded'] and v['account']['phone'] == '+919876543210')
T = v['token']
check('code cannot be reused', req('POST', '/v1/auth/verify', body={'phone': '+919876543210', 'code': st['devCode']})[0] == 410)
s, boot, _ = req('POST', '/v1/bootstrap', T)
check('bootstrap with session', s == 200 and len(boot['teams']) > 50 and len(boot['players']) > 300 and 'account' in boot)
check('phone not in public user', 'phone' not in boot['me'])

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
v2 = sign_in('+447700900123')
T2 = v2['token']
check('second fan cannot take the handle', req('PUT', '/v1/me', T2, {'name': 'Other', 'handle': 'swathi'})[0] == 409)
req('PUT', '/v1/me', T2, {'name': 'Ravi', 'handle': 'ravi_k'})
check('not onboarded yet: hidden from search', not any(u['handle'] == 'ravi_k' for u in req('GET', '/v1/search?q=ravi', T)[1]['users']))
req('POST', '/v1/me/onboarded', T2)
check('onboarded: in search', any(u['handle'] == 'ravi_k' for u in req('GET', '/v1/search?q=ravi', T)[1]['users']))
import hashlib
h = hashlib.sha256(b'+447700900123').hexdigest()
s, cm, _ = req('POST', '/v1/contacts/match', T, {'hashes': [h, hashlib.sha256(b'+15550000000').hexdigest()]})
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

# ---- second device: same account (wait out the 30-second resend gap first) ----
time.sleep(31)
v3 = sign_in('+919876543210')
check('second device signs in to the same account', not v3['isNew'] and v3['me']['id'] == me['id'] and v3['account']['onboarded'])
T3 = v3['token']
check('logout 204', req('POST', '/v1/auth/logout', T3)[0] == 204)
check('logged-out token is refused', req('GET', '/v1/home', T3)[0] == 401)
check('first device still signed in', req('GET', '/v1/home', T)[0] == 200)

check('delete account 204', req('DELETE', '/v1/me', T)[0] == 204)
check('deleted account token refused', req('GET', '/v1/home', T)[0] == 401)
for p in ['/v1/home', '/v1/notifications', '/v1/friends', '/v1/users/me', '/v1/search?q=swathi']:
    s, b, _ = req('GET', p, T2)
    check(f'after delete, other fan GET {p} -> {s}', s == 200 and not bad_nulls(b))
check('deleted stub is gone', req('GET', f'/v1/logs/{LID}', T2)[0] == 404)
check('deleted fan not in search', not any(u['handle'] == 'swathi' for u in req('GET', '/v1/search?q=swathi', T2)[1]['users']))
time.sleep(31)
v4 = sign_in('+919876543210')
check('same number after delete starts a new account', v4['isNew'] and v4['me']['id'] != me['id'])

print(f'PASSED {ok} checks; FAILED {len(fail)}')
for x in fail:
    print(' FAIL', x)
