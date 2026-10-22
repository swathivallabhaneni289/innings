// A match ticket that grows with its content. The shared TicketStub needs a fixed height, which the Matches list
// and the Match Page masthead can't promise (status lines and long team names wrap), so this prints the same
// ticket — card stock, dotted perforation, notches, tear-off stub — with flexible height.
import type { ReactNode } from 'react';
import { StyleSheet, View, type ViewStyle } from 'react-native';
import Svg, { Line } from 'react-native-svg';

import type { Match } from '@/domain/types';
import { useTheme } from '@/theme';

const NOTCH = 7;
const LONG = 2000; // the dotted line is drawn long and clipped by its holder

/** "No. 0412" — a deterministic serial per match, so a match prints the same number wherever it appears. */
export function serialFor(match: Pick<Match, 'id'>): string {
  let h = 2166136261;
  for (let i = 0; i < match.id.length; i++) h = Math.imul(h ^ match.id.charCodeAt(i), 16777619);
  return `No. ${String((Math.abs(h) % 9999) + 1).padStart(4, '0')}`;
}

/**
 * Card stock with a perforated stub. `side="right"` tears off on the right (list cards);
 * `side="bottom"` tears off along the bottom (the Match Page masthead). `stub` is the tear-off's content.
 */
export function TicketFrame({
  stock,
  children,
  stub,
  side = 'right',
  stubWidth = 76,
  style,
}: {
  stock: string;
  children: ReactNode;
  stub: ReactNode;
  side?: 'right' | 'bottom';
  stubWidth?: number;
  style?: ViewStyle;
}) {
  const t = useTheme();
  const notch = { backgroundColor: t.paper, width: NOTCH * 2, height: NOTCH * 2, borderRadius: NOTCH };
  if (side === 'bottom') {
    return (
      <View style={[{ backgroundColor: stock }, style]}>
        <View style={styles.mainBottom}>{children}</View>
        <View style={styles.perfRow}>
          <View style={styles.clipRow}>
            <Svg width={LONG} height={2}>
              <Line x1={0} y1={1} x2={LONG} y2={1} stroke={t.cardEdge} strokeWidth={1.6} strokeDasharray="1.5 4" strokeLinecap="round" />
            </Svg>
          </View>
          <View style={[styles.notch, notch, { left: -NOTCH, top: 1 - NOTCH }]} />
          <View style={[styles.notch, notch, { right: -NOTCH, top: 1 - NOTCH }]} />
        </View>
        <View style={styles.stubBottom}>{stub}</View>
      </View>
    );
  }
  return (
    <View style={[styles.row, { backgroundColor: stock }, style]}>
      <View style={styles.mainRight}>{children}</View>
      <View style={styles.perfCol}>
        <View style={styles.clipCol}>
          <Svg width={2} height={LONG}>
            <Line x1={1} y1={0} x2={1} y2={LONG} stroke={t.cardEdge} strokeWidth={1.6} strokeDasharray="1.5 4" strokeLinecap="round" />
          </Svg>
        </View>
        <View style={[styles.notch, notch, { top: -NOTCH, left: 1 - NOTCH }]} />
        <View style={[styles.notch, notch, { bottom: -NOTCH, left: 1 - NOTCH }]} />
      </View>
      <View style={[styles.stubRight, { width: stubWidth }]}>{stub}</View>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row' },
  mainRight: { flex: 1, paddingHorizontal: 14, paddingVertical: 13 },
  perfCol: { width: 2, alignSelf: 'stretch' },
  clipCol: { position: 'absolute', top: NOTCH + 5, bottom: NOTCH + 5, left: 0, width: 2, overflow: 'hidden' },
  stubRight: { paddingVertical: 13, paddingHorizontal: 8, alignItems: 'center', justifyContent: 'space-between' },
  mainBottom: { paddingHorizontal: 16, paddingTop: 16, paddingBottom: 14 },
  perfRow: { height: 2 },
  clipRow: { position: 'absolute', left: NOTCH + 5, right: NOTCH + 5, top: 0, height: 2, overflow: 'hidden' },
  stubBottom: { paddingHorizontal: 16, paddingTop: 12, paddingBottom: 14 },
  notch: { position: 'absolute' },
});
