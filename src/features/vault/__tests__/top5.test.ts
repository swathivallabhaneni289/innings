import { beforeEach, describe, expect, it, jest } from '@jest/globals';
import { act, createElement } from 'react';
import { create, type ReactTestRenderer } from 'react-test-renderer';

import type { ClassicInfo } from '@/domain/types';
import { anniversary, competitionNoYear, matchYear, useTop5 } from '../top5';

const mockSetTop5 = jest.fn<(matchIds: string[]) => Promise<unknown>>();
jest.mock('@/data', () => ({ useActions: () => ({ setTop5: mockSetTop5 }) }));

const classic = (date: string): ClassicInfo => ({
  date,
  competitionName: 'ICC Cricket World Cup 2011',
  headline: 'A night to remember',
  scoreline: '',
  collections: [],
  era: '2010s',
});

describe('matchYear', () => {
  it('prints a classic’s year from its recorded date, whatever the start time says', () => {
    // A start time that reads late on New Year's Eve on this clock: the recorded date still says 2011.
    expect(matchYear({ startTime: new Date(2010, 11, 31, 23, 30).toISOString(), classic: classic('2011-01-01') })).toBe('2011');
  });
  it('reads a fixture’s year off its start time, on the fan’s own calendar', () => {
    expect(matchYear({ startTime: new Date(2026, 0, 1, 0, 30).toISOString() })).toBe('2026');
  });
});

describe('competitionNoYear', () => {
  it('drops a trailing year or season', () => {
    expect(competitionNoYear("ICC Men's T20 World Cup 2022")).toBe("ICC Men's T20 World Cup");
    expect(competitionNoYear('Big Bash League 2020-21')).toBe('Big Bash League');
    expect(competitionNoYear('The Ashes 2019-2020')).toBe('The Ashes');
  });
  it('leaves names without a year at the end alone', () => {
    expect(competitionNoYear('Indian Premier League')).toBe('Indian Premier League');
    expect(competitionNoYear('1983 World Cup Final')).toBe('1983 World Cup Final');
    expect(competitionNoYear('World Cup 2011 Final')).toBe('World Cup 2011 Final');
  });
});

describe('anniversary', () => {
  it('prints a classic’s day from its recorded date, without a leading zero', () => {
    // A start time that reads 1 April on this clock: the recorded date wins.
    expect(anniversary({ startTime: new Date(2011, 3, 1, 23, 0).toISOString(), classic: classic('2011-04-02') })).toBe('2 Apr');
  });
  it('names the first and last months', () => {
    expect(anniversary({ startTime: '2021-01-19T00:00:00.000Z', classic: classic('2021-01-19') })).toBe('19 Jan');
    expect(anniversary({ startTime: '2019-12-31T00:00:00.000Z', classic: classic('2019-12-31') })).toBe('31 Dec');
  });
  it('reads a fixture’s day off its start time, on the fan’s own calendar', () => {
    expect(anniversary({ startTime: new Date(2026, 9, 22, 19, 30).toISOString() })).toBe('22 Oct');
  });
});

type Top5 = ReturnType<typeof useTop5>;

function Probe({ ids, onRender }: { ids: string[]; onRender: (top5: Top5) => void }) {
  onRender(useTop5(ids));
  return null;
}

/** Renders useTop5 over a Top 5 (match ids in order). `current` is what the hook returned on its latest render. */
function renderTop5(ids: string[]) {
  const latest: { value?: Top5 } = {};
  const onRender = (top5: Top5) => {
    latest.value = top5;
  };
  let root!: ReactTestRenderer;
  act(() => {
    root = create(createElement(Probe, { ids, onRender }));
  });
  return {
    get current(): Top5 {
      if (!latest.value) throw new Error('useTop5 has not rendered');
      return latest.value;
    },
    rerender(next: string[]) {
      act(() => root.update(createElement(Probe, { ids: next, onRender })));
    },
  };
}

const FULL = ['a', 'b', 'c', 'd', 'e'];

describe('useTop5', () => {
  beforeEach(() => {
    mockSetTop5.mockReset();
    mockSetTop5.mockResolvedValue({ top5: [] });
  });

  it('pins a match after the ones already there', async () => {
    const pin = renderTop5(['a', 'b']);
    await act(async () => {
      await pin.current.toggle('c');
    });
    expect(mockSetTop5).toHaveBeenCalledWith(['a', 'b', 'c']);
  });

  it('unpins a match and keeps the rest in order', async () => {
    const pin = renderTop5(['a', 'b', 'c']);
    expect(pin.current.isPinned('b')).toBe(true);
    await act(async () => {
      await pin.current.toggle('b');
    });
    expect(mockSetTop5).toHaveBeenLastCalledWith(['a', 'c']);
    await act(async () => {
      await pin.current.unpin('a');
    });
    expect(mockSetTop5).toHaveBeenLastCalledWith(['b', 'c']);
  });

  it('refuses a sixth pin without asking the server', async () => {
    const pin = renderTop5(FULL);
    expect(pin.current.full).toBe(true);
    await act(async () => {
      await pin.current.toggle('f');
    });
    expect(mockSetTop5).not.toHaveBeenCalled();
    expect(pin.current.message).toEqual({ matchId: 'f', text: 'Your Top 5 is full. Tap Edit on it in the Vault to take one out.' });
  });

  it('still unpins when the Top 5 is full', async () => {
    const pin = renderTop5(FULL);
    await act(async () => {
      await pin.current.toggle('c');
    });
    expect(mockSetTop5).toHaveBeenCalledWith(['a', 'b', 'd', 'e']);
  });

  it('marks the match being saved as busy until the server answers', async () => {
    let answer = () => {};
    mockSetTop5.mockReturnValue(
      new Promise((resolve) => {
        answer = () => resolve({ top5: [] });
      }),
    );
    const pin = renderTop5(['a']);
    let saving: Promise<void> | undefined;
    act(() => {
      saving = pin.current.toggle('b');
    });
    expect(pin.current.busyId).toBe('b');
    await act(async () => {
      answer();
      await saving;
    });
    expect(pin.current.busyId).toBeNull();
    expect(pin.current.message).toBeNull();
  });

  it('reports a failed save against the match that failed', async () => {
    mockSetTop5.mockRejectedValue(new Error('The groundsman has the covers on.'));
    const pin = renderTop5(['a']);
    await act(async () => {
      await pin.current.toggle('b');
    });
    expect(pin.current.message).toEqual({ matchId: 'b', text: 'The groundsman has the covers on.' });
    expect(pin.current.busyId).toBeNull();
  });

  it('says something plain when the failure carries no message', async () => {
    mockSetTop5.mockRejectedValue('timeout');
    const pin = renderTop5(['a']);
    await act(async () => {
      await pin.current.toggle('b');
    });
    expect(pin.current.message).toEqual({ matchId: 'b', text: 'Could not update your Top 5.' });
  });

  it('clears an old message once the next save starts', async () => {
    const pin = renderTop5(FULL);
    await act(async () => {
      await pin.current.toggle('f');
    });
    expect(pin.current.message?.matchId).toBe('f');
    await act(async () => {
      await pin.current.toggle('a');
    });
    expect(pin.current.message).toBeNull();
  });

  it('follows the Top 5 it is given as that changes', () => {
    const pin = renderTop5(['a']);
    expect(pin.current.isPinned('b')).toBe(false);
    expect(pin.current.full).toBe(false);
    pin.rerender(FULL);
    expect(pin.current.isPinned('b')).toBe(true);
    expect(pin.current.full).toBe(true);
  });
});
