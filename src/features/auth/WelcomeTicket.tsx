import { useFocusEffect } from 'expo-router';
import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { Pressable, StyleSheet, useWindowDimensions, View, type LayoutChangeEvent, type ViewStyle } from 'react-native';
import Animated, {
  Easing,
  interpolateColor,
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withDelay,
  withTiming,
} from 'react-native-reanimated';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import Svg, { Line } from 'react-native-svg';

import { BallRating, Barcode, Button, Rise, Stamp, Txt } from '@/components';
import { verdict } from '@/features/log/draft';
import { fonts, useTheme, type StockKey } from '@/theme';

const CYCLE = 4800; // one match on the ticket, long enough to read its fun fact
const ROLL = 640; // the team codes rolling over, like a scoreboard
const BALLS_AT = 560;
const BALL_EVERY = 110;
const STAMP_AT = 1900;
const EASE = Easing.bezier(0.16, 1, 0.3, 1);
const NOTCH = 11;
const STAMP_W = 150; // room kept clear on the right of the codes for LOGGED
const TEXT_SCALE = 1.2; // how far the small print grows with the system text size

/** The matches the ticket prints in turn, each on its competition's card stock, with where it was played and a fact. */
const MATCHES: { a: string; b: string; line: string; venue: string; rating: number; stock: StockKey; fact: string }[] = [
  {
    a: 'MI',
    b: 'CSK',
    line: 'IPL · Match 21',
    venue: 'Chepauk, Chennai',
    rating: 4.5,
    stock: 'ipl',
    fact: 'Mumbai and Chennai have met in four IPL finals. Mumbai won three of them.',
  },
  {
    a: 'IND',
    b: 'AUS',
    line: '3rd T20I',
    venue: 'Chinnaswamy Stadium, Bengaluru',
    rating: 4,
    stock: 'international',
    fact: 'Australia played in the very first men’s T20 international, against New Zealand in 2005.',
  },
  {
    a: 'ENG',
    b: 'IND',
    line: '2nd Test · Day 5',
    venue: 'Lord’s, London',
    rating: 5,
    stock: 'test',
    fact: 'Every Test hundred scored at Lord’s goes up on the honours boards in its dressing rooms.',
  },
  {
    a: 'SL',
    b: 'IND',
    line: 'World Cup final · 2011',
    venue: 'Wankhede Stadium, Mumbai',
    rating: 5,
    stock: 'world-cup',
    fact: 'MS Dhoni won it with a six and finished on 91 not out.',
  },
];

/**
 * The Welcome screen as one big match ticket, edge to edge: the card stock is the screen, the team codes are the
 * headline, and the way in sits on the stub below the perforation. Every few seconds the ticket prints the next
 * match: the codes roll over, your rating fills in ball by ball, LOGGED lands beside the codes, the stub number
 * goes up and the stub shows a fun fact about the match. It only moves while the screen is in front.
 * `lead` holds the first match for the launch splash.
 */
export function WelcomeTicket({ lead = 0, onStart, onSignIn }: { lead?: number; onStart: () => void; onSignIn: () => void }) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const { width, fontScale } = useWindowDimensions();
  const grow = Math.min(Math.max(fontScale, 1), TEXT_SCALE);
  const [heroH, setHeroH] = useState(0);
  // `n` counts every new ticket, so each mounts fresh; `prev` stays mounted while it rolls away.
  const [step, setStep] = useState<{ i: number; n: number; prev: number | null }>({ i: 0, n: 0, prev: null });
  const hold = step.n === 0 ? lead : 0;

  useFocusEffect(
    useCallback(() => {
      const id = setTimeout(() => setStep({ i: (step.i + 1) % MATCHES.length, n: step.n + 1, prev: step.i }), hold + CYCLE);
      return () => clearTimeout(id);
    }, [hold, step]),
  );

  // The card stock washes from one match's colour to the next as the codes roll.
  const to = t.stock[MATCHES[step.i].stock];
  const from = step.prev === null ? to : t.stock[MATCHES[step.prev].stock];
  const c = useSharedValue(1);
  const fromC = useSharedValue(to);
  const toC = useSharedValue(to);
  useEffect(() => {
    fromC.set(from);
    toC.set(to);
    c.set(0);
    c.set(withTiming(1, { duration: 760, easing: EASE }));
  }, [step.n, from, to, c, fromC, toC]);
  const wash = useAnimatedStyle(() => ({ backgroundColor: interpolateColor(c.value, [0, 1], [fromC.value, toC.value]) }));

  // The codes are as big as the space allows: tall enough to fill the hero, narrow enough to leave LOGGED its room.
  const venueH = Math.ceil(26 * grow); // a line of text, with room for the pin emoji, which sits taller
  const rest = 16 + 10 + 38 + 12 + venueH + 18 + Math.ceil(14 * grow) + 6 + 34;
  const code = Math.floor(Math.min(190, Math.max(64, (heroH - rest - 36) / 1.68), (width - 40 - STAMP_W - 12) / 1.24));
  // Each code's slot is about as tall as its capitals, so the two lines sit tight around the "vs", with a little
  // room above and below for the round letters (C, G, O, S), which rise over the flat tops and dip under the line.
  const lh = Math.round(code * 0.84);
  const factH = Math.ceil(36 * grow);
  const slots =
    step.prev === null
      ? [{ i: step.i, n: step.n, leaving: false }]
      : [
          { i: step.prev, n: step.n - 1, leaving: true },
          { i: step.i, n: step.n, leaving: false },
        ];
  const first = (n: number) => (n === 0 ? lead : 0);
  // A roll waits for the splash only on the very first arrival; leaving, it moves in step with the one coming in.
  const lead0 = (s: { n: number; leaving: boolean }) => (s.leaving ? 0 : first(s.n));

  return (
    <Animated.View style={[styles.page, wash, { paddingTop: insets.top + 8, paddingBottom: Math.max(insets.bottom, 12) + 4 }]}>
      <Rise delay={lead} style={styles.top}>
        <View>
          <Txt accessibilityRole="header" allowFontScaling={false} style={[styles.wordmark, { color: t.ink }]}>
            innings
          </Txt>
          <Txt variant="label" tone="soft" maxFontSizeMultiplier={TEXT_SCALE}>
            A matchday diary
          </Txt>
        </View>
        <View style={styles.topRight}>
          <Txt variant="label" maxFontSizeMultiplier={TEXT_SCALE}>
            Admit one
          </Txt>
          <Mask h={18} style={styles.serialSlot}>
            {slots.map((s) => (
              <Roll key={s.n} h={18} delay={lead0(s) + 200} leaving={s.leaving}>
                <Txt variant="mono" tone="red" allowFontScaling={false} style={styles.serial}>
                  No. {String(s.n + 1).padStart(4, '0')}
                </Txt>
              </Roll>
            ))}
          </Mask>
        </View>
      </Rise>

      <View
        style={styles.hero}
        onLayout={(e: LayoutChangeEvent) => setHeroH(e.nativeEvent.layout.height)}
        accessible
        accessibilityRole="image"
        accessibilityLabel="A match ticket: log every match you watch, rate it out of five cricket balls and keep the stub.">
        {heroH > 0 ? (
          <View importantForAccessibility="no-hide-descendants" accessibilityElementsHidden>
            <Mask h={16}>
              {slots.map((s) => (
                <Roll key={s.n} h={16} delay={lead0(s) + 40} leaving={s.leaving}>
                  <Txt variant="label" numberOfLines={1} maxFontSizeMultiplier={1}>
                    {MATCHES[s.i].line}
                  </Txt>
                </Roll>
              ))}
            </Mask>

            <View style={{ marginTop: 10 }}>
              <Mask h={lh}>
                {slots.map((s) => (
                  <Roll key={s.n} h={lh} delay={lead0(s) + 80} leaving={s.leaving}>
                    <Code size={code} text={MATCHES[s.i].a} />
                  </Roll>
                ))}
              </Mask>
              <Txt allowFontScaling={false} style={[styles.vs, { color: t.ink }]}>
                vs
              </Txt>
              <Mask h={lh}>
                {slots.map((s) => (
                  <Roll key={s.n} h={lh} delay={lead0(s) + 150} leaving={s.leaving}>
                    <Code size={code} text={MATCHES[s.i].b} />
                  </Roll>
                ))}
              </Mask>
              {/* LOGGED lands on bare card to the right of the codes, level with the "vs". */}
              <View pointerEvents="none" style={styles.stampAt}>
                {slots.map((s) => (
                  <StampSlot key={s.n} at={first(s.n) + STAMP_AT} leaving={s.leaving} />
                ))}
              </View>
            </View>

            <View style={{ height: venueH, marginTop: 12 }}>
              {slots.map((s) => (
                <Swap key={s.n} delay={lead0(s) + 220} leaving={s.leaving}>
                  <Txt variant="body" numberOfLines={1} maxFontSizeMultiplier={TEXT_SCALE}>
                    📍 {MATCHES[s.i].venue}
                  </Txt>
                </Swap>
              ))}
            </View>

            <Txt variant="label" tone="soft" maxFontSizeMultiplier={TEXT_SCALE} style={{ marginTop: 18 }}>
              Your rating
            </Txt>
            <View style={styles.ratingRow}>
              {slots.map((s) => (
                <Rating key={s.n} value={MATCHES[s.i].rating} at={first(s.n) + BALLS_AT} leaving={s.leaving} />
              ))}
            </View>
          </View>
        ) : null}
      </View>

      <Perforation colour={t.ink} edge={t.cardEdge} />

      <Rise delay={lead + 120}>
        <View style={styles.stubRow}>
          <View style={styles.fact}>
            <Txt variant="label" tone="red" maxFontSizeMultiplier={TEXT_SCALE}>
              Fun fact
            </Txt>
            <View style={{ height: factH, marginTop: 4 }}>
              {slots.map((s) => (
                <Swap key={s.n} delay={lead0(s) + 280} leaving={s.leaving}>
                  <Txt variant="bodySmall" numberOfLines={2} adjustsFontSizeToFit minimumFontScale={0.85} maxFontSizeMultiplier={TEXT_SCALE}>
                    {MATCHES[s.i].fact}
                  </Txt>
                </Swap>
              ))}
            </View>
          </View>
          <Barcode seed="innings-admit-one" width={46} height={factH + 18} />
        </View>
        <Button label="Get started" size="lg" onPress={onStart} maxFontSizeMultiplier={1.25} />
        <Pressable onPress={onSignIn} accessibilityRole="button" style={({ pressed }) => [styles.secondary, { opacity: pressed ? 0.6 : 1 }]}>
          <Txt variant="bodySmall" maxFontSizeMultiplier={1.25} style={{ fontFamily: fonts.label }}>
            I already have an account
          </Txt>
        </Pressable>
      </Rise>
    </Animated.View>
  );
}

/** A window one line tall: whatever rolls inside it is cut off at its edges, like a scoreboard's slot. */
function Mask({ h, style, children }: { h: number; style?: ViewStyle; children: ReactNode }) {
  return <View style={[{ height: h, overflow: 'hidden' }, style]}>{children}</View>;
}

/**
 * Rolls up into its slot after `delay`, and once `leaving`, on up and out of it after `delay` again: the one going
 * and the one coming move together on the same curve, a slot apart, so they never overlap.
 */
function Roll({ h, delay, leaving, children }: { h: number; delay: number; leaving: boolean; children: ReactNode }) {
  const p = useSharedValue(0);
  const q = useSharedValue(0);
  useEffect(() => {
    p.set(withDelay(delay, withTiming(1, { duration: ROLL, easing: EASE })));
  }, [delay, p]);
  useEffect(() => {
    if (leaving) q.set(withDelay(delay, withTiming(1, { duration: ROLL, easing: EASE })));
  }, [leaving, delay, q]);
  const anim = useAnimatedStyle(() => ({ transform: [{ translateY: (1 - p.value) * h * 1.08 - q.value * h * 1.08 }] }));
  return <Animated.View style={[StyleSheet.absoluteFill, anim]}>{children}</Animated.View>;
}

/** For lines that wrap or grow with the text size: fades up into place after `delay`, and fades up and out once `leaving`. */
function Swap({ delay, leaving, children }: { delay: number; leaving: boolean; children: ReactNode }) {
  const p = useSharedValue(0);
  const q = useSharedValue(0);
  useEffect(() => {
    p.set(withDelay(delay, withTiming(1, { duration: 420, easing: EASE })));
  }, [delay, p]);
  useEffect(() => {
    if (leaving) q.set(withTiming(1, { duration: 200, easing: Easing.in(Easing.quad) }));
  }, [leaving, q]);
  const anim = useAnimatedStyle(() => ({
    opacity: p.value * (1 - q.value),
    transform: [{ translateY: (1 - p.value) * 8 - q.value * 6 }],
  }));
  return <Animated.View style={[StyleSheet.absoluteFill, anim]}>{children}</Animated.View>;
}

/**
 * A team code in the big display face, set a little down from the top of its slot so the round letters keep their
 * tops. Its line box runs on past the slot, below the capitals, and the slot cuts that empty part off.
 */
function Code({ size, text }: { size: number; text: string }) {
  const t = useTheme();
  return (
    <Txt
      allowFontScaling={false}
      numberOfLines={1}
      style={{ fontFamily: fonts.display, fontSize: size, lineHeight: Math.round(size * 1.02), paddingTop: Math.round(size * 0.04), color: t.ink, letterSpacing: -1 }}>
      {text}
    </Txt>
  );
}

/**
 * Your rating, filling in half a ball at a time from `at` ms, then the word for it, as the app says it ("4.5 / 5 ·
 * Thriller"). The row fades up just before the balls start and fades out when the ticket rolls on.
 */
function Rating({ value, at, leaving }: { value: number; at: number; leaving: boolean }) {
  const reduce = useReducedMotion();
  const [v, setV] = useState(reduce ? value : 0);
  useEffect(() => {
    if (reduce) return;
    let tick: ReturnType<typeof setInterval> | undefined;
    let shown = 0;
    const begin = setTimeout(() => {
      tick = setInterval(() => {
        shown = Math.min(value, shown + 0.5);
        setV(shown);
        if (shown >= value) clearInterval(tick);
      }, BALL_EVERY);
    }, at);
    return () => {
      clearTimeout(begin);
      clearInterval(tick);
    };
  }, [reduce, value, at]);
  const o = useSharedValue(0);
  useEffect(() => {
    o.set(withDelay(Math.max(0, at - 160), withTiming(1, { duration: 160 })));
  }, [at, o]);
  useEffect(() => {
    if (leaving) o.set(withTiming(0, { duration: 160 }));
  }, [leaving, o]);
  const anim = useAnimatedStyle(() => ({ opacity: o.value }));
  return (
    <Animated.View style={[StyleSheet.absoluteFill, styles.rating, anim]}>
      <BallRating value={v} size={24} gap={5} />
      <Txt allowFontScaling={false} style={styles.score}>
        {v.toFixed(1)}
      </Txt>
      <Txt variant="mono" tone="soft" allowFontScaling={false} style={{ marginLeft: 4 }}>
        / 5
      </Txt>
      {v >= value ? (
        <Rise style={styles.verdict}>
          <Txt variant="h3" numberOfLines={1} adjustsFontSizeToFit minimumFontScale={0.8} maxFontSizeMultiplier={1}>
            {verdict(value)}
          </Txt>
        </Rise>
      ) : null}
    </Animated.View>
  );
}

/** LOGGED, landing straight at `at` ms and lifting off when the ticket rolls on. */
function StampSlot({ at, leaving }: { at: number; leaving: boolean }) {
  const o = useSharedValue(0);
  useEffect(() => {
    o.set(withDelay(at, withTiming(1, { duration: 0 })));
  }, [at, o]);
  useEffect(() => {
    if (leaving) o.set(withTiming(0, { duration: 200 }));
  }, [leaving, o]);
  const anim = useAnimatedStyle(() => ({ opacity: o.value }));
  return (
    <Animated.View style={[StyleSheet.absoluteFill, styles.stampBox, anim]}>
      <Stamp visible delay={at} size="md" angle={0} />
    </Animated.View>
  );
}

/** The tear line across the whole screen, with a notch bitten out of each edge. */
function Perforation({ colour, edge }: { colour: string; edge: string }) {
  const { width } = useWindowDimensions();
  return (
    <View style={styles.perf} pointerEvents="none">
      <Svg width={width} height={NOTCH * 2} style={styles.perfLine}>
        <Line x1={NOTCH + 6} y1={NOTCH} x2={width - NOTCH - 6} y2={NOTCH} stroke={edge} strokeWidth={1.8} strokeDasharray="2 5" strokeLinecap="round" />
      </Svg>
      <View style={[styles.notch, { left: -NOTCH, backgroundColor: colour }]} />
      <View style={[styles.notch, { right: -NOTCH, backgroundColor: colour }]} />
    </View>
  );
}

const styles = StyleSheet.create({
  page: { flex: 1, paddingHorizontal: 20 },
  top: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'flex-start' },
  wordmark: { fontFamily: fonts.heading, fontSize: 28, lineHeight: 32, letterSpacing: -0.7 },
  topRight: { alignItems: 'flex-end', marginTop: 9 },
  serialSlot: { width: 96, marginTop: 4 },
  serial: { fontSize: 14, lineHeight: 18, letterSpacing: 0.6, textAlign: 'right' },
  hero: { flex: 1, justifyContent: 'center' },
  vs: { fontFamily: fonts.heading, fontSize: 22, lineHeight: 26, marginVertical: 6 },
  stampAt: { position: 'absolute', right: 0, top: 0, bottom: 0, width: STAMP_W },
  stampBox: { alignItems: 'center', justifyContent: 'center' },
  ratingRow: { height: 34, marginTop: 6 },
  rating: { flexDirection: 'row', alignItems: 'center' },
  score: { fontFamily: fonts.display, fontSize: 30, lineHeight: 34, marginLeft: 12 },
  verdict: { flexShrink: 1, marginLeft: 10 },
  perf: { height: NOTCH * 2, marginHorizontal: -20, marginBottom: 14 },
  perfLine: { position: 'absolute', left: 0, top: 0 },
  notch: { position: 'absolute', top: 0, width: NOTCH * 2, height: NOTCH * 2, borderRadius: NOTCH },
  stubRow: { flexDirection: 'row', alignItems: 'flex-start', marginBottom: 16 },
  fact: { flex: 1, marginRight: 16 },
  secondary: { minHeight: 48, alignItems: 'center', justifyContent: 'center', marginTop: 4 },
});
