import { useEffect, type ReactNode } from 'react';
import { StyleSheet, View, type ViewStyle } from 'react-native';
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withDelay,
  withSequence,
  withSpring,
  withTiming,
} from 'react-native-reanimated';

import { useTheme, fonts } from '@/theme';
import { CricketBall } from './CricketBall';

/** A cricket ball that does a quick spin every time `spinKey` changes (rating submitted). */
export function SpinningBall({ spinKey, size = 28 }: { spinKey: unknown; size?: number }) {
  const rot = useSharedValue(0);
  useEffect(() => {
    if (spinKey === undefined) return;
    rot.value = 0;
    rot.value = withTiming(720, { duration: 700, easing: Easing.out(Easing.cubic) });
  }, [spinKey, rot]);
  const style = useAnimatedStyle(() => ({ transform: [{ rotate: `${rot.value}deg` }] }));
  return (
    <Animated.View style={style}>
      <CricketBall size={size} />
    </Animated.View>
  );
}

/**
 * Fades its children in and lifts them into place once, `delay` ms after mounting, so a screen's lines can
 * arrive in turn. With Reduce Motion on, Reanimated skips straight to the end.
 */
export function Rise({ delay = 0, children, style }: { delay?: number; children: ReactNode; style?: ViewStyle }) {
  const p = useSharedValue(0);
  useEffect(() => {
    p.value = withDelay(delay, withTiming(1, { duration: 380, easing: Easing.out(Easing.cubic) }));
  }, [delay, p]);
  const anim = useAnimatedStyle(() => ({ opacity: p.value, transform: [{ translateY: (1 - p.value) * 14 }] }));
  return <Animated.View style={[style, anim]}>{children}</Animated.View>;
}

/**
 * The rubber "LOGGED" stamp: slams down from above, overshoots, settles at `angle` degrees.
 * Render it over a ticket stub; it animates when `visible` turns true.
 */
export function Stamp({
  visible,
  text = 'LOGGED',
  size = 'md',
  delay = 0,
  angle = -11,
}: {
  visible: boolean;
  text?: string;
  size?: 'sm' | 'md' | 'lg';
  delay?: number;
  angle?: number;
}) {
  const t = useTheme();
  const scale = useSharedValue(visible ? 1 : 2.4);
  const opacity = useSharedValue(visible ? 0.92 : 0);
  useEffect(() => {
    if (visible) {
      scale.value = 2.4;
      opacity.value = 0;
      scale.value = withDelay(delay, withSequence(withTiming(0.92, { duration: 170, easing: Easing.in(Easing.quad) }), withSpring(1, { damping: 7, stiffness: 260 })));
      opacity.value = withDelay(delay, withTiming(0.92, { duration: 140 }));
    } else {
      opacity.value = withTiming(0, { duration: 120 });
    }
  }, [visible, delay, scale, opacity]);
  const anim = useAnimatedStyle(() => ({ opacity: opacity.value, transform: [{ rotate: `${angle}deg` }, { scale: scale.value }] }));
  const fontSize = { sm: 14, md: 24, lg: 34 }[size];
  return (
    <Animated.View pointerEvents="none" style={[styles.stamp, { borderColor: t.red }, anim]}>
      <View style={[styles.inner, { borderColor: t.red }]}>
        {/* A stamp is a graphic: it keeps its size whatever the system text size, so it never grows over the ticket's lines. */}
        <Animated.Text allowFontScaling={false} style={{ fontFamily: fonts.display, fontSize, letterSpacing: fontSize / 6, color: t.red }}>
          {text}
        </Animated.Text>
      </View>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  stamp: { borderWidth: 2.5, padding: 3, alignSelf: 'center' },
  inner: { borderWidth: 1, paddingHorizontal: 10, paddingVertical: 2, alignItems: 'center' },
});
