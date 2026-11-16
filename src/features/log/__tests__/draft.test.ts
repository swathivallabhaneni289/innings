import { describe, expect, it } from '@jest/globals';

import { RATING_STEPS } from '@/domain/constants';
import type { MatchLog } from '@/domain/types';
import { draftFromLog, emptyDraft, lastPlaceOf, placeLine, relog, toNewLog, verdict, type Draft } from '../draft';

const stub: MatchLog = {
  id: 'log-1',
  userId: 'u-1',
  matchId: 'm-1',
  rating: 4.5,
  instantClassic: true,
  playerRatings: { 'p-kohli': 9, 'p-starc': 6 },
  fanPotmId: 'p-kohli',
  watchedAt: 'hostel',
  placeNote: 'Hostel room',
  watchedWith: ['u-2', 'u-3'],
  companions: 6,
  photoUri: 'file:///stub-photo.jpg',
  review: 'What a chase.',
  createdAt: '2026-10-22T18:00:00.000Z',
  stubNo: 42,
};

const xi = new Set(['p-kohli', 'p-starc', 'p-root', 'p-wood']);
const draft = (patch: Partial<Draft>): Draft => ({ ...emptyDraft('m-1'), ...patch });

describe('emptyDraft', () => {
  it('starts unrated, with nothing picked and the place left to the default', () => {
    expect(emptyDraft('m-1')).toEqual({
      matchId: 'm-1',
      rating: 0,
      instantClassic: false,
      playerRatings: {},
      fanPotmId: null,
      watchedAt: null,
      placeNote: '',
      watchedWith: [],
      companions: 0,
      photoUri: undefined,
      review: '',
    });
  });
  it('can start before a match is picked', () => {
    expect(emptyDraft(null).matchId).toBeNull();
  });
});

describe('draftFromLog', () => {
  it('fills the draft from the stub being re-logged', () => {
    expect(draftFromLog(stub)).toEqual({
      matchId: 'm-1',
      rating: 4.5,
      instantClassic: true,
      playerRatings: { 'p-kohli': 9, 'p-starc': 6 },
      fanPotmId: 'p-kohli',
      watchedAt: 'hostel',
      placeNote: 'Hostel room',
      watchedWith: ['u-2', 'u-3'],
      companions: 6,
      photoUri: 'file:///stub-photo.jpg',
      review: 'What a chase.',
    });
  });

  it('fills what a sparse stub left out with blanks', () => {
    const sparse: MatchLog = { ...stub, placeNote: undefined, companions: undefined, photoUri: undefined, review: undefined };
    expect(draftFromLog(sparse)).toMatchObject({ placeNote: '', companions: 0, photoUri: undefined, review: '' });
  });

  it('copies the ratings and tagged friends, so editing the draft never touches the cached stub', () => {
    const d = draftFromLog(stub);
    d.playerRatings['p-root'] = 7;
    d.watchedWith.push('u-9');
    expect(stub.playerRatings).toEqual({ 'p-kohli': 9, 'p-starc': 6 });
    expect(stub.watchedWith).toEqual(['u-2', 'u-3']);
  });
});

describe('toNewLog', () => {
  it('keeps whole-number player ratings, only for players in the XI', () => {
    const d = draft({ playerRatings: { 'p-kohli': 7.4, 'p-starc': 7.5, 'p-root': 1, 'p-wood': 0.9, 'p-sub': 8 } });
    // 0.9 would round up to 1, but anything under 1 is a skipped player; p-sub never made the XI.
    expect(toNewLog(d, 'm-1', 'home', xi).playerRatings).toEqual({ 'p-kohli': 7, 'p-starc': 8, 'p-root': 1 });
  });

  it('drops a fan POTM who is not in the XI', () => {
    expect(toNewLog(draft({ fanPotmId: 'p-sub' }), 'm-1', 'home', xi).fanPotmId).toBeNull();
    expect(toNewLog(draft({ fanPotmId: 'p-root' }), 'm-1', 'home', xi).fanPotmId).toBe('p-root');
    expect(toNewLog(draft({ fanPotmId: null }), 'm-1', 'home', xi).fanPotmId).toBeNull();
  });

  it('trims the place note and review, and leaves blank ones out', () => {
    const typed = toNewLog(draft({ placeNote: '  Uncle’s place ', review: '\n What a finish!  ' }), 'm-1', 'home', xi);
    expect(typed).toMatchObject({ placeNote: 'Uncle’s place', review: 'What a finish!' });
    const blank = toNewLog(draft({ placeNote: '   ', review: '\n\n' }), 'm-1', 'home', xi);
    expect(blank.placeNote).toBeUndefined();
    expect(blank.review).toBeUndefined();
  });

  it('leaves out the head count when nobody else was there', () => {
    expect(toNewLog(draft({ companions: 0 }), 'm-1', 'home', xi).companions).toBeUndefined();
    expect(toNewLog(draft({ companions: 3 }), 'm-1', 'home', xi).companions).toBe(3);
  });

  it('saves the place the flow settled on when the draft kept the default', () => {
    const d = draft({ watchedAt: null });
    expect(toNewLog(d, 'm-1', 'office', xi).watchedAt).toBe('office');
  });

  it('gives back the same stub when a re-log changes nothing, photo included', () => {
    // The log step no longer offers a photo, so a stub's existing one has to ride along untouched.
    expect(toNewLog(draftFromLog(stub), stub.matchId, stub.watchedAt, xi)).toEqual({
      matchId: 'm-1',
      rating: 4.5,
      instantClassic: true,
      playerRatings: { 'p-kohli': 9, 'p-starc': 6 },
      fanPotmId: 'p-kohli',
      watchedAt: 'hostel',
      placeNote: 'Hostel room',
      watchedWith: ['u-2', 'u-3'],
      companions: 6,
      photoUri: 'file:///stub-photo.jpg',
      review: 'What a chase.',
    });
  });
});

describe('verdict', () => {
  it('has nothing to say before a rating', () => {
    expect(verdict(0)).toBe('Not rated yet');
    expect(verdict(-1)).toBe('Not rated yet');
  });

  it('gives every half-ball step its own word, from a dead rubber to an all-timer', () => {
    const words = RATING_STEPS.map(verdict);
    expect(new Set(words).size).toBe(RATING_STEPS.length);
    expect(words[0]).toBe('Dead rubber');
    expect(verdict(2.5)).toBe('Par score');
    expect(words[words.length - 1]).toBe('All-timer');
  });

  it('reads an in-between rating as the nearest half ball', () => {
    expect(verdict(4.2)).toBe(verdict(4));
    expect(verdict(4.3)).toBe(verdict(4.5));
  });

  it('stays on the scale at both ends', () => {
    expect(verdict(0.1)).toBe('Dead rubber');
    expect(verdict(7)).toBe('All-timer');
  });
});

describe('placeLine', () => {
  it('prints the fan’s own words for the place, with the head count', () => {
    expect(placeLine({ watchedAt: 'hostel', placeNote: '  Hostel room ', companions: 6 })).toBe('📍 Hostel room, with 6 people');
  });

  it('says person for one', () => {
    expect(placeLine({ watchedAt: 'friends', companions: 1 })).toBe("📍 Friend's place, with 1 person");
  });

  it('falls back to the place’s name when the note is blank, and leaves out an empty room', () => {
    expect(placeLine({ watchedAt: 'cafe', placeNote: '   ', companions: 0 })).toBe('📍 Café or pub');
    expect(placeLine({ watchedAt: 'home' })).toBe('📍 Home');
  });
});

describe('lastPlaceOf', () => {
  it('is the place of the newest log, whatever order they come in', () => {
    const logs = [
      { watchedAt: 'home' as const, createdAt: '2026-10-20T18:00:00.000Z' },
      { watchedAt: 'stadium' as const, createdAt: '2026-10-22T18:00:00.000Z' },
      { watchedAt: 'office' as const, createdAt: '2026-10-21T18:00:00.000Z' },
    ];
    expect(lastPlaceOf(logs)).toBe('stadium');
  });
  it('is null before the first log', () => {
    expect(lastPlaceOf([])).toBeNull();
  });
});

describe('relog', () => {
  it('changes only the rating on a stub, keeping players, place, friends and review', () => {
    expect(relog(stub, 'm-1', 'home', { rating: 3 })).toEqual({
      matchId: 'm-1',
      rating: 3,
      instantClassic: true,
      playerRatings: { 'p-kohli': 9, 'p-starc': 6 },
      fanPotmId: 'p-kohli',
      watchedAt: 'hostel',
      placeNote: 'Hostel room',
      watchedWith: ['u-2', 'u-3'],
      companions: 6,
      photoUri: 'file:///stub-photo.jpg',
      review: 'What a chase.',
    });
  });
  it('changes only the review, trimmed, and an emptied review is taken off', () => {
    const edited = relog(stub, 'm-1', 'home', { review: '  Still the best chase I have seen.  ' });
    expect(edited.review).toBe('Still the best chase I have seen.');
    expect(edited.rating).toBe(4.5);
    expect(edited.watchedAt).toBe('hostel');
    expect(relog(stub, 'm-1', 'home', { review: '   ' }).review).toBeUndefined();
  });
  it('starts a first log from the tapped rating, watched at the given place', () => {
    expect(relog(null, 'm-9', 'friends', { rating: 4 })).toEqual({
      matchId: 'm-9',
      rating: 4,
      instantClassic: false,
      playerRatings: {},
      fanPotmId: null,
      watchedAt: 'friends',
      placeNote: undefined,
      watchedWith: [],
      companions: undefined,
      photoUri: undefined,
      review: undefined,
    });
    expect(relog(null, 'm-9', 'home', { rating: 2.5, review: 'Rain spoiled it.' }).review).toBe('Rain spoiled it.');
  });
});
