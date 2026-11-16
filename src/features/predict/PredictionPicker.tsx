// CONTRACT (owned by the Home agent): props below are fixed; the Match Page imports this too.
import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { Fragment, useMemo, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';

import { Button, Chip, TeamBadge, Txt } from '@/components';
import { getTeam, isLockedError, playerName, squads, useActions, useNow, type Picks } from '@/data';
import type { PredictionSplit } from '@/domain/api';
import type { Match, Player, PlayerRole, Prediction } from '@/domain/types';
import { isLocked, pct } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';

export interface PredictionPickerProps {
  match: Match;
  mine: Prediction | null; // current saved picks, if any
  split?: PredictionSplit | null; // shows the friends' split for the winner pick when given
  locked: boolean; // true from the first ball: show the picks read-only with "Locked 🔒"
  compact?: boolean; // Home "Your Call" card variant
  onSaved?: () => void;
}

type Step = 0 | 1 | 2 | 'done';
type Status = { kind: 'idle' } | { kind: 'saving' } | { kind: 'saved' } | { kind: 'error'; message: string };

const FIELDS = [
  { key: 'winnerId', tab: 'Winner', ask: 'Who wins it?', row: 'Winner' },
  { key: 'topRunScorerId', tab: 'Most runs', ask: 'Who makes the most runs?', row: 'Top run-scorer' },
  { key: 'topWicketTakerId', tab: 'Most wickets', ask: 'Who takes the most wickets?', row: 'Top wicket-taker' },
] as const;

const toPicks = (p: Prediction | null): Partial<Picks> =>
  p ? { winnerId: p.winnerId, topRunScorerId: p.topRunScorerId, topWicketTakerId: p.topWicketTakerId } : {};
const isComplete = (p: Partial<Picks>): p is Picks => !!(p.winnerId && p.topRunScorerId && p.topWicketTakerId);
const samePicks = (a: Partial<Picks>, b: Partial<Picks> | null) =>
  !!b && a.winnerId === b.winnerId && a.topRunScorerId === b.topRunScorerId && a.topWicketTakerId === b.topWicketTakerId;
const firstMissing = (p: Partial<Picks>): 0 | 1 | 2 => (!p.winnerId ? 0 : !p.topRunScorerId ? 1 : 2);
const surname = (id: string | undefined) => (id ? playerName(id).split(' ').slice(-1)[0] : '');

// Bowlers first for the wicket-taker pick, so the likely names are the first taps.
const BOWLING_ORDER: Record<PlayerRole, number> = { BOWL: 0, AR: 1, BAT: 2, WK: 3 };

/** Winner, top run-scorer and top wicket-taker in three quick taps. */
export function PredictionPicker(props: PredictionPickerProps) {
  // Keyed by match: when the slot moves on to another match (Home's "Your Call"), the picker starts over from
  // that match's saved picks, with nothing carried across from the last one.
  return <Picker key={props.match.id} {...props} />;
}

function Picker({ match, mine, split, locked, compact = false, onSaved }: PredictionPickerProps) {
  const now = useNow(5_000);
  const { predict } = useActions();
  const [picks, setPicks] = useState<Partial<Picks>>(() => toPicks(mine));
  const [saved, setSaved] = useState<Picks | null>(() => (mine ? (toPicks(mine) as Picks) : null));
  const [step, setStep] = useState<Step>(() => (mine ? 'done' : 0));
  const [status, setStatus] = useState<Status>({ kind: 'idle' });
  const [lockedByServer, setLockedByServer] = useState(false);

  // Follow the server's copy of my picks, unless the fan is mid-change.
  const mineKey = mine ? `${mine.winnerId}|${mine.topRunScorerId}|${mine.topWicketTakerId}` : '';
  const [syncedKey, setSyncedKey] = useState(mineKey);
  if (syncedKey !== mineKey) {
    setSyncedKey(mineKey);
    if (mine) {
      const server = toPicks(mine) as Picks;
      const dirty = !samePicks(picks, saved) && Object.keys(picks).length > 0;
      setSaved(server);
      if (!dirty) {
        setPicks(server);
        setStep('done');
      }
    }
  }

  const save = async (p: Picks) => {
    setStatus({ kind: 'saving' });
    try {
      await predict(match.id, p);
      setSaved(p);
      setStatus({ kind: 'saved' });
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => {});
      onSaved?.();
    } catch (e) {
      if (isLockedError(e)) {
        setLockedByServer(true);
        setStatus({ kind: 'idle' });
      } else {
        setStatus({ kind: 'error', message: e instanceof Error ? e.message : 'Could not save your call.' });
      }
    }
  };

  const choose = (key: keyof Picks, id: string) => {
    Haptics.selectionAsync().catch(() => {});
    const next = { ...picks, [key]: id };
    setPicks(next);
    if (isComplete(next)) {
      setStep('done');
      if (!samePicks(next, saved)) void save(next);
      else setStatus({ kind: 'idle' });
    } else {
      setStep(firstMissing(next));
    }
  };

  if (locked || lockedByServer || isLocked(match, now)) {
    return <LockedPicks match={match} picks={saved ?? (mine ? (toPicks(mine) as Picks) : null)} compact={compact} />;
  }

  return (
    <View>
      <StepTabs picks={picks} step={step} onStep={setStep} compact={compact} />
      <View style={{ marginTop: compact ? 12 : 16 }}>
        {step === 0 ? (
          <WinnerStep match={match} split={split} selected={picks.winnerId} compact={compact} onPick={(id) => choose('winnerId', id)} />
        ) : step === 1 || step === 2 ? (
          <PlayerStep
            match={match}
            step={step}
            selected={step === 1 ? picks.topRunScorerId : picks.topWicketTakerId}
            compact={compact}
            onPick={(id) => choose(step === 1 ? 'topRunScorerId' : 'topWicketTakerId', id)}
          />
        ) : (
          <Confirmation status={status} onRetry={() => isComplete(picks) && save(picks)} />
        )}
      </View>
      {status.kind === 'error' && step !== 'done' ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
          {status.message}
        </Txt>
      ) : null}
    </View>
  );
}

/** Three tabs: the step, and what's been picked so far. Tap one to change it. */
function StepTabs({ picks, step, onStep, compact }: { picks: Partial<Picks>; step: Step; onStep: (s: Step) => void; compact: boolean }) {
  const t = useTheme();
  const values = [picks.winnerId ? getTeam(picks.winnerId).short : '', surname(picks.topRunScorerId), surname(picks.topWicketTakerId)];
  return (
    <View style={[styles.tabs, { borderBottomColor: t.rule }]}>
      {FIELDS.map((f, i) => {
        const active = step === i;
        return (
          <Pressable
            key={f.key}
            accessibilityRole="tab"
            accessibilityState={{ selected: active }}
            onPress={() => onStep(i as 0 | 1 | 2)}
            style={[styles.tab, { borderBottomColor: active ? t.ink : 'transparent' }]}>
            <Txt variant="bodySmall" tone={active ? 'ink' : 'soft'} numberOfLines={1} style={{ fontFamily: active ? fonts.uiBold : fonts.ui }}>
              {f.tab}
            </Txt>
            <Txt
              variant="bodySmall"
              tone={values[i] ? 'ink' : 'faint'}
              numberOfLines={1}
              style={{ fontFamily: values[i] ? fonts.uiBold : fonts.ui, fontSize: compact ? 14 : 15, lineHeight: 20, marginTop: 1 }}>
              {values[i] || '—'}
            </Txt>
          </Pressable>
        );
      })}
    </View>
  );
}

function WinnerStep({
  match,
  split,
  selected,
  compact,
  onPick,
}: {
  match: Match;
  split?: PredictionSplit | null;
  selected?: string;
  compact: boolean;
  onPick: (teamId: string) => void;
}) {
  const t = useTheme();
  const stock = stockFor(match, t);
  return (
    <View>
      {!compact ? (
        <Txt variant="h3" style={{ marginBottom: 10 }}>
          {FIELDS[0].ask}
        </Txt>
      ) : null}
      <View style={styles.teamRow}>
        {match.teamIds.map((id, i) => {
          const team = getTeam(id);
          const on = selected === id;
          const fg = on ? t.onInk : t.ink;
          return (
            <Pressable
              key={id}
              accessibilityRole="button"
              accessibilityState={{ selected: on }}
              accessibilityLabel={`${team.name} to win`}
              onPress={() => onPick(id)}
              style={({ pressed }) => [
                styles.teamButton,
                {
                  marginLeft: i === 1 ? 10 : 0,
                  backgroundColor: on ? t.ink : stock,
                  opacity: pressed ? 0.85 : 1,
                },
              ]}>
              <View style={{ height: 3, alignSelf: 'stretch', backgroundColor: team.colour }} />
              <View style={{ paddingTop: compact ? 10 : 14, paddingBottom: compact ? 12 : 16, paddingHorizontal: 10, alignItems: 'center' }}>
                <Txt variant="display" numberOfLines={1} style={{ color: fg, fontSize: compact ? 28 : 36, lineHeight: compact ? 32 : 40 }}>
                  {team.short}
                </Txt>
                <Txt variant="bodySmall" numberOfLines={2} style={{ color: on ? t.onInk : t.inkSoft, marginTop: 2, textAlign: 'center' }}>
                  {team.name}
                </Txt>
                {split ? (
                  <Txt variant="bodySmall" style={{ color: on ? t.onInk : t.inkSoft, marginTop: 4 }}>
                    Friends{' '}
                    <Txt variant="mono" style={{ fontSize: 12.5, color: on ? t.onInk : t.ink }}>
                      {split.friendsTotal > 0 ? pct(split.friendShares[id] ?? 0) : '–'}
                    </Txt>
                  </Txt>
                ) : null}
              </View>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

function PlayerStep({
  match,
  step,
  selected,
  compact,
  onPick,
}: {
  match: Match;
  step: 1 | 2;
  selected?: string;
  compact: boolean;
  onPick: (playerId: string) => void;
}) {
  const groups = useMemo(
    () =>
      squads(match).map((g) => ({
        team: g.team,
        players:
          step === 2
            ? g.players
                .map((p, i) => ({ p, i }))
                .sort((a, b) => BOWLING_ORDER[a.p.role] - BOWLING_ORDER[b.p.role] || a.i - b.i)
                .map(({ p }) => p)
            : g.players,
      })),
    [match, step],
  );
  const ask = FIELDS[step].ask;
  if (groups.every((g) => g.players.length === 0)) {
    return (
      <Txt variant="bodySmall" tone="soft">
        The XIs aren&apos;t out yet. Check back at the toss.
      </Txt>
    );
  }
  const chip = (p: Player) => <Chip key={p.id} label={p.name} selected={selected === p.id} onPress={() => onPick(p.id)} />;
  if (compact) {
    return (
      <View>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui, marginBottom: 8 }}>
          {ask}
        </Txt>
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.chipStrip}>
          {groups.map((g, gi) => (
            <Fragment key={g.team.id}>
              <View style={{ marginRight: 8, marginLeft: gi > 0 ? 8 : 0 }}>
                <TeamBadge teamId={g.team.id} size="sm" />
              </View>
              {g.players.map(chip)}
            </Fragment>
          ))}
        </ScrollView>
      </View>
    );
  }
  return (
    <View>
      <Txt variant="h3" style={{ marginBottom: 10 }}>
        {ask}
      </Txt>
      {groups.map((g) => (
        <View key={g.team.id} style={{ marginBottom: 10 }}>
          <View style={{ marginBottom: 8 }}>
            <TeamBadge teamId={g.team.id} size="sm" />
          </View>
          <View style={styles.wrap}>
            {g.players.map((p) => (
              <View key={p.id} style={{ marginBottom: 8 }}>
                {chip(p)}
              </View>
            ))}
          </View>
        </View>
      ))}
    </View>
  );
}

function Confirmation({ status, onRetry }: { status: Status; onRetry: () => void }) {
  const t = useTheme();
  if (status.kind === 'saving') {
    return (
      <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
        Saving your call…
      </Txt>
    );
  }
  if (status.kind === 'error') {
    return (
      <View>
        <Txt variant="bodySmall" tone="red">
          {status.message}
        </Txt>
        <Button label="Try again" kind="outline" size="sm" onPress={onRetry} style={{ alignSelf: 'flex-start', marginTop: 8 }} />
      </View>
    );
  }
  return (
    <View style={styles.confirm}>
      <View style={[styles.check, { backgroundColor: t.green }]}>
        <Feather name="check" size={14} color={t.onRed} />
      </View>
      <View style={{ marginLeft: 10, flex: 1 }}>
        <Txt variant="body" tone="green" style={{ fontFamily: fonts.uiBold }}>
          {status.kind === 'saved' ? 'Call saved' : 'Your call is in'}
        </Txt>
        <Txt variant="bodySmall" tone="soft">
          Tap a pick above to change it before the first ball.
        </Txt>
      </View>
    </View>
  );
}

/** From the first ball: the saved picks, read-only. Correct picks get a green check once there's a result. */
function LockedPicks({ match, picks, compact }: { match: Match; picks: Picks | null; compact: boolean }) {
  const t = useTheme();
  const r = match.state === 'completed' ? match.result : undefined;
  const correct = picks && r
    ? [r.winnerId === picks.winnerId, r.topRunScorerIds.includes(picks.topRunScorerId), r.topWicketTakerIds.includes(picks.topWicketTakerId)]
    : null;
  return (
    <View>
      <View style={styles.lockedHead}>
        <Txt variant="body" style={{ fontFamily: fonts.uiBold }}>
          {picks ? 'Your call' : 'No prediction made'}
        </Txt>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
          Locked 🔒
        </Txt>
      </View>
      {picks ? (
        FIELDS.map((f, i) => (
          <View key={f.key} style={[styles.lockedRow, { borderTopColor: t.rule, paddingVertical: compact ? 7 : 10 }]}>
            <Txt variant="bodySmall" tone="soft" style={{ width: 130 }}>
              {f.row}
            </Txt>
            <View style={{ flex: 1, flexDirection: 'row', alignItems: 'center', justifyContent: 'flex-end' }}>
              {i === 0 ? (
                <TeamBadge teamId={picks.winnerId} size="sm" />
              ) : (
                <Txt variant="bodySmall" numberOfLines={1} style={{ flexShrink: 1, fontFamily: fonts.uiBold }}>
                  {playerName(i === 1 ? picks.topRunScorerId : picks.topWicketTakerId)}
                </Txt>
              )}
              {correct ? (
                <View style={{ width: 22, alignItems: 'flex-end' }}>
                  {correct[i] ? <Feather name="check" size={15} color={t.green} /> : <Feather name="x" size={14} color={t.inkFaint} />}
                </View>
              ) : null}
            </View>
          </View>
        ))
      ) : (
        <Txt variant="review" tone="soft" style={{ marginTop: 6 }}>
          The first ball has been bowled — calls are closed for this one.
        </Txt>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  tabs: { flexDirection: 'row', borderBottomWidth: StyleSheet.hairlineWidth },
  tab: { flex: 1, minHeight: 48, paddingTop: 8, paddingBottom: 6, paddingRight: 8, borderBottomWidth: 2, marginBottom: -StyleSheet.hairlineWidth },
  teamRow: { flexDirection: 'row' },
  teamButton: { flex: 1, overflow: 'hidden' },
  chipStrip: { alignItems: 'center', paddingRight: 8 },
  wrap: { flexDirection: 'row', flexWrap: 'wrap' },
  confirm: { flexDirection: 'row', alignItems: 'center' },
  check: { width: 22, height: 22, alignItems: 'center', justifyContent: 'center' },
  lockedHead: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 },
  lockedRow: { flexDirection: 'row', alignItems: 'center', borderTopWidth: StyleSheet.hairlineWidth },
});
