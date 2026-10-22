// Profile and Settings constants: the teams a fan can back, and the avatar colours they can pick.
import { paper } from '@/theme';

/** Favourite-team choices in Settings (looked up with getTeam). Franchises show their short code on chips. */
export const FAV_TEAMS = {
  international: ['ind', 'aus', 'eng', 'sa', 'nz', 'pak', 'sl', 'wi', 'ind-w', 'aus-w'],
  franchise: ['csk', 'mi', 'rcb', 'kkr'],
} as const;

/**
 * Avatar colours. These are stored on the user and drawn behind cream initials in both themes,
 * so they come from the fixed paper palette (never useTheme(), whose ink turns cream under floodlights).
 * Pitch green is left out (it means wins only); leather and heather are muted tones that sit beside ink and red.
 */
export const AVATAR_SWATCHES: { colour: string; name: string }[] = [
  { colour: paper.red, name: 'Ball red' },
  { colour: paper.ink, name: 'Ink navy' },
  { colour: paper.inkSoft, name: 'Slate' },
  { colour: '#5A3E36', name: 'Leather' },
  { colour: '#6B4E71', name: 'Heather' },
];

export const BIO_MAX = 60;
export const NAME_MAX = 40;

/** "Aditi" from "Aditi Rao". */
export const firstName = (name: string) => name.trim().split(/\s+/)[0] || name;

/** 1240 → "1,240". */
export const thousands = (n: number) => String(Math.round(n)).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
