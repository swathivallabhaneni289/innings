import { StyleSheet, View } from 'react-native';

import { BallRating, Empty, Panel, Txt } from '@/components';
import type { FanSummary } from '@/data';
import { useTheme } from '@/theme';

const BAR_MAX = 72;

const thousands = (n: number) => String(n).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
const compact = (n: number) => (n >= 10_000 ? `${Math.round(n / 1000)}k` : n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n));

/** AFTER: the crowd's verdict — big average, five balls, and the ten half-ball buckets as a bar chart. */
export function FanRating({ fans }: { fans: FanSummary }) {
  const t = useTheme();
  if (fans.ratingCount === 0) {
    return <Empty title="No ratings yet" body="Nobody has rated this one. Log it and set the tone." />;
  }
  const buckets = fans.buckets.length === 10 ? fans.buckets : Array.from({ length: 10 }, (_, i) => fans.buckets[i] ?? 0);
  const max = Math.max(1, ...buckets);
  const peak = buckets.indexOf(Math.max(...buckets));
  return (
    <Panel>
      <View style={styles.head}>
        <Txt variant="display" style={styles.avg}>
          {fans.ratingAvg.toFixed(1)}
        </Txt>
        <View style={{ marginLeft: 14, flex: 1 }}>
          <BallRating value={fans.ratingAvg} size={22} gap={4} />
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 6 }}>
            Out of 5, from {thousands(fans.ratingCount)} {fans.ratingCount === 1 ? 'rating' : 'ratings'}
          </Txt>
        </View>
      </View>

      <View style={styles.chart} accessibilityLabel="Distribution of fan ratings from half a ball to five balls">
        {buckets.map((count, i) => (
          <View key={i} style={styles.col}>
            <Txt variant="monoSmall" tone={i === peak ? 'ink' : 'faint'} style={styles.count}>
              {count > 0 ? compact(count) : ''}
            </Txt>
            <View
              style={{
                width: '100%',
                height: count > 0 ? Math.max(2, Math.round((count / max) * BAR_MAX)) : 1,
                backgroundColor: count === 0 ? t.rule : i === peak ? t.red : t.ink,
              }}
            />
          </View>
        ))}
      </View>
      <View style={[styles.axis, { borderTopColor: t.rule }]}>
        {buckets.map((_, i) => (
          <View key={i} style={styles.col}>
            <Txt variant="monoSmall" tone="faint" style={styles.tick}>
              {i % 2 === 1 ? String((i + 1) / 2) : ''}
            </Txt>
          </View>
        ))}
      </View>
      <Txt variant="label" tone="faint" align="right" style={{ fontSize: 9, marginTop: 2 }}>
        Balls
      </Txt>
    </Panel>
  );
}

const styles = StyleSheet.create({
  head: { flexDirection: 'row', alignItems: 'center' },
  avg: { fontSize: 64, lineHeight: 70 },
  chart: { flexDirection: 'row', alignItems: 'flex-end', height: BAR_MAX + 18, marginTop: 14 },
  col: { flex: 1, alignItems: 'center', justifyContent: 'flex-end', marginHorizontal: 2 },
  count: { fontSize: 9, lineHeight: 12, marginBottom: 3 },
  axis: { flexDirection: 'row', borderTopWidth: StyleSheet.hairlineWidth, paddingTop: 4 },
  tick: { fontSize: 9.5, lineHeight: 12 },
});
