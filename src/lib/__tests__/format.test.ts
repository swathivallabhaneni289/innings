import { describe, expect, it } from '@jest/globals';

import { countdown, dayLabel, inningsLine, isLocked, pct, relative, stubNo, ticketDate } from '../format';

describe('inningsLine', () => {
  it('prints runs/wickets and overs', () => {
    expect(inningsLine({ teamId: 'ind', runs: 287, wickets: 5, overs: '48.2' })).toBe('287/5 (48.2)');
  });
  it('prints an all-out innings as runs only, as scorebooks do', () => {
    expect(inningsLine({ teamId: 'eng', runs: 412, wickets: 10, overs: '132.1' })).toBe('412 (132.1)');
  });
  it('leaves out unknown overs', () => {
    expect(inningsLine({ teamId: 'aus', runs: 99, wickets: 3, overs: '' })).toBe('99/3');
  });
});

describe('isLocked', () => {
  const start = '2026-10-22T14:00:00.000Z';
  it('stays open before the first ball', () => {
    expect(isLocked({ startTime: start, state: 'upcoming' }, Date.parse(start) - 1)).toBe(false);
  });
  it('locks at the first ball even if the state has not caught up', () => {
    expect(isLocked({ startTime: start, state: 'upcoming' }, Date.parse(start))).toBe(true);
  });
  it('is locked once live or completed', () => {
    expect(isLocked({ startTime: start, state: 'live' }, 0)).toBe(true);
    expect(isLocked({ startTime: start, state: 'completed' }, 0)).toBe(true);
  });
});

describe('countdown', () => {
  const t = Date.parse('2026-10-22T14:00:00.000Z');
  it('shows minutes and seconds under an hour', () => {
    expect(countdown('2026-10-22T14:24:10.000Z', t)).toBe('24:10');
  });
  it('shows hours when an hour or more away', () => {
    expect(countdown('2026-10-22T15:04:09.000Z', t)).toBe('1:04:09');
  });
  it('never goes negative', () => {
    expect(countdown('2026-10-22T13:00:00.000Z', t)).toBe('0:00');
  });
});

describe('ticket text', () => {
  it('prints a classic on its recorded date, whatever the time zone', () => {
    expect(ticketDate({ startTime: '2011-04-02T12:00:00Z', classic: { date: '2011-04-02' } as never })).toBe('02 APR 2011');
  });
  it('pads stub numbers to four digits', () => {
    expect(stubNo(42)).toBe('No. 0042');
  });
  it('rounds shares to whole percentages', () => {
    expect(pct(0.415)).toBe('42%');
  });
});

describe('relative and dayLabel', () => {
  const now = new Date(2026, 9, 22, 12, 0, 0).getTime();
  it('says how far away a time is', () => {
    expect(relative(new Date(now + 25 * 60_000).toISOString(), now)).toBe('in 25 min');
    expect(relative(new Date(now - 2 * 3_600_000).toISOString(), now)).toBe('2 h ago');
  });
  it('names today, tomorrow and yesterday', () => {
    expect(dayLabel(new Date(2026, 9, 22, 19, 30).toISOString(), now)).toBe('Today');
    expect(dayLabel(new Date(2026, 9, 23, 9, 0).toISOString(), now)).toBe('Tomorrow');
    expect(dayLabel(new Date(2026, 9, 21, 9, 0).toISOString(), now)).toBe('Yesterday');
  });
});
