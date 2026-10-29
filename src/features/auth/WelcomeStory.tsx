import { useFocusEffect } from 'expo-router';
import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { AccessibilityInfo, Pressable, StyleSheet, useWindowDimensions, View, type StyleProp, type ViewStyle } from 'react-native';
import Animated, {
  Easing,
  Extrapolation,
  interpolate,
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withDelay,
  withSequence,
  withSpring,
  withTiming,
} from 'react-native-reanimated';

import { Avatar, BallRating, Barcode, Rise, Stamp, TicketStub, Txt, type Tone } from '@/components';
import { useTheme } from '@/theme';

const H = 204; // every scene is one ticket of this height
const HEADROOM = 24; // room above it for the Vault's fanned tickets
const SCENE_MS = 3400;
// Ticket text stays close to its designed size, and a line that still won't fit shrinks rather than ending in "…":
// the ticket is a fixed-height graphic, and big system text would push its lines off the card.
const FIXED = { maxFontSizeMultiplier: 1.15, adjustsFontSizeToFit: true, minimumFontScale: 0.75 } as const;
// The captions grow with the system text up to this much, and their space grows with them, so the screen still
// fits without scrolling on a small phone.
const CAPTION_SCALE = 1.25;
const CAPTION_HEIGHT = 92; // one line of headline and up to two of text, at 1x

/**
 * The app in five tickets, in the order a matchday goes. Each has a short headline and one line under it; the
 * tag is only read out to screen readers.
 */
const SCENES = [
  { tag: 'Toss', title: 'Call it.', line: 'Pick the winner before the first ball.', Ticket: CallTicket },
  { tag: 'Stumps', title: 'Rate it.', line: 'Out of five cricket balls, at stumps.', Ticket: RateTicket },
  { tag: 'Stub', title: 'Keep the stub.', line: 'Where you watched it, and who with.', Ticket: StubTicket },
  { tag: 'Huddle', title: 'Bring your people.', line: 'Right calls climb the Bragging Board.', Ticket: BoardTicket },
  { tag: 'Vault', title: 'Relive the classics.', line: '58 real matches everyone remembers.', Ticket: VaultTicket },
];

/**
 * The Welcome screen's story: one ticket that plays each part of the app in turn, with story-style progress bars
 * and a headline under it. Each ticket has one moment (a stamp lands, the balls fill, you climb the board). It moves
 * on by itself every few seconds while the screen is in front (not with a screen reader on); a tap on the ticket or
 * a bar moves it on. `lead` holds the first scene for the launch splash.
 */
export function WelcomeStory({ lead = 0, style }: { lead?: number; style?: ViewStyle }) {
  const t = useTheme();
  const { fontScale } = useWindowDimensions();
  // `n` counts every change of scene, so each one mounts fresh and plays from the start. `prev` is the scene on its
  // way out: it stays mounted under the same key while it slides away, so the stage is never empty in between.
  const [step, setStep] = useState<{ i: number; n: number; prev: number | null }>({ i: 0, n: 0, prev: null });
  const [reader, setReader] = useState(false);
  const holdFor = (n: number) => (n === 0 ? lead : 0);
  const go = useCallback(
    (to: number) => setStep((s) => ({ i: (to + SCENES.length) % SCENES.length, n: s.n + 1, prev: s.i })),
    [],
  );

  useEffect(() => {
    AccessibilityInfo.isScreenReaderEnabled().then(setReader, () => {});
    const sub = AccessibilityInfo.addEventListener('screenReaderChanged', setReader);
    return () => sub.remove();
  }, []);

  const hold = holdFor(step.n);
  useFocusEffect(
    useCallback(() => {
      if (reader) return;
      const id = setTimeout(() => go(step.i + 1), hold + SCENE_MS);
      return () => clearTimeout(id);
    }, [go, reader, step.i, hold]),
  );

  const scene = SCENES[step.i];
  const slots =
    step.prev === null
      ? [{ i: step.i, n: step.n, leaving: false }]
      : [
          { i: step.prev, n: step.n - 1, leaving: true },
          { i: step.i, n: step.n, leaving: false },
        ];
  return (
    <View style={style}>
      <Pressable
        onPress={() => go(step.i + 1)}
        accessibilityRole="button"
        accessibilityLabel={`${scene.title} ${scene.line}`}
        accessibilityHint="Shows the next part of Innings"
        style={styles.stage}>
        {slots.map((s) => {
          const Ticket = SCENES[s.i].Ticket;
          return (
            <Slot key={s.n} first={s.n === 0} hold={holdFor(s.n)} leaving={s.leaving} style={styles.ticketSlot}>
              <Ticket start={holdFor(s.n) + 420} />
            </Slot>
          );
        })}
      </Pressable>

      <View style={styles.progress}>
        {SCENES.map((s, k) => (
          <Pressable
            key={s.tag}
            onPress={() => go(k)}
            hitSlop={{ top: 14, bottom: 14 }}
            accessibilityRole="button"
            accessibilityLabel={`${s.tag}: ${s.title}`}
            style={styles.segHit}>
            <View style={[styles.seg, { backgroundColor: t.rule }]}>
              {k < step.i ? <Full /> : k === step.i ? <Fill key={step.n} hold={hold} paused={reader} /> : null}
            </View>
          </Pressable>
        ))}
      </View>

      <View style={[styles.caption, { height: Math.ceil(CAPTION_HEIGHT * Math.min(Math.max(fontScale, 1), CAPTION_SCALE)) }]}>
        {slots.map((s) => (
          <Slot key={s.n} first={false} hold={holdFor(s.n) + 60} leaving={s.leaving} leaveMs={180} distance={0} style={StyleSheet.absoluteFill}>
            <Caption scene={SCENES[s.i]} delay={holdFor(s.n) + 200} />
          </Slot>
        ))}
      </View>
    </View>
  );
}

/** A scene's caption: the headline wipes up from behind its own baseline, then the line fades up under it. */
function Caption({ scene, delay }: { scene: (typeof SCENES)[number]; delay: number }) {
  const p = useSharedValue(0);
  useEffect(() => {
    p.set(withDelay(delay, withTiming(1, { duration: 560, easing: Easing.bezier(0.16, 1, 0.3, 1) })));
  }, [delay, p]);
  const wipe = useAnimatedStyle(() => ({
    opacity: interpolate(p.value, [0, 0.35], [0, 1], Extrapolation.CLAMP),
    transform: [{ translateY: (1 - p.value) * 40 }],
  }));
  return (
    <>
      <View style={styles.clip}>
        <Animated.View style={wipe}>
          <Txt variant="h1" numberOfLines={1} adjustsFontSizeToFit maxFontSizeMultiplier={CAPTION_SCALE} style={styles.headline}>
            {scene.title}
          </Txt>
        </Animated.View>
      </View>
      <Rise delay={delay + 180}>
        <Txt variant="body" tone="soft" numberOfLines={2} adjustsFontSizeToFit maxFontSizeMultiplier={CAPTION_SCALE} style={{ marginTop: 6 }}>
          {scene.line}
        </Txt>
      </Rise>
    </>
  );
}

/**
 * One scene on the stage. It rises into place on a soft spring (from a little lower the first time) while the one
 * before it, once `leaving`, eases up and fades out, so the two overlap and the stage is never empty. No sideways
 * slide and no tilt. `distance` is how far it travels; the caption block uses 0 and lets its lines rise instead.
 */
function Slot({
  first,
  hold,
  leaving,
  leaveMs = 300,
  distance = 18,
  style,
  children,
}: {
  first: boolean;
  hold: number;
  leaving: boolean;
  leaveMs?: number;
  distance?: number;
  style?: ViewStyle;
  children: ReactNode;
}) {
  const arrive = useSharedValue(0);
  const leave = useSharedValue(0);
  useEffect(() => {
    arrive.set(withDelay(hold, withSpring(1, { damping: 20, stiffness: first ? 120 : 150, mass: 1 })));
  }, [first, hold, arrive]);
  useEffect(() => {
    if (leaving) leave.set(withTiming(1, { duration: leaveMs, easing: Easing.out(Easing.quad) }));
  }, [leaving, leaveMs, leave]);
  const rise = first ? distance * 1.7 : distance;
  const anim = useAnimatedStyle(() => ({
    opacity: interpolate(arrive.value, [0, 0.55], [0, 1], Extrapolation.CLAMP) * (1 - leave.value),
    transform: [
      { translateY: (1 - arrive.value) * rise - leave.value * distance * 0.6 },
      { scale: 0.97 + 0.03 * Math.min(1, arrive.value) - 0.03 * leave.value },
    ],
  }));
  return (
    <Animated.View pointerEvents={leaving ? 'none' : 'auto'} style={[style, anim]}>
      {children}
    </Animated.View>
  );
}

function Full() {
  const t = useTheme();
  return <View style={[StyleSheet.absoluteFill, { backgroundColor: t.ink }]} />;
}

/** The current scene's bar, filling over the scene's time. */
function Fill({ hold, paused }: { hold: number; paused: boolean }) {
  const t = useTheme();
  const p = useSharedValue(paused ? 1 : 0);
  useEffect(() => {
    if (!paused) p.value = withDelay(hold, withTiming(1, { duration: SCENE_MS, easing: Easing.linear }));
  }, [hold, paused, p]);
  const anim = useAnimatedStyle(() => ({ width: `${p.value * 100}%` }));
  return <Animated.View style={[styles.fill, { backgroundColor: t.ink }, anim]} />;
}

/** Shows its child from `at` ms on; before that it isn't there at all, so a stamp never flashes early. */
function Late({ at, children, style }: { at: number; children: ReactNode; style?: StyleProp<ViewStyle> }) {
  const shown = useSharedValue(0);
  useEffect(() => {
    shown.value = withDelay(at, withTiming(1, { duration: 0 }));
  }, [at, shown]);
  const anim = useAnimatedStyle(() => ({ opacity: shown.value }));
  return <Animated.View style={[style, anim]}>{children}</Animated.View>;
}

/** Counts up to `to` in `step`s from `at` ms on: the balls filling in. With Reduce Motion on, it starts full. */
function useCountUp(to: number, step: number, at: number, every: number) {
  const reduce = useReducedMotion();
  const [v, setV] = useState(reduce ? to : 0);
  useEffect(() => {
    if (reduce) return;
    let tick: ReturnType<typeof setInterval> | undefined;
    let shown = 0;
    const begin = setTimeout(() => {
      tick = setInterval(() => {
        shown = Math.min(to, shown + step);
        setV(shown);
        if (shown >= to) clearInterval(tick);
      }, every);
    }, at);
    return () => {
      clearTimeout(begin);
      clearInterval(tick);
    };
  }, [reduce, to, step, at, every]);
  return v;
}

/** Gives the ticket a small press at `at` ms, the moment a stamp lands on it. */
function Thump({ at, children }: { at: number; children: ReactNode }) {
  const s = useSharedValue(1);
  useEffect(() => {
    s.set(withDelay(at, withSequence(withTiming(0.975, { duration: 70, easing: Easing.out(Easing.quad) }), withSpring(1, { damping: 9, stiffness: 300 }))));
  }, [at, s]);
  const anim = useAnimatedStyle(() => ({ transform: [{ scale: s.value }] }));
  return <Animated.View style={anim}>{children}</Animated.View>;
}

// ---------- the five tickets ----------

// A stamp takes this long from its delay to hit the card (see Stamp in components/Motion).
const STAMP_LANDS = 170;

function CallTicket({ start }: { start: number }) {
  const t = useTheme();
  const stamp = start + 700;
  return (
    <Thump at={stamp + STAMP_LANDS}>
      <TicketStub
        stock={t.stock.ipl}
        height={H}
        overlay={
          <Late at={stamp} style={styles.stampAt}>
            <Stamp visible delay={stamp} text="LOCKED" size="sm" />
          </Late>
        }
        main={
          <View style={styles.main}>
            <Txt {...FIXED} variant="label" numberOfLines={1}>
              IPL · Match 21 · Your call
            </Txt>
            <Teams a="MI" b="CSK" />
            <Pick at={start} label="Winner" value="Mumbai Indians" />
          </View>
        }
        side={<Side top="Today" seed="welcome-call" />}
      />
    </Thump>
  );
}

function RateTicket({ start }: { start: number }) {
  const t = useTheme();
  const rating = useCountUp(4.5, 0.5, start, 150);
  return (
    <TicketStub
      stock={t.stock.international}
      height={H}
      main={
        <View style={styles.main}>
          <Txt {...FIXED} variant="label" numberOfLines={1}>
            Intl · 3rd T20I · Your rating
          </Txt>
          <Teams a="IND" b="AUS" />
          <View style={styles.ratingRow}>
            <BallRating value={rating} size={22} gap={4} />
            <Txt {...FIXED} variant="score" style={{ marginLeft: 10 }}>
              {rating.toFixed(1)}
            </Txt>
          </View>
        </View>
      }
      side={<Side top="Full time" seed="welcome-rate" />}
    />
  );
}

function StubTicket({ start }: { start: number }) {
  const t = useTheme();
  const stamp = start + 200;
  return (
    <Thump at={stamp + STAMP_LANDS}>
      <TicketStub
        stock={t.stock.test}
        height={H}
        overlay={
          <Late at={stamp} style={styles.stampAt}>
            <Stamp visible delay={stamp} size="sm" />
          </Late>
        }
        main={
          <View style={styles.main}>
            <Txt {...FIXED} variant="label" numberOfLines={1}>
              Admit one · 2nd Test · Day 5
            </Txt>
            <Teams a="IND" b="ENG" />
            <Txt {...FIXED} variant="bodySmall" numberOfLines={1}>
              📍 Home, with Aditi and Kabir
            </Txt>
          </View>
        }
        side={<Side top="No. 0007" tone="red" seed="welcome-stub" />}
      />
    </Thump>
  );
}

const ROW = 42; // one place on the board
const BOARD = [
  { name: 'Aditi Rao', colour: '#1C2333', points: 85, gain: 0 },
  { name: 'You', colour: '#3B5873', points: 70, gain: 20 },
  { name: 'Kabir Mehta', colour: '#5A3E36', points: 40, gain: 0 },
];

/** A right call comes in: your points tick up, your bar grows, and you climb past Aditi into first. */
function BoardTicket({ start }: { start: number }) {
  const gain = useCountUp(20, 5, start + 600, 90);
  const rows = BOARD.map((row) => ({ ...row, points: row.points + (row.gain ? gain : 0) }));
  // Ties keep their order, so you only move up once you're clear of Aditi.
  const order = [...rows].sort((a, b) => b.points - a.points);
  return (
    <TicketStub
      height={H}
      main={
        <View style={styles.main}>
          <Txt {...FIXED} variant="label" numberOfLines={1}>
            Bragging Board · This week
          </Txt>
          <View style={{ height: ROW * BOARD.length }}>
            {rows.map((row) => (
              <BoardRow key={row.name} row={row} rank={order.indexOf(row)} />
            ))}
          </View>
        </View>
      }
      side={<Side top="Week" seed="welcome-board" />}
    />
  );
}

/** One place on the board. It springs to its new place when its rank changes, the climber on top of the rest. */
function BoardRow({ row, rank }: { row: (typeof BOARD)[number]; rank: number }) {
  const t = useTheme();
  const y = useSharedValue(rank * ROW);
  const share = useSharedValue(row.points / 100);
  useEffect(() => {
    y.set(withSpring(rank * ROW, { damping: 17, stiffness: 170 }));
  }, [rank, y]);
  useEffect(() => {
    share.set(withTiming(row.points / 100, { duration: 160 }));
  }, [row.points, share]);
  const place = useAnimatedStyle(() => ({ transform: [{ translateY: y.value }] }));
  const fill = useAnimatedStyle(() => ({ width: `${share.value * 100}%` }));
  return (
    <Animated.View style={[styles.boardRow, { backgroundColor: t.card, zIndex: row.gain ? 1 : 0 }, place]}>
      <Avatar user={{ name: row.name, avatarColour: row.colour }} size={28} />
      <Txt {...FIXED} variant="h3" numberOfLines={1} adjustsFontSizeToFit style={styles.boardName}>
        {row.name}
      </Txt>
      <View style={[styles.track, { backgroundColor: t.rule }]}>
        <Animated.View style={[styles.fill, { backgroundColor: t.ink }, fill]} />
      </View>
      <Txt {...FIXED} variant="mono" style={styles.points}>
        {row.points}
      </Txt>
    </Animated.View>
  );
}

/** A real classic from the Vault (see docs/VAULT_SOURCES.md), with more of the collection stacked behind it. */
function VaultTicket({ start }: { start: number }) {
  const t = useTheme();
  return (
    <View style={{ height: H }}>
      <View style={StyleSheet.absoluteFill} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
        <Stacked at={start + 120} lift={24} stock={t.stock['the-hundred']} />
        <Stacked at={start} lift={12} stock={t.stock['world-cup']} />
      </View>
      <TicketStub
        stock={t.stock.classic}
        height={H}
        main={
          <View style={styles.main}>
            <Txt {...FIXED} variant="label" numberOfLines={1}>
              The Vault · World Cup Final
            </Txt>
            <Teams a="SL" b="IND" />
            <View>
              <Txt {...FIXED} variant="bodySmall" tone="soft" numberOfLines={1}>
                India won by 6 wickets
              </Txt>
              <Txt {...FIXED} variant="monoSmall" numberOfLines={1} style={{ marginTop: 2 }}>
                02 APR 2011 · Wankhede, Mumbai
              </Txt>
            </View>
          </View>
        }
        side={<Side top="Classic" seed="wc-2011-final" />}
      />
    </View>
  );
}

/**
 * A plain ticket from the collection, rising from behind the classic until its top edge shows. The further back
 * it is (`lift`), the smaller it is. Straight, never tilted.
 */
function Stacked({ at, lift, stock }: { at: number; lift: number; stock: string }) {
  const p = useSharedValue(0);
  useEffect(() => {
    p.set(withDelay(at, withTiming(1, { duration: 480, easing: Easing.out(Easing.back(1.4)) })));
  }, [at, p]);
  const scale = 1 - lift / 240;
  const anim = useAnimatedStyle(() => ({ transform: [{ translateY: -p.value * lift }, { scale }] }));
  return (
    <Animated.View style={[StyleSheet.absoluteFill, anim]}>
      <TicketStub stock={stock} height={H} main={null} />
    </Animated.View>
  );
}

// ---------- ticket parts ----------

function Teams({ a, b }: { a: string; b: string }) {
  return (
    <View style={styles.teams}>
      <Txt {...FIXED} variant="display" numberOfLines={1} style={styles.code}>
        {a}
      </Txt>
      <Txt {...FIXED} variant="h3" style={{ marginHorizontal: 8 }}>
        vs
      </Txt>
      <Txt {...FIXED} variant="display" numberOfLines={1} style={styles.code}>
        {b}
      </Txt>
    </View>
  );
}

function Pick({ at, label, value }: { at: number; label: string; value: string }) {
  return (
    <Rise delay={at}>
      <View style={styles.pick}>
        <Txt {...FIXED} variant="label" tone="soft" numberOfLines={1} adjustsFontSizeToFit style={styles.pickLabel}>
          {label}
        </Txt>
        <Txt {...FIXED} variant="mono" numberOfLines={1} style={{ flexShrink: 1 }}>
          {value}
        </Txt>
      </View>
    </Rise>
  );
}

function Side({ top, tone = 'soft', seed }: { top: string; tone?: Tone; seed: string }) {
  return (
    <View style={styles.side}>
      <Txt {...FIXED} variant="label" tone={tone} numberOfLines={1} adjustsFontSizeToFit style={styles.serial}>
        {top}
      </Txt>
      <Barcode seed={seed} width={46} height={70} />
    </View>
  );
}

const styles = StyleSheet.create({
  stage: { height: H + HEADROOM },
  ticketSlot: { position: 'absolute', left: 0, right: 0, bottom: 0 },
  progress: { flexDirection: 'row', gap: 6, marginTop: 14 },
  caption: { marginTop: 14 },
  clip: { overflow: 'hidden' },
  headline: { fontSize: 32, letterSpacing: -0.8 },
  segHit: { flex: 1, paddingVertical: 6 },
  seg: { height: 3, overflow: 'hidden' },
  fill: { position: 'absolute', left: 0, top: 0, bottom: 0 },
  main: { flex: 1, justifyContent: 'space-between' },
  code: { fontSize: 46, lineHeight: 46 },
  teams: { flexDirection: 'row', alignItems: 'baseline' },
  pick: { flexDirection: 'row', alignItems: 'baseline' },
  pickLabel: { width: 92 },
  ratingRow: { flexDirection: 'row', alignItems: 'center' },
  boardRow: { position: 'absolute', left: 0, right: 0, top: 0, height: ROW, flexDirection: 'row', alignItems: 'center' },
  boardName: { width: 92, marginLeft: 8 },
  points: { width: 26, textAlign: 'right', marginLeft: 8 },
  track: { flex: 1, height: 6, marginLeft: 6, overflow: 'hidden' },
  side: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  serial: { fontSize: 9.5, letterSpacing: 0.4 },
  // Stamps land across the perforation beside the team codes, on blank card clear of every line of text.
  stampAt: { position: 'absolute', top: 54, right: 30 },
});
