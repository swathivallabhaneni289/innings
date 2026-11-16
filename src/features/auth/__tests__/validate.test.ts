import { describe, expect, it } from '@jest/globals';

import { lengthStatus, looksLikeEmail, normEmail, passwordLength } from '../validate';

describe('email', () => {
  it('normalises like the server: trimmed and lower-cased', () => {
    expect(normEmail('  Swathi.V@Gmail.COM ')).toBe('swathi.v@gmail.com');
  });

  it('accepts real addresses', () => {
    for (const e of ['swathi@gmail.com', ' Fan+innings@example.co.uk ', "o'brien@example.ie", 'a_b-c@mail.example-site.org']) {
      expect(looksLikeEmail(e)).toBe(true);
    }
  });

  it('refuses what mail cannot reach', () => {
    for (const e of ['', 'swathi', 'swathi@', '@gmail.com', 'swathi@gmail', 'a b@gmail.com', 'a@b@c.com', 'swathi@gmail..com', 'swathi@.com', `a@${'b'.repeat(250)}.com`]) {
      expect(looksLikeEmail(e)).toBe(false);
    }
  });
});

describe('password length', () => {
  it('counts characters, not bytes', () => {
    expect(passwordLength('abc')).toBe(3);
    expect(passwordLength('सात सात')).toBe(7);
    expect(passwordLength('🏏🏏')).toBe(2);
  });

  it('says how many are still to go, and when it is long enough', () => {
    expect(lengthStatus('')).toEqual({ ok: false, toGo: 8, tooLong: false });
    expect(lengthStatus('kohli18')).toEqual({ ok: false, toGo: 1, tooLong: false });
    expect(lengthStatus('kohli183')).toEqual({ ok: true, toGo: 0, tooLong: false });
    expect(lengthStatus('x'.repeat(128)).ok).toBe(true);
    expect(lengthStatus('x'.repeat(129))).toEqual({ ok: false, toGo: 0, tooLong: true });
  });
});
