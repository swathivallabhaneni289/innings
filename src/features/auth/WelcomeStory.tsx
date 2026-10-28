import { useFocusEffect } from 'expo-router';
import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { AccessibilityInfo, Pressable, StyleSheet, useWindowDimensions, View, type StyleProp, type ViewStyle } from 'react-native';
import Animated, {
  Easing,
  interpolate,
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withDelay,
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
const CAPTION_HEIGHT = 150; // a tag, two lines of title and three of text, at 1x

/** The app in five tickets, in the order a matchday goes. */
const SCENES = [
  {
    tag: 'Toss',
    title: 'Call it before the first ball.',
    line: 'Pick the winner and the top scorers. Right calls earn bragging points.',
    Ticket: CallTicket,
  },
  {
    tag: 'Stumps',
    title: 'Rate it after the last.',
    line: 'Half-balls out of five for the match, marks out of ten for the players.',
    Ticket: RateTicket,
  },
  {
    tag: 'Stub',
    title: 'Keep the stub.',
    line: 'Where you watched it and who with, printed for your collection.',
    Ticket: StubTicket,
  },
  {
    tag: 'Huddle',
    title: 'Bring your people.',
    line: 'Find friends from your contacts and race them on the Bragging Board.',
    Ticket: BoardTicket,
  },
  {
    tag: 'Vault',
    title: 'Relive the classics.',
    line: '58 real matches everyone remembers, from World Cup finals to Super Overs.',
    Ticket: VaultTicket,
  },
];

/**
 * The Welcome screen's story: one ticket that plays each part of the app in turn, with a line under it and
 * story-style progress bars. It moves on by itself every few seconds while the screen is in front (not with a
 * screen reader on); a tap on the ticket or a bar moves it on. `lead` holds the first scene for the launch splash.
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

      <View style={[styles.caption, { height: Math.ceil(CAPTION_HEIGHT * Math.min(Math.max(fontScale, 1), CAPTION_SCALE)) }]}>
        {slots.map((s) => (
          <Slot key={s.n} first={false} hold={holdFor(s.n) + 160} leaving={s.leaving} leaveMs={150} style={StyleSheet.absoluteFill}>
            <Caption scene={SCENES[s.i]} />
          </Slot>
        ))}
      </View>

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
    </View>
  );
}

function Caption({ scene }: { scene: (typeof SCENES)[number] }) {
  return (
    <>
      <Txt variant="label" tone="soft" maxFontSizeMultiplier={CAPTION_SCALE}>
        {scene.tag}
      </Txt>
      <Txt variant="h1" numberOfLines={2} adjustsFontSizeToFit maxFontSizeMultiplier={CAPTION_SCALE} style={{ marginTop: 4 }}>
        {scene.title}
      </Txt>
      <Txt variant="body" tone="soft" numberOfLines={3} adjustsFontSizeToFit maxFontSizeMultiplier={CAPTION_SCALE} style={{ marginTop: 4 }}>
        {scene.line}
      </Txt>
    </>
  );
}

/**
 * One scene on the stage. It arrives (dealt up onto the table the first time, slid in from the right after that)
 * and, once `leaving`, slides off to the left while the next one comes in.
 */
function Slot({
  first,
  hold,
  leaving,
  leaveMs = 320,
  style,
  children,
}: {
  first: boolean;
  hold: number;
  leaving: boolean;
  leaveMs?: number;
  style?: ViewStyle;
  children: ReactNode;
}) {
  const arrive = useSharedValue(0);
  const leave = useSharedValue(0);
  useEffect(() => {
    arrive.value = withDelay(
      hold,
      withTiming(1, { duration: first ? 560 : 440, easing: first ? Easing.out(Easing.back(1.2)) : Easing.out(Easing.cubic) }),
    );
  }, [first, hold, arrive]);
  useEffect(() => {
    if (leaving) leave.value = withTiming(1, { duration: leaveMs, easing: Easing.in(Easing.cubic) });
  }, [leaving, leaveMs, leave]);
  const anim = useAnimatedStyle(() => {
    const rest = 1 - arrive.value;
    return {
      opacity: interpolate(arrive.value, [0, 0.35, 1], [0, 1, 1]) * (1 - leave.value),
      transform: [
        { translateX: (first ? 0 : rest * 24) - leave.value * 24 },
        { translateY: first ? rest * 44 : 0 },
        { rotate: `${rest * (first ? 3 : 2) - leave.value * 2}deg` },
      ],
    };
  });
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

/** Grows from empty to `share` of its track, from `at` ms on. */
function Bar({ share, at }: { share: number; at: number }) {
  const t = useTheme();
  const p = useSharedValue(0);
  useEffect(() => {
    p.value = withDelay(at, withTiming(share, { duration: 700, easing: Easing.out(Easing.cubic) }));
  }, [at, share, p]);
  const anim = useAnimatedStyle(() => ({ width: `${p.value * 100}%` }));
  return (
    <View style={[styles.track, { backgroundColor: t.rule }]}>
      <Animated.View style={[styles.fill, { backgroundColor: t.ink }, anim]} />
    </View>
  );
}

// ---------- the five tickets ----------

function CallTicket({ start }: { start: number }) {
  const t = useTheme();
  return (
    <TicketStub
      stock={t.stock.ipl}
      height={H}
      overlay={
        <Late at={start + 1300} style={styles.stampAt}>
          <Stamp visible delay={start + 1300} text="LOCKED" size="sm" />
        </Late>
      }
      main={
        <View style={styles.main}>
          <Txt {...FIXED} variant="label" numberOfLines={1}>
            IPL · Match 21 · Your call
          </Txt>
          <Teams a="MI" b="CSK" />
          <View style={{ gap: 4 }}>
            <Pick at={start} label="Winner" value="Mumbai Indians" />
            <Pick at={start + 260} label="Most runs" value="Rohit Sharma" />
            <Pick at={start + 520} label="Most wickets" value="Jasprit Bumrah" />
          </View>
        </View>
      }
      side={<Side top="Today" seed="welcome-call" />}
    />
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
          <View>
            <View style={styles.ratingRow}>
              <BallRating value={rating} size={22} gap={4} />
              <Txt {...FIXED} variant="score" style={{ marginLeft: 10 }}>
                {rating.toFixed(1)}
              </Txt>
            </View>
            <Rise delay={start + 1500}>
              <Txt {...FIXED} variant="monoSmall" numberOfLines={1} adjustsFontSizeToFit minimumFontScale={0.8} style={{ marginTop: 6 }}>
                Player of the Match: Abhishek, 9/10
              </Txt>
            </Rise>
          </View>
        </View>
      }
      side={<Side top="Full time" seed="welcome-rate" />}
    />
  );
}

function StubTicket({ start }: { start: number }) {
  const t = useTheme();
  return (
    <TicketStub
      stock={t.stock.test}
      height={H}
      overlay={
        <Late at={start + 200} style={styles.stampAt}>
          <Stamp visible delay={start + 200} size="sm" />
        </Late>
      }
      main={
        <View style={styles.main}>
          <Txt {...FIXED} variant="label" numberOfLines={1}>
            Admit one · 2nd Test · Day 5
          </Txt>
          <Teams a="IND" b="ENG" />
          <View>
            <Txt {...FIXED} variant="bodySmall" numberOfLines={1}>
              📍 Home
            </Txt>
            <Txt {...FIXED} variant="monoSmall" tone="soft" numberOfLines={1} style={{ marginTop: 2 }}>
              With Aditi and Kabir
            </Txt>
          </View>
        </View>
      }
      side={<Side top="No. 0007" tone="red" seed="welcome-stub" />}
    />
  );
}

const BOARD = [
  { name: 'Aditi Rao', colour: '#1C2333', points: 85 },
  { name: 'You', colour: '#3B5873', points: 70 },
  { name: 'Kabir Mehta', colour: '#5A3E36', points: 40 },
];

function BoardTicket({ start }: { start: number }) {
  return (
    <TicketStub
      height={H}
      main={
        <View style={styles.main}>
          <Txt {...FIXED} variant="label" numberOfLines={1}>
            Bragging Board · This week
          </Txt>
          <View style={{ gap: 10 }}>
            {BOARD.map((row, k) => (
              <Rise key={row.name} delay={start + k * 180}>
                <View style={styles.boardRow}>
                  <Avatar user={{ name: row.name, avatarColour: row.colour }} size={26} />
                  <Txt {...FIXED} variant="h3" numberOfLines={1} adjustsFontSizeToFit style={styles.boardName}>
                    {row.name}
                  </Txt>
                  <Bar share={row.points / 100} at={start + 200 + k * 180} />
                  <Txt {...FIXED} variant="mono" style={styles.points}>
                    {row.points}
                  </Txt>
                </View>
              </Rise>
            ))}
          </View>
        </View>
      }
      side={<Side top="Week" seed="welcome-board" />}
    />
  );
}

/** A real classic from the Vault (see docs/VAULT_SOURCES.md), with more of the collection fanning out behind it. */
function VaultTicket({ start }: { start: number }) {
  const t = useTheme();
  return (
    <View style={{ height: H }}>
      <View style={StyleSheet.absoluteFill} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
        <Fanned at={start} rotate={-5} x={-10} stock={t.stock['world-cup']} />
        <Fanned at={start + 120} rotate={4.5} x={10} stock={t.stock['the-hundred']} />
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

/** A plain ticket from the collection, sliding out from behind the classic. */
function Fanned({ at, rotate, x, stock }: { at: number; rotate: number; x: number; stock: string }) {
  const p = useSharedValue(0);
  useEffect(() => {
    p.value = withDelay(at, withTiming(1, { duration: 520, easing: Easing.out(Easing.back(1.4)) }));
  }, [at, p]);
  const anim = useAnimatedStyle(() => ({
    transform: [{ translateX: p.value * x }, { translateY: p.value * -10 }, { rotate: `${p.value * rotate}deg` }, { scale: 0.94 }],
  }));
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
  caption: { marginTop: 22 },
  progress: { flexDirection: 'row', gap: 6, marginTop: 8 },
  segHit: { flex: 1, paddingVertical: 6 },
  seg: { height: 3, overflow: 'hidden' },
  fill: { position: 'absolute', left: 0, top: 0, bottom: 0 },
  main: { flex: 1, justifyContent: 'space-between' },
  code: { fontSize: 46, lineHeight: 46 },
  teams: { flexDirection: 'row', alignItems: 'baseline' },
  pick: { flexDirection: 'row', alignItems: 'baseline' },
  pickLabel: { width: 92 },
  ratingRow: { flexDirection: 'row', alignItems: 'center' },
  boardRow: { flexDirection: 'row', alignItems: 'center' },
  boardName: { width: 92, marginLeft: 8 },
  points: { width: 26, textAlign: 'right', marginLeft: 8 },
  track: { flex: 1, height: 6, marginLeft: 6, overflow: 'hidden' },
  side: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  serial: { fontSize: 9.5, letterSpacing: 0.4 },
  // Stamps land across the perforation beside the team codes, on blank card clear of every line of text.
  stampAt: { position: 'absolute', top: 54, right: 30 },
});
