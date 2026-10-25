import { describe, expect, it } from '@jest/globals';

import { callingCode, countryOf, formatAsYouType, formatPhone, maskPhone, normalizeContactNumber, splitInternational, toE164 } from '../phone';

describe('normalizeContactNumber', () => {
  it.each([
    ['098765 43210'], // national, with the trunk 0
    ['98765 43210'], // national, no trunk prefix
    ['+91 98765 43210'], // international
    ['0091 98765 43210'], // international, with the 00 dialling prefix
    ['+91-98765-43210'], // dashes
    ['(+91) 98765.43210'], // brackets and dots
    ['919876543210'], // country code without the plus
  ])('reads the Indian number %j as +919876543210', (raw) => {
    expect(normalizeContactNumber(raw, 'IN')).toBe('+919876543210');
  });

  it('reads a UK mobile with the GB default', () => {
    expect(normalizeContactNumber('07911 123456', 'GB')).toBe('+447911123456');
  });

  it('reads UK numbers with a Crown Dependency default (+44 numbers can parse as GG, JE or IM)', () => {
    expect(normalizeContactNumber('07400 123456', 'GG')).toBe('+447400123456');
    expect(normalizeContactNumber('020 7946 0958', 'IM')).toBe('+442079460958');
  });

  it('keeps a number’s own country code whatever the default', () => {
    expect(normalizeContactNumber('+44 7911 123456', 'IN')).toBe('+447911123456');
    expect(normalizeContactNumber('+91 98765 43210', 'GB')).toBe('+919876543210');
  });

  it.each([[''], ['   '], ['abc'], ['N/A'], ['+'], ['---']])('drops the junk entry %j', (raw) => {
    expect(normalizeContactNumber(raw, 'IN')).toBeNull();
  });

  it('drops a number that is too short', () => {
    expect(normalizeContactNumber('98765', 'IN')).toBeNull();
    expect(normalizeContactNumber('12345', 'GB')).toBeNull();
  });

  it('drops a number that is too long', () => {
    expect(normalizeContactNumber('1234567890123456', 'IN')).toBeNull();
  });
});

describe('typing a number', () => {
  it('formats as you type, in the country’s own style', () => {
    expect(formatAsYouType('9876543210', 'IN')).toBe('98765 43210');
    expect(formatAsYouType('', 'IN')).toBe('');
  });
  it('gives E.164 only once the number is valid', () => {
    expect(toE164('98765 4321', 'IN')).toBeNull();
    expect(toE164('98765 43210', 'IN')).toBe('+919876543210');
  });
  it('follows a pasted international number to its country', () => {
    expect(splitInternational('+91 98765 43210', 'GB')).toEqual({ country: 'IN', national: '9876543210' });
    expect(splitInternational('+44 7911 123456', 'GB')).toEqual({ country: 'GB', national: '7911123456' });
    expect(splitInternational('98765 43210', 'IN')).toBeNull();
  });
  it('knows calling codes and countries', () => {
    expect(callingCode('IN')).toBe('+91');
    expect(callingCode('GB')).toBe('+44');
    expect(countryOf('+919876543210')).toBe('IN');
    expect(countryOf('')).toBeUndefined();
  });
});

describe('printing a number', () => {
  it('formats E.164 for display', () => {
    expect(formatPhone('+919876543210')).toBe('+91 98765 43210');
    expect(formatPhone('+447911123456')).toBe('+44 7911 123456');
  });
  it('masks all but the country code and the last 4 digits', () => {
    expect(maskPhone('+919876543210')).toBe('+91 ••••• •3210');
    expect(maskPhone('+447911123456')).toBe('+44 •••• ••3456');
  });
  it('still hides most of a number it cannot parse', () => {
    expect(maskPhone('12345678')).toBe('•••• 5678');
  });
});
