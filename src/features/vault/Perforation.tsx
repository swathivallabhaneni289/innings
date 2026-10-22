import { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import Svg, { Line } from 'react-native-svg';

import { useTheme } from '@/theme';

const NOTCH = 7;

/**
 * The tear line across a ticket printed on card stock: a dotted rule with a notch bitten out of each edge.
 * The notches are painted in the ground colour, so use it on tickets that sit straight on the screen ground.
 * Pass `inset` when the ticket's content is padded, so the line runs edge to edge.
 */
export function Perforation({ inset = 0 }: { inset?: number }) {
  const t = useTheme();
  const [w, setW] = useState(0);
  return (
    <View
      style={[styles.wrap, { marginHorizontal: -inset }]}
      onLayout={(e) => setW(Math.round(e.nativeEvent.layout.width))}
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants">
      {w > 0 ? (
        <Svg width={w} height={2} style={styles.line}>
          <Line x1={NOTCH + 4} y1={1} x2={w - NOTCH - 4} y2={1} stroke={t.cardEdge} strokeWidth={1.6} strokeDasharray="1.5 4" strokeLinecap="round" />
        </Svg>
      ) : null}
      <View style={[styles.notch, styles.left, { backgroundColor: t.paper }]} />
      <View style={[styles.notch, styles.right, { backgroundColor: t.paper }]} />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { height: NOTCH * 2, justifyContent: 'center' },
  line: { position: 'absolute', top: NOTCH - 1, left: 0 },
  notch: { position: 'absolute', top: 0, width: NOTCH * 2, height: NOTCH * 2, borderRadius: NOTCH },
  left: { left: -NOTCH },
  right: { right: -NOTCH },
});
