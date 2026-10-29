import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useEffect, useState } from 'react';
import { Pressable, StyleSheet, View, type LayoutChangeEvent } from 'react-native';
import Animated, { Easing, useAnimatedStyle, useSharedValue, withRepeat, withSequence, withTiming } from 'react-native-reanimated';

import { Avatar, Txt } from '@/components';
import { ApiError, getTeam, useActions } from '@/data';
import type { CallView } from '@/domain/api';
import type { CallOption, Match, User } from '@/domain/types';
import { pct } from '@/lib/format';
import { fonts, useTheme } from '@/theme';

const EASE = Easing.bezier(0.16, 1, 0.3, 1);

/** Predictions during the match: every call on it, newest first. */
export function LiveCalls({ calls, match, users }: { calls: CallView[]; match?: Match; users?: Record<string, User> }) {
  return (
    <View style={{ gap: 14 }}>
      {calls.map((cv) => (
        <CallCard key={cv.call.id} view={cv} match={match} asker={cv.call.createdBy ? users?.[cv.call.createdBy] : undefined} />
      ))}
    </View>
  );
}

/** A range call's boundaries from its option ids ("lt150", "150", "170", "190plus" → 150, 170, 190), or null. */
function cutsOf(options: CallOption[]): number[] | null {
  if (options.length < 3) return null;
  const cuts = options.slice(1).map((o) => Number.parseInt(o.id.replace(/plus$/, ''), 10));
  return cuts.every((n) => Number.isFinite(n)) ? cuts : null;
}

const KIND: Record<string, string> = { total: 'Innings total', chase: 'The chase', milestone: 'Batter', runs: 'Batter', wickets: 'Bowler', teamTotal: 'Team total', win: 'Result' };

/**
 * One call on match-ticket stock. Yes/no calls are a ticket torn down the middle, each half filling with the share
 * of fans who called it; ranges are a stack of meters. Your pick wears the red stamp; the answer, once it's in,
 * a green rule. Tap an option while the call is open.
 */
export function CallCard({ view, compact = false, asker, match }: { view: CallView; compact?: boolean; asker?: User; match?: Match }) {
  const t = useTheme();
  const { pickCall } = useActions();
  const [picked, setPicked] = useState<string | null>(null); // optimistic until the page refetches
  const [error, setError] = useState<string | null>(null);
  const { call, points } = view;
  const mine = picked ?? view.mine;
  const open = call.state === 'open';

  // Counts with my optimistic move applied, so the meters answer the tap at once.
  const counts: Record<string, number> = { ...view.counts };
  if (picked && picked !== view.mine) {
    counts[picked] = (counts[picked] ?? 0) + 1;
    if (view.mine) counts[view.mine] = Math.max(0, (counts[view.mine] ?? 0) - 1);
  }
  const total = Object.values(counts).reduce((a, b) => a + b, 0);
  const share = (id: string) => (total > 0 ? (counts[id] ?? 0) / total : 0);

  const pick = async (optionId: string) => {
    if (!open || optionId === mine) return;
    setError(null);
    setPicked(optionId);
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light).catch(() => {});
    try {
      await pickCall(call.matchId, call.id, optionId);
    } catch (e) {
      setPicked(null);
      setError(e instanceof ApiError && e.status === 409 ? 'Too late, that one has locked.' : 'That didn’t go through. Try again.');
    }
  };

  const footer =
    call.state === 'open'
      ? `Locks ${call.locksAt}`
      : call.state === 'locked'
        ? 'Locked. Waiting on the result.'
        : call.state === 'settled'
          ? call.result
          : `No result: ${call.result}`;
  const fans = total === 0 ? 'Nobody has called it yet' : total === 1 ? '1 fan has called it' : `${total} fans have called it`;
  const yesNo = call.options.length === 2;
  const cuts = cutsOf(call.options);
  // Where the batting side stands now, for a total call's gauge.
  const first = (match?.live?.innings ?? match?.result?.innings ?? [])[0];
  const now = call.kind === 'total' && first ? { runs: first.runs, label: `${getTeam(first.teamId).short} ${first.runs}` } : undefined;

  return (
    <View style={[styles.card, { backgroundColor: t.card, borderColor: t.rule }, compact && { paddingTop: 12, paddingBottom: 12 }]}>
      <View style={styles.head}>
        <Txt variant="label" tone="soft">
          {asker ? 'Fan call' : (KIND[call.kind] ?? 'Call')}
        </Txt>
        <StatePill state={call.state} />
      </View>
      {asker ? (
        <View style={styles.asker}>
          <Avatar user={asker} size={18} />
          <Txt variant="bodySmall" tone="soft" style={{ marginLeft: 6 }}>
            Asked by {asker.name.split(' ')[0]}
          </Txt>
        </View>
      ) : null}
      <Txt style={[styles.question, { color: t.ink }, compact && { fontSize: 17, lineHeight: 22 }]}>{call.question}</Txt>

      {yesNo ? (
        <SplitTicket
          options={call.options}
          share={share}
          mine={mine}
          answer={call.state === 'settled' ? call.answerId : null}
          open={open}
          showShares={total > 0}
          compact={compact}
          onPick={pick}
        />
      ) : cuts ? (
        <Gauge
          options={call.options}
          cuts={cuts}
          share={share}
          mine={mine}
          answer={call.state === 'settled' ? call.answerId : null}
          open={open}
          showShares={total > 0}
          now={now}
          onPick={pick}
        />
      ) : (
        <View style={{ gap: 8, marginTop: 14 }}>
          {call.options.map((o) => (
            <Meter
              key={o.id}
              option={o}
              share={share(o.id)}
              showShare={total > 0}
              chosen={mine === o.id}
              right={call.state === 'settled' && call.answerId === o.id}
              open={open}
              onPick={() => pick(o.id)}
            />
          ))}
        </View>
      )}

      <View style={[styles.foot, { borderTopColor: t.rule }]}>
        <Txt variant="monoSmall" tone="soft" style={{ flexShrink: 1 }}>
          {fans}
        </Txt>
        {call.state === 'settled' && mine ? (
          <Txt variant="monoSmall" tone={points > 0 ? 'green' : 'soft'} style={{ marginLeft: 10 }}>
            {points > 0 ? `+${points} pts` : 'No points'}
          </Txt>
        ) : null}
      </View>
      <Txt variant="bodySmall" tone={call.state === 'settled' && points > 0 ? 'green' : 'soft'} style={{ marginTop: 4 }}>
        {footer}
      </Txt>
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 4 }}>
          {error}
        </Txt>
      ) : null}
    </View>
  );
}

/** "Open" with a beating red dot, "Locked" with a padlock, or how it ended. */
function StatePill({ state }: { state: CallView['call']['state'] }) {
  const t = useTheme();
  const beat = useSharedValue(1);
  useEffect(() => {
    if (state !== 'open') return;
    beat.set(withRepeat(withSequence(withTiming(0.35, { duration: 700 }), withTiming(1, { duration: 700 })), -1));
  }, [state, beat]);
  const dot = useAnimatedStyle(() => ({ opacity: beat.get() }));
  const tone = state === 'open' ? t.red : state === 'settled' ? t.green : t.inkSoft;
  return (
    <View style={[styles.pill, { borderColor: tone }]}>
      {state === 'open' ? <Animated.View style={[styles.dot, { backgroundColor: t.red }, dot]} /> : null}
      {state === 'locked' ? <Feather name="lock" size={10} color={tone} style={{ marginRight: 5 }} /> : null}
      {state === 'settled' ? <Feather name="check" size={11} color={tone} style={{ marginRight: 4 }} /> : null}
      <Txt variant="label" style={{ color: tone, fontSize: 9.5 }}>
        {state === 'open' ? 'Open' : state === 'locked' ? 'Locked' : state === 'settled' ? 'Settled' : 'Void'}
      </Txt>
    </View>
  );
}

/** A yes/no call as one ticket torn in two: each half fills from the bottom with the share of fans behind it. */
function SplitTicket({
  options,
  share,
  mine,
  answer,
  open,
  showShares,
  compact,
  onPick,
}: {
  options: CallOption[];
  share: (id: string) => number;
  mine: string | null;
  answer: string | null;
  open: boolean;
  showShares: boolean;
  compact: boolean;
  onPick: (id: string) => void;
}) {
  const t = useTheme();
  const height = compact ? 84 : 104;
  return (
    <View style={[styles.ticket, { height, borderColor: t.ink }]}>
      {options.map((o, k) => (
        <View key={o.id} style={{ flex: 1, flexDirection: 'row' }}>
          {k > 0 ? <Perforation height={height} /> : null}
          <Half
            option={o}
            share={share(o.id)}
            showShare={showShares}
            chosen={mine === o.id}
            right={answer === o.id}
            wrong={answer !== null && answer !== o.id}
            open={open}
            height={height}
            onPick={() => onPick(o.id)}
          />
        </View>
      ))}
    </View>
  );
}

function Half({
  option,
  share,
  showShare,
  chosen,
  right,
  wrong,
  open,
  height,
  onPick,
}: {
  option: CallOption;
  share: number;
  showShare: boolean;
  chosen: boolean;
  right: boolean;
  wrong: boolean;
  open: boolean;
  height: number;
  onPick: () => void;
}) {
  const t = useTheme();
  const fill = useSharedValue(0);
  useEffect(() => {
    fill.set(withTiming(showShare ? share : 0, { duration: 650, easing: EASE }));
  }, [share, showShare, fill]);
  const fillStyle = useAnimatedStyle(() => ({ height: fill.get() * height }));
  const fg = chosen ? t.onInk : t.ink;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ selected: chosen, disabled: !open }}
      accessibilityLabel={`${option.label}${showShare ? `, ${pct(share)} of fans` : ''}${right ? ', came true' : ''}${chosen ? ', your call' : ''}`}
      disabled={!open}
      onPress={onPick}
      style={({ pressed }) => [{ flex: 1, overflow: 'hidden', backgroundColor: chosen ? t.ink : 'transparent', transform: [{ scale: pressed ? 0.98 : 1 }] }]}>
      <Animated.View style={[styles.rise, { backgroundColor: chosen ? 'rgba(255,255,255,0.12)' : right ? t.greenSoft : t.redSoft }, fillStyle]} />
      <View style={styles.halfBody}>
        <Txt style={{ fontFamily: fonts.display, fontSize: height > 90 ? 34 : 28, lineHeight: height > 90 ? 36 : 30, color: fg, opacity: wrong ? 0.4 : 1 }} allowFontScaling={false}>
          {option.label.toUpperCase()}
        </Txt>
        <Txt style={{ fontFamily: fonts.monoMedium, fontSize: 12, color: chosen ? t.onInk : t.inkSoft, marginTop: 2, opacity: wrong ? 0.5 : 1 }}>
          {showShare ? pct(share) : open ? 'Tap to call' : '–'}
        </Txt>
      </View>
      {right ? <View style={[styles.rightRule, { backgroundColor: t.green }]} /> : null}
      {chosen ? (
        <View style={[styles.mineTag, { backgroundColor: t.red }]} pointerEvents="none">
          <Txt variant="label" style={{ color: t.onRed, fontSize: 8.5, lineHeight: 11 }}>
            Your call
          </Txt>
        </View>
      ) : null}
    </Pressable>
  );
}

/**
 * A range call as a scoreboard gauge: one strip cut into the ranges, each filling with its share of fans, over a
 * runs scale with the boundaries marked and a red pointer at the batting side's score as it climbs.
 */
function Gauge({
  options,
  cuts,
  share,
  mine,
  answer,
  open,
  showShares,
  now,
  onPick,
}: {
  options: CallOption[];
  cuts: number[];
  share: (id: string) => number;
  mine: string | null;
  answer: string | null;
  open: boolean;
  showShares: boolean;
  now?: { runs: number; label: string };
  onPick: (id: string) => void;
}) {
  const t = useTheme();
  const [width, setWidth] = useState(0);
  const tile = width / options.length;
  // The pointer: which range the score is in, and how far through it (the end ranges span 40 runs either side).
  const at = (() => {
    if (!now) return null;
    const edges = [cuts[0] - 40, ...cuts, cuts[cuts.length - 1] + 40];
    for (let k = 0; k < options.length; k++) {
      if (now.runs < edges[k + 1] || k === options.length - 1) {
        const frac = Math.min(1, Math.max(0, (now.runs - edges[k]) / (edges[k + 1] - edges[k])));
        return (k + frac) * tile;
      }
    }
    return null;
  })();
  const pointer = useSharedValue(0);
  useEffect(() => {
    if (at !== null) pointer.set(withTiming(at, { duration: 700, easing: EASE }));
  }, [at, pointer]);
  const pointerStyle = useAnimatedStyle(() => ({ transform: [{ translateX: pointer.get() - 6 }] }));
  return (
    <View style={{ marginTop: 14 }} onLayout={(e) => setWidth(e.nativeEvent.layout.width)}>
      <View style={[styles.strip, { borderColor: t.ink }]}>
        {options.map((o, k) => (
          <GaugeTile
            key={o.id}
            option={o}
            first={k === 0}
            share={share(o.id)}
            showShare={showShares}
            chosen={mine === o.id}
            right={answer === o.id}
            wrong={answer !== null && answer !== o.id}
            open={open}
            onPick={() => onPick(o.id)}
          />
        ))}
      </View>
      <View style={[styles.scale, { borderTopColor: t.ink }]}>
        {cuts.map((c, k) => (
          <View key={c} style={[styles.tick, { left: tile * (k + 1) - 0.75, backgroundColor: t.ink }]} />
        ))}
        {cuts.map((c, k) => (
          <Txt key={`l${c}`} variant="monoSmall" tone="soft" style={[styles.tickLabel, { left: tile * (k + 1) - 20 }]}>
            {c}
          </Txt>
        ))}
        {now && width > 0 ? (
          <Animated.View style={[styles.pointer, pointerStyle]}>
            <View style={[styles.triangle, { borderBottomColor: t.red }]} />
          </Animated.View>
        ) : null}
      </View>
      {now ? (
        <Txt variant="monoSmall" tone="red" style={{ marginTop: 16 }}>
          ▲ {now.label} now
        </Txt>
      ) : null}
    </View>
  );
}

function GaugeTile({
  option,
  first,
  share,
  showShare,
  chosen,
  right,
  wrong,
  open,
  onPick,
}: {
  option: CallOption;
  first: boolean;
  share: number;
  showShare: boolean;
  chosen: boolean;
  right: boolean;
  wrong: boolean;
  open: boolean;
  onPick: () => void;
}) {
  const t = useTheme();
  const fill = useSharedValue(0);
  useEffect(() => {
    fill.set(withTiming(showShare ? share : 0, { duration: 650, easing: EASE }));
  }, [share, showShare, fill]);
  const fillStyle = useAnimatedStyle(() => ({ height: `${fill.get() * 100}%` }));
  // "Under 150" / "150–169" / "190 or more" → a short scoreboard label: "<150", "150s", "190+".
  const short = option.label.startsWith('Under') ? `<${option.label.replace(/\D+/g, '')}` : option.label.includes('or more') ? `${option.label.replace(/\D+/g, '')}+` : option.label;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ selected: chosen, disabled: !open }}
      accessibilityLabel={`${option.label}${showShare ? `, ${pct(share)} of fans` : ''}${right ? ', came true' : ''}${chosen ? ', your call' : ''}`}
      disabled={!open}
      onPress={onPick}
      style={({ pressed }) => [
        styles.gaugeTile,
        { borderLeftWidth: first ? 0 : 1, borderLeftColor: t.ink, backgroundColor: chosen ? t.ink : 'transparent', transform: [{ scale: pressed ? 0.97 : 1 }] },
      ]}>
      <Animated.View style={[styles.rise, { backgroundColor: chosen ? 'rgba(255,255,255,0.12)' : right ? t.greenSoft : t.redSoft }, fillStyle]} />
      <Txt
        style={{ fontFamily: fonts.condensed, fontSize: 19, lineHeight: 21, color: chosen ? t.onInk : t.ink, opacity: wrong ? 0.4 : 1 }}
        allowFontScaling={false}
        numberOfLines={1}
        adjustsFontSizeToFit>
        {short}
      </Txt>
      <Txt style={{ fontFamily: fonts.monoMedium, fontSize: 11, color: chosen ? t.onInk : t.inkSoft, marginTop: 3, opacity: wrong ? 0.5 : 1 }}>
        {showShare ? pct(share) : '–'}
      </Txt>
      {right ? <View style={[styles.rightRule, { backgroundColor: t.green }]} /> : null}
      {chosen ? <View style={[styles.mineDot, { backgroundColor: t.red }]} /> : null}
    </Pressable>
  );
}

/** The tear line between two halves of a ticket: a dashed rule with a half-circle bitten out of each edge. */
function Perforation({ height }: { height: number }) {
  return (
    <View style={{ width: 14, height, alignItems: 'center' }}>
      <Bite edge="top" />
      <View style={{ flex: 1, marginVertical: 11, borderLeftWidth: 1.5, borderStyle: 'dashed', borderColor: 'rgba(22,23,26,0.32)' }} />
      <Bite edge="bottom" />
    </View>
  );
}

/** A half circle cut into the ticket's edge: only the inner half of the ring shows, over the border. */
function Bite({ edge }: { edge: 'top' | 'bottom' }) {
  const t = useTheme();
  return (
    <View style={[styles.biteClip, edge === 'top' ? { top: -1.5 } : { bottom: -1.5 }]} pointerEvents="none">
      <View style={[styles.biteRing, { backgroundColor: t.card, borderColor: t.ink }, edge === 'top' ? { top: -7.5 } : { bottom: -7.5 }]} />
    </View>
  );
}

/** One range of a multi-option call: a meter that fills to the share of fans who picked it. */
function Meter({
  option,
  share,
  showShare,
  chosen,
  right,
  open,
  onPick,
}: {
  option: CallOption;
  share: number;
  showShare: boolean;
  chosen: boolean;
  right: boolean;
  open: boolean;
  onPick: () => void;
}) {
  const t = useTheme();
  const [width, setWidth] = useState(0);
  const fill = useSharedValue(0);
  useEffect(() => {
    fill.set(withTiming(showShare ? share : 0, { duration: 650, easing: EASE }));
  }, [share, showShare, fill]);
  const fillStyle = useAnimatedStyle(() => ({ width: fill.get() * width }));
  const onLayout = (e: LayoutChangeEvent) => setWidth(e.nativeEvent.layout.width);
  const edge = right ? t.green : chosen ? t.ink : t.rule;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ selected: chosen, disabled: !open }}
      accessibilityLabel={`${option.label}${showShare ? `, ${pct(share)} of fans` : ''}${right ? ', came true' : ''}${chosen ? ', your call' : ''}`}
      disabled={!open}
      onPress={onPick}
      onLayout={onLayout}
      style={({ pressed }) => [styles.meter, { borderColor: edge, borderWidth: chosen || right ? 1.5 : 1, transform: [{ scale: pressed ? 0.985 : 1 }] }]}>
      <Animated.View style={[styles.meterFill, { backgroundColor: chosen ? t.ink : right ? t.greenSoft : t.redSoft, opacity: chosen ? 0.1 : 1 }, fillStyle]} />
      {chosen ? <View style={[styles.meterNotch, { backgroundColor: t.red }]} /> : null}
      <Txt style={{ fontFamily: fonts.condensed, fontSize: 20, lineHeight: 22, color: t.ink, flex: 1 }} allowFontScaling={false}>
        {option.label}
      </Txt>
      {right ? <Feather name="check" size={15} color={t.green} style={{ marginRight: 8 }} /> : null}
      {chosen ? (
        <Txt variant="label" tone="red" style={{ fontSize: 9, marginRight: 8 }}>
          Your call
        </Txt>
      ) : null}
      <Txt style={{ fontFamily: fonts.monoMedium, fontSize: 12.5, color: t.inkSoft, minWidth: 38, textAlign: 'right' }}>{showShare ? pct(share) : ''}</Txt>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: { borderWidth: StyleSheet.hairlineWidth, paddingHorizontal: 16, paddingTop: 14, paddingBottom: 14 },
  head: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  asker: { flexDirection: 'row', alignItems: 'center', marginTop: 8 },
  question: { fontFamily: fonts.heading, fontSize: 20, lineHeight: 25, letterSpacing: -0.3, marginTop: 6 },
  pill: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, paddingHorizontal: 7, paddingVertical: 3 },
  dot: { width: 6, height: 6, borderRadius: 3, marginRight: 5 },
  ticket: { flexDirection: 'row', borderWidth: 1.5, marginTop: 14 },
  rise: { position: 'absolute', left: 0, right: 0, bottom: 0 },
  halfBody: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  rightRule: { position: 'absolute', left: 0, right: 0, bottom: 0, height: 4 },
  mineTag: { position: 'absolute', top: 0, left: 0, paddingHorizontal: 6, paddingVertical: 3 },
  strip: { flexDirection: 'row', borderWidth: 1.5, height: 64 },
  gaugeTile: { flex: 1, alignItems: 'center', justifyContent: 'center', overflow: 'hidden', paddingHorizontal: 4 },
  mineDot: { position: 'absolute', top: 6, right: 6, width: 7, height: 7, borderRadius: 3.5 },
  scale: { height: 16, borderTopWidth: 1.5, marginTop: 6 },
  tick: { position: 'absolute', top: -1.5, width: 1.5, height: 8 },
  tickLabel: { position: 'absolute', top: 7, width: 40, textAlign: 'center', fontSize: 10.5 },
  pointer: { position: 'absolute', top: 1, left: 0, width: 12, alignItems: 'center' },
  triangle: { width: 0, height: 0, borderLeftWidth: 6, borderRightWidth: 6, borderBottomWidth: 9, borderLeftColor: 'transparent', borderRightColor: 'transparent' },
  biteClip: { position: 'absolute', left: 0, width: 14, height: 9, overflow: 'hidden', zIndex: 2 },
  biteRing: { position: 'absolute', left: 0, width: 14, height: 14, borderRadius: 7, borderWidth: 1.5 },
  meter: { flexDirection: 'row', alignItems: 'center', minHeight: 48, paddingHorizontal: 14, overflow: 'hidden' },
  meterFill: { position: 'absolute', left: 0, top: 0, bottom: 0 },
  meterNotch: { position: 'absolute', left: 0, top: 0, bottom: 0, width: 4 },
  foot: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', borderTopWidth: StyleSheet.hairlineWidth, marginTop: 14, paddingTop: 10 },
});
