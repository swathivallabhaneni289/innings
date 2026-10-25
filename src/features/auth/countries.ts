// Countries for the sign-in picker. libphonenumber-js has no names, and Hermes can't be relied on for
// Intl.DisplayNames, so the English names are printed here (generated from ICU, every country it can dial).
import { getCountries, getCountryCallingCode, type CountryCode } from 'libphonenumber-js';

export interface Country {
  code: CountryCode;
  name: string;
  dial: string; // "+91"
}

/** Cricket nations first, in roughly the order fans will look for them; the West Indies islands close the list. */
export const CRICKET_NATIONS: CountryCode[] = [
  'IN', 'GB', 'AU', 'PK', 'LK', 'BD', 'NZ', 'ZA', 'AF', 'IE', 'ZW', 'NL', 'NP', 'AE', 'US', 'CA', 'JM', 'TT', 'BB', 'GY',
];

export const DEFAULT_COUNTRY: CountryCode = 'IN';

const NAMES: Partial<Record<CountryCode, string>> = {
  AC: "Ascension Island", AD: "Andorra", AE: "United Arab Emirates", AF: "Afghanistan", AG: "Antigua and Barbuda",
  AI: "Anguilla", AL: "Albania", AM: "Armenia", AO: "Angola", AR: "Argentina", AS: "American Samoa", AT: "Austria",
  AU: "Australia", AW: "Aruba", AX: "Åland Islands", AZ: "Azerbaijan", BA: "Bosnia and Herzegovina", BB: "Barbados",
  BD: "Bangladesh", BE: "Belgium", BF: "Burkina Faso", BG: "Bulgaria", BH: "Bahrain", BI: "Burundi", BJ: "Benin",
  BL: "Saint Barthélemy", BM: "Bermuda", BN: "Brunei", BO: "Bolivia", BQ: "Caribbean Netherlands", BR: "Brazil",
  BS: "Bahamas", BT: "Bhutan", BW: "Botswana", BY: "Belarus", BZ: "Belize", CA: "Canada",
  CC: "Cocos (Keeling) Islands", CD: "Congo (DRC)", CF: "Central African Republic", CG: "Congo (Republic)",
  CH: "Switzerland", CI: "Côte d’Ivoire", CK: "Cook Islands", CL: "Chile", CM: "Cameroon", CN: "China",
  CO: "Colombia", CR: "Costa Rica", CU: "Cuba", CV: "Cape Verde", CW: "Curaçao", CX: "Christmas Island",
  CY: "Cyprus", CZ: "Czechia", DE: "Germany", DJ: "Djibouti", DK: "Denmark", DM: "Dominica",
  DO: "Dominican Republic", DZ: "Algeria", EC: "Ecuador", EE: "Estonia", EG: "Egypt", EH: "Western Sahara",
  ER: "Eritrea", ES: "Spain", ET: "Ethiopia", FI: "Finland", FJ: "Fiji", FK: "Falkland Islands", FM: "Micronesia",
  FO: "Faroe Islands", FR: "France", GA: "Gabon", GB: "United Kingdom", GD: "Grenada", GE: "Georgia",
  GF: "French Guiana", GG: "Guernsey", GH: "Ghana", GI: "Gibraltar", GL: "Greenland", GM: "Gambia", GN: "Guinea",
  GP: "Guadeloupe", GQ: "Equatorial Guinea", GR: "Greece", GT: "Guatemala", GU: "Guam", GW: "Guinea-Bissau",
  GY: "Guyana", HK: "Hong Kong", HN: "Honduras", HR: "Croatia", HT: "Haiti", HU: "Hungary", ID: "Indonesia",
  IE: "Ireland", IL: "Israel", IM: "Isle of Man", IN: "India", IO: "British Indian Ocean Territory", IQ: "Iraq",
  IR: "Iran", IS: "Iceland", IT: "Italy", JE: "Jersey", JM: "Jamaica", JO: "Jordan", JP: "Japan", KE: "Kenya",
  KG: "Kyrgyzstan", KH: "Cambodia", KI: "Kiribati", KM: "Comoros", KN: "Saint Kitts and Nevis", KP: "North Korea",
  KR: "South Korea", KW: "Kuwait", KY: "Cayman Islands", KZ: "Kazakhstan", LA: "Laos", LB: "Lebanon",
  LC: "Saint Lucia", LI: "Liechtenstein", LK: "Sri Lanka", LR: "Liberia", LS: "Lesotho", LT: "Lithuania",
  LU: "Luxembourg", LV: "Latvia", LY: "Libya", MA: "Morocco", MC: "Monaco", MD: "Moldova", ME: "Montenegro",
  MF: "Saint Martin", MG: "Madagascar", MH: "Marshall Islands", MK: "North Macedonia", ML: "Mali", MM: "Myanmar",
  MN: "Mongolia", MO: "Macao", MP: "Northern Mariana Islands", MQ: "Martinique", MR: "Mauritania", MS: "Montserrat",
  MT: "Malta", MU: "Mauritius", MV: "Maldives", MW: "Malawi", MX: "Mexico", MY: "Malaysia", MZ: "Mozambique",
  NA: "Namibia", NC: "New Caledonia", NE: "Niger", NF: "Norfolk Island", NG: "Nigeria", NI: "Nicaragua",
  NL: "Netherlands", NO: "Norway", NP: "Nepal", NR: "Nauru", NU: "Niue", NZ: "New Zealand", OM: "Oman", PA: "Panama",
  PE: "Peru", PF: "French Polynesia", PG: "Papua New Guinea", PH: "Philippines", PK: "Pakistan", PL: "Poland",
  PM: "Saint Pierre and Miquelon", PR: "Puerto Rico", PS: "Palestine", PT: "Portugal", PW: "Palau", PY: "Paraguay",
  QA: "Qatar", RE: "Réunion", RO: "Romania", RS: "Serbia", RU: "Russia", RW: "Rwanda", SA: "Saudi Arabia",
  SB: "Solomon Islands", SC: "Seychelles", SD: "Sudan", SE: "Sweden", SG: "Singapore", SH: "Saint Helena",
  SI: "Slovenia", SJ: "Svalbard and Jan Mayen", SK: "Slovakia", SL: "Sierra Leone", SM: "San Marino", SN: "Senegal",
  SO: "Somalia", SR: "Suriname", SS: "South Sudan", ST: "São Tomé and Príncipe", SV: "El Salvador",
  SX: "Sint Maarten", SY: "Syria", SZ: "Eswatini", TA: "Tristan da Cunha", TC: "Turks and Caicos Islands",
  TD: "Chad", TG: "Togo", TH: "Thailand", TJ: "Tajikistan", TK: "Tokelau", TL: "Timor-Leste", TM: "Turkmenistan",
  TN: "Tunisia", TO: "Tonga", TR: "Türkiye", TT: "Trinidad and Tobago", TV: "Tuvalu", TW: "Taiwan", TZ: "Tanzania",
  UA: "Ukraine", UG: "Uganda", US: "United States", UY: "Uruguay", UZ: "Uzbekistan", VA: "Vatican City",
  VC: "Saint Vincent and Grenadines", VE: "Venezuela", VG: "British Virgin Islands", VI: "U.S. Virgin Islands",
  VN: "Vietnam", VU: "Vanuatu", WF: "Wallis and Futuna", WS: "Samoa", XK: "Kosovo", YE: "Yemen", YT: "Mayotte",
  ZA: "South Africa", ZM: "Zambia", ZW: "Zimbabwe",
};

const toCountry = (code: CountryCode): Country => ({ code, name: NAMES[code] ?? code, dial: `+${getCountryCallingCode(code)}` });

export const countryName = (code: CountryCode) => NAMES[code] ?? code;

/** Two shelves: cricket nations in their fixed order, then everyone else A–Z by name. */
export const COUNTRIES: { cricket: Country[]; others: Country[] } = (() => {
  const known = new Set<string>(getCountries());
  const cricket = CRICKET_NATIONS.filter((c) => known.has(c)).map(toCountry);
  const others = getCountries()
    .filter((c) => !CRICKET_NATIONS.includes(c))
    .map(toCountry)
    .sort((a, b) => a.name.localeCompare(b.name));
  return { cricket, others };
})();

/** Matches a name ("sri"), an ISO code ("lk") or a calling code ("94", "+94"). */
export function matchesCountry(c: Country, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  const digits = q.replace(/^\+/, '');
  if (/^\d+$/.test(digits)) return c.dial.slice(1).startsWith(digits);
  return c.name.toLowerCase().includes(q) || c.code.toLowerCase() === q || c.name.toLowerCase().split(/[\s-]+/).some((w) => w.startsWith(q));
}
