// "Matchday Ticket": concrete ground, flat white cards, every match printed on coloured ticket stock.
// Ink for text and primary actions, overprint red for stamps, live and serials, pitch green only for wins.
import type { Match } from '@/domain/types';

export interface Palette {
  mode: 'paper' | 'floodlights';
  paper: string; // screen ground (concrete)
  card: string; // flat cards, the tab bar
  cardEdge: string; // perforations, ticket dots
  ink: string; // primary text, primary buttons
  onInk: string; // text on ink
  inkSoft: string; // secondary text
  inkFaint: string; // tertiary text, disabled
  rule: string; // hairlines, chip outlines
  red: string; // overprint red: stamps, live, serials, accents
  redSoft: string;
  onRed: string;
  green: string; // wins / success only
  greenSoft: string;
  seam: string; // ball seam stitching
  shadow: string;
  stock: Record<StockKey, string>; // ticket card stock per competition
}

export type StockKey =
  | 'international'
  | 'test'
  | 'ipl'
  | 'wpl'
  | 'world-cup'
  | 'bbl'
  | 'sa20'
  | 'the-hundred'
  | 'classic';

export const paper: Palette = {
  mode: 'paper',
  paper: '#ECEAE4',
  card: '#FFFFFF',
  cardEdge: 'rgba(22,23,26,0.30)',
  ink: '#16171A',
  onInk: '#FFFFFF',
  inkSoft: '#55575E',
  inkFaint: '#8B8D93',
  rule: 'rgba(22,23,26,0.13)',
  red: '#D23A2E',
  redSoft: 'rgba(210,58,46,0.10)',
  onRed: '#FFFFFF',
  green: '#2F6B3A',
  greenSoft: 'rgba(47,107,58,0.12)',
  seam: '#FBE3D2',
  shadow: 'rgba(22,23,26,0.10)',
  stock: {
    international: '#D7E4EE',
    test: '#F3E7C9',
    ipl: '#F4D6C8',
    wpl: '#E6DCF0',
    'world-cup': '#EFE3A8',
    bbl: '#D6E8E2',
    sa20: '#F2D9E2',
    'the-hundred': '#DCDDF0',
    classic: '#E8E0CC',
  },
};

export const floodlights: Palette = {
  mode: 'floodlights',
  paper: '#121316',
  card: '#1D1F23',
  cardEdge: 'rgba(241,239,234,0.30)',
  ink: '#F1EFEA',
  onInk: '#121316',
  inkSoft: '#B4B5B9',
  inkFaint: '#7D7F86',
  rule: 'rgba(241,239,234,0.12)',
  red: '#E5584B',
  redSoft: 'rgba(229,88,75,0.16)',
  onRed: '#FFFFFF',
  green: '#6DB27A',
  greenSoft: 'rgba(109,178,122,0.16)',
  seam: '#FBE3D2',
  shadow: 'rgba(0,0,0,0.4)',
  stock: {
    international: '#223142',
    test: '#3A3224',
    ipl: '#402A23',
    wpl: '#32293F',
    'world-cup': '#3A3420',
    bbl: '#203630',
    sa20: '#3D2731',
    'the-hundred': '#2A2B42',
    classic: '#35302A',
  },
};

/** The card stock a match's ticket is printed on. */
export function stockFor(m: Pick<Match, 'kind' | 'format' | 'competitionId'>, p: Palette): string {
  if (m.kind === 'classic') return p.stock.classic;
  if (m.format === 'Test') return p.stock.test;
  return p.stock[m.competitionId as StockKey] ?? p.stock.international;
}

export const space = { xs: 4, sm: 8, md: 12, lg: 16, xl: 24, xxl: 32 } as const;

/** Tickets and cards are square-cut; only balls, stamps and avatars are round. */
export const radius = { none: 0, hair: 2, sm: 4, round: 999 } as const;

// Keys are roles (kept stable for feature code); values are the Matchday Ticket faces.
export const fonts = {
  display: 'BigShouldersDisplay_900Black', // team codes, big numerals
  heading: 'SchibstedGrotesk_800ExtraBold',
  label: 'SchibstedGrotesk_600SemiBold',
  condensed: 'BigShouldersDisplay_800ExtraBold',
  serif: 'SchibstedGrotesk_400Regular', // body and reviews (role name kept for compatibility)
  serifItalic: 'SchibstedGrotesk_400Regular_Italic',
  serifBold: 'SchibstedGrotesk_600SemiBold',
  ui: 'SchibstedGrotesk_500Medium',
  uiBold: 'SchibstedGrotesk_700Bold',
  mono: 'IBMPlexMono_400Regular',
  monoMedium: 'IBMPlexMono_500Medium',
  monoBold: 'IBMPlexMono_600SemiBold',
} as const;

/** Type scale. Big Shoulders = codes and numbers, Schibsted Grotesk = everything you read, Plex Mono = ticket fine print and stats. */
export const type = {
  display: { fontFamily: fonts.display, fontSize: 44, lineHeight: 44, letterSpacing: 0.2 },
  h1: { fontFamily: fonts.heading, fontSize: 26, lineHeight: 31, letterSpacing: -0.4 },
  h2: { fontFamily: fonts.uiBold, fontSize: 19, lineHeight: 24, letterSpacing: -0.2 },
  h3: { fontFamily: fonts.uiBold, fontSize: 15, lineHeight: 20, letterSpacing: -0.1 },
  label: { fontFamily: fonts.monoMedium, fontSize: 10.5, lineHeight: 14, letterSpacing: 0.8, textTransform: 'uppercase' as const },
  score: { fontFamily: fonts.condensed, fontSize: 26, lineHeight: 28, letterSpacing: 0.2 },
  body: { fontFamily: fonts.serif, fontSize: 15, lineHeight: 22 },
  bodySmall: { fontFamily: fonts.serif, fontSize: 13, lineHeight: 18 },
  review: { fontFamily: fonts.serif, fontSize: 16, lineHeight: 23 },
  mono: { fontFamily: fonts.monoMedium, fontSize: 13, lineHeight: 18 },
  monoSmall: { fontFamily: fonts.mono, fontSize: 11.5, lineHeight: 15 },
  stat: { fontFamily: fonts.condensed, fontSize: 30, lineHeight: 32 },
} as const;

export type TypeVariant = keyof typeof type;
