// Phone numbers: parsing what fans type and what their address book holds into E.164, and printing them back.
// Pure functions over libphonenumber-js. Hashing for contact matching lives with the contacts code, not here,
// so these stay testable without native crypto.
import { AsYouType, getCountryCallingCode, parsePhoneNumberFromString, type CountryCode } from 'libphonenumber-js';

export type { CountryCode };

/**
 * A number as saved in someone's contacts → E.164 ("+919876543210"), or null when it isn't a real number.
 * `defaultCountry` applies to numbers saved without a country code ("098765 43210", "98765 43210");
 * numbers with one ("+91…", "0091…") keep theirs.
 */
export function normalizeContactNumber(raw: string, defaultCountry: CountryCode): string | null {
  if (!raw || !/\d/.test(raw)) return null;
  const p = parsePhoneNumberFromString(raw.trim(), defaultCountry);
  return p && p.isValid() ? p.number : null;
}

/** The national number a fan typed, plus the country they picked → E.164, or null until it's a valid number. */
export function toE164(national: string, country: CountryCode): string | null {
  return normalizeContactNumber(national, country);
}

/** Formats digits as they're typed, in the country's own style: "98765 43210", "07911 123456". */
export function formatAsYouType(digits: string, country: CountryCode): string {
  return digits ? new AsYouType(country).input(digits) : '';
}

/** "+91" for IN. */
export const callingCode = (country: CountryCode) => `+${getCountryCallingCode(country)}`;

/**
 * A pasted or autofilled "+44 7911 123456" → the country and the national digits, so the picker can follow it.
 * Keeps `current` when it shares the calling code (a +44 number stays GB rather than turning into Guernsey).
 */
export function splitInternational(raw: string, current: CountryCode): { country: CountryCode; national: string } | null {
  if (!raw.trim().startsWith('+')) return null;
  const p = parsePhoneNumberFromString(raw.trim());
  if (!p) return null;
  const country = p.countryCallingCode === getCountryCallingCode(current) ? current : p.country;
  return country ? { country, national: p.nationalNumber } : null;
}

/** The country an E.164 number belongs to (a UK mobile may come back as GB, GG, IM or JE: same +44 rules). */
export function countryOf(e164: string | null | undefined): CountryCode | undefined {
  return e164 ? parsePhoneNumberFromString(e164)?.country : undefined;
}

/** "+919876543210" → "+91 98765 43210". Anything unparseable comes back as given. */
export function formatPhone(e164: string): string {
  const p = parsePhoneNumberFromString(e164);
  return p ? p.formatInternational() : e164;
}

/** "+919876543210" → "+91 ••••• •3210": the country code and the last 4 digits show, the rest is dotted. */
export function maskPhone(e164: string): string {
  const p = parsePhoneNumberFromString(e164);
  if (!p) {
    const digits = e164.replace(/\D/g, '');
    return digits.length > 4 ? `•••• ${digits.slice(-4)}` : e164;
  }
  const cc = `+${p.countryCallingCode}`;
  const rest = p.formatInternational().slice(cc.length);
  const total = rest.replace(/\D/g, '').length;
  let seen = 0;
  const masked = rest.replace(/\d/g, (d) => (++seen > total - 4 ? d : '•'));
  return cc + masked;
}
