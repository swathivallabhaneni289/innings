import { describe, expect, it } from '@jest/globals';

import { listLine, makeMain, sameTeams, toggleTeam } from '../teams';

describe('toggleTeam', () => {
  it('adds new teams to the end and drops picked ones', () => {
    expect(toggleTeam(['ind'], 'rcb')).toEqual({ ids: ['ind', 'rcb'], full: false });
    expect(toggleTeam(['ind', 'rcb', 'mi'], 'rcb')).toEqual({ ids: ['ind', 'mi'], full: false });
  });

  it('lets the next team take over when the main one is dropped', () => {
    expect(toggleTeam(['ind', 'rcb'], 'ind').ids).toEqual(['rcb']);
  });

  it('stops at five and says so', () => {
    const five = ['ind', 'aus', 'eng', 'rcb', 'mi'];
    expect(toggleTeam(five, 'csk')).toEqual({ ids: five, full: true });
    expect(toggleTeam(five, 'mi')).toEqual({ ids: ['ind', 'aus', 'eng', 'rcb'], full: false });
  });
});

describe('makeMain', () => {
  it('moves a picked team to the front and leaves the rest in order', () => {
    expect(makeMain(['ind', 'rcb', 'mi'], 'mi')).toEqual(['mi', 'ind', 'rcb']);
    expect(makeMain(['ind', 'rcb'], 'ind')).toEqual(['ind', 'rcb']);
  });

  it('ignores a team that is not picked', () => {
    expect(makeMain(['ind'], 'aus')).toEqual(['ind']);
  });
});

describe('listLine and sameTeams', () => {
  it('reads like a sentence', () => {
    expect(listLine([])).toBe('');
    expect(listLine(['India'])).toBe('India');
    expect(listLine(['India', 'RCB'])).toBe('India and RCB');
    expect(listLine(['India', 'RCB', 'MI'])).toBe('India, RCB and MI');
  });

  it('compares order as well as members', () => {
    expect(sameTeams(['ind', 'rcb'], ['ind', 'rcb'])).toBe(true);
    expect(sameTeams(['ind', 'rcb'], ['rcb', 'ind'])).toBe(false);
    expect(sameTeams(['ind'], ['ind', 'rcb'])).toBe(false);
  });
});
