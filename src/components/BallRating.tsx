import * as Haptics from 'expo-haptics';
import { Pressable, StyleSheet, View } from 'react-native';

import { CricketBall, type BallFill } from './CricketBall';
import { Txt } from './Txt';

const fillFor = (value: number, i: number): BallFill => (value >= i + 1 ? 'full' : value >= i + 0.5 ? 'half' : 'empty');

/**
 * Five cricket balls instead of stars, with half-balls. Pass `onChange` to make it interactive:
 * tapping the left half of a ball gives x.5, the right half gives a whole ball.
 */
export function BallRating({
  value,
  size = 16,
  gap = 3,
  onChange,
  showValue = false,
}: {
  value: number;
  size?: number;
  gap?: number;
  onChange?: (v: number) => void;
  showValue?: boolean;
}) {
  return (
    <View style={styles.row} accessibilityRole={onChange ? 'adjustable' : 'image'} accessibilityLabel={`${value} out of 5 balls`}>
      {[0, 1, 2, 3, 4].map((i) => (
        <View key={i} style={{ width: size, height: size, marginRight: i < 4 ? gap : 0 }}>
          <CricketBall size={size} fill={fillFor(value, i)} />
          {onChange ? (
            <View style={StyleSheet.absoluteFill}>
              <Pressable
                hitSlop={{ top: 10, bottom: 10 }}
                style={styles.half}
                onPress={() => {
                  Haptics.selectionAsync();
                  onChange(i + 0.5);
                }}
              />
              <Pressable
                hitSlop={{ top: 10, bottom: 10 }}
                style={[styles.half, { left: size / 2 }]}
                onPress={() => {
                  Haptics.selectionAsync();
                  onChange(i + 1);
                }}
              />
            </View>
          ) : null}
        </View>
      ))}
      {showValue ? (
        <Txt variant="mono" tone="soft" style={{ marginLeft: 8 }}>
          {value > 0 ? value.toFixed(1) : '–'}
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center' },
  half: { position: 'absolute', top: 0, bottom: 0, left: 0, width: '50%' },
});
