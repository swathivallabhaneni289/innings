import { StyleSheet, View, type ViewStyle } from 'react-native';

import { Txt } from '@/components';
import { verdict } from './draft';

/** "4.5 / 5 · Thriller": the number in stadium type, the out-of in mono, the verdict in plain words. */
export function RatingLine({
  rating,
  size = 'md',
  word,
  style,
}: {
  rating: number;
  size?: 'md' | 'sm';
  word?: string; // defaults to the verdict for the rating
  style?: ViewStyle;
}) {
  const num = size === 'md' ? 34 : 26;
  return (
    <View style={[styles.row, style]} accessible accessibilityLabel={`${rating} out of 5, ${word ?? verdict(rating)}`}>
      <Txt variant="stat" style={{ fontSize: num, lineHeight: num + 2 }}>
        {rating.toFixed(1)}
      </Txt>
      <Txt variant="mono" tone="soft" style={{ marginLeft: 5 }}>
        / 5
      </Txt>
      <Txt variant="mono" tone="faint" style={{ marginHorizontal: 8 }}>
        ·
      </Txt>
      <Txt variant={size === 'md' ? 'h2' : 'h3'} numberOfLines={1} style={{ flexShrink: 1 }}>
        {word ?? verdict(rating)}
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', flexShrink: 1 },
});
