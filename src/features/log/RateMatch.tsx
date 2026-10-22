import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useState } from 'react';
import { Pressable, StyleSheet, useWindowDimensions, View } from 'react-native';

import { BallRating, Panel, SpinningBall, Txt } from '@/components';
import { useTheme } from '@/theme';
import { RatingLine } from './RatingLine';

/** Step: the big ball rating (half-balls) and the optional ❤️ instant-classic tag. */
export function RateMatch({
  rating,
  instantClassic,
  onRating,
  onInstantClassic,
  error,
}: {
  rating: number;
  instantClassic: boolean;
  onRating: (v: number) => void;
  onInstantClassic: (v: boolean) => void;
  error?: string | null;
}) {
  const t = useTheme();
  const { width } = useWindowDimensions();
  const ball = Math.min(50, Math.floor((width - 40 - 32 - 4 * 10) / 5)); // fit 5 balls inside the card
  const [spinKey, setSpinKey] = useState<number | undefined>(undefined); // undefined = no spin on mount

  const change = (v: number) => {
    setSpinKey((k) => (k ?? 0) + 1);
    onRating(v);
  };

  return (
    <View>
      <Panel style={styles.panel}>
        <BallRating value={rating} size={ball} gap={10} onChange={change} />
        <View style={styles.valueRow}>
          {rating > 0 ? (
            <>
              <SpinningBall spinKey={spinKey} size={20} />
              <RatingLine rating={rating} style={{ marginLeft: 10 }} />
            </>
          ) : (
            <Txt variant="h2" tone="faint">
              Not rated yet
            </Txt>
          )}
        </View>
        <Txt variant="bodySmall" tone="faint" align="center" style={{ marginTop: 6 }}>
          Tap the left half of a ball for a half-ball.
        </Txt>
        {error ? (
          <Txt variant="bodySmall" tone="red" align="center" style={{ marginTop: 8 }}>
            {error}
          </Txt>
        ) : null}
      </Panel>

      <Pressable
        onPress={() => {
          Haptics.selectionAsync();
          onInstantClassic(!instantClassic);
        }}
        accessibilityRole="checkbox"
        accessibilityState={{ checked: instantClassic }}
        style={({ pressed }) => [styles.toggle, { backgroundColor: t.card, opacity: pressed ? 0.8 : 1 }]}>
        <Txt style={{ fontSize: 22 }}>❤️</Txt>
        <View style={{ flex: 1, marginLeft: 12 }}>
          <Txt variant="h3">Instant classic</Txt>
          <Txt variant="bodySmall" tone="soft">
            You'll still be talking about this one in ten years.
          </Txt>
        </View>
        <View style={[styles.box, { borderColor: instantClassic ? t.ink : t.inkFaint, backgroundColor: instantClassic ? t.ink : 'transparent' }]}>
          {instantClassic ? <Feather name="check" size={16} color={t.onInk} /> : null}
        </View>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  panel: { alignItems: 'center', paddingTop: 26, paddingBottom: 20, marginTop: 12 },
  valueRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center', minHeight: 40, marginTop: 20 },
  toggle: { flexDirection: 'row', alignItems: 'center', padding: 16, marginTop: 12 },
  box: { width: 24, height: 24, borderWidth: 1.5, alignItems: 'center', justifyContent: 'center', marginLeft: 12 },
});
