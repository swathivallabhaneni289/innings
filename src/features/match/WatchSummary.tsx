import { StyleSheet, View } from 'react-native';

import { Panel, Txt } from '@/components';
import type { FanSummary } from '@/data';
import { WATCH_PLACES } from '@/domain/constants';
import type { WatchPlace } from '@/domain/types';
import { pct } from '@/lib/format';
import { useTheme } from '@/theme';

const PHRASE: Record<WatchPlace, string> = {
  stadium: 'at the stadium',
  home: 'at home',
  friends: 'with friends',
  hostel: 'in a hostel or PG',
  office: 'at the office',
  cafe: 'at a café or pub',
  'on-the-go': 'on the go',
};

/** AFTER: "62% at home · 21% with friends · 9% at the stadium" over a thin stacked bar with a small key. */
export function WatchSummary({ watch }: { watch: FanSummary['watch'] }) {
  const t = useTheme();
  const rows = watch.filter((w) => w.share > 0);
  if (rows.length === 0) {
    return (
      <Txt variant="bodySmall" tone="soft">
        Nobody has said where they watched this one yet.
      </Txt>
    );
  }
  const swatches = [t.ink, t.inkSoft, t.inkFaint, t.cardEdge, t.rule];
  return (
    <Panel>
      <Txt variant="review">
        {rows
          .slice(0, 3)
          .map((w) => `${pct(w.share)} ${PHRASE[w.place]}`)
          .join(' · ')}
      </Txt>
      <View style={styles.bar}>
        {rows.map((w, i) => (
          <View key={w.place} style={{ flex: w.share, backgroundColor: swatches[i % swatches.length], marginLeft: i === 0 ? 0 : 1.5 }} />
        ))}
      </View>
      <View style={styles.key}>
        {rows.map((w, i) => {
          const place = WATCH_PLACES.find((p) => p.id === w.place);
          return (
            <View key={w.place} style={styles.keyItem}>
              <View style={[styles.swatch, { backgroundColor: swatches[i % swatches.length], borderColor: t.rule }]} />
              <Txt variant="bodySmall" tone="soft">
                {place?.label ?? w.place}{' '}
                <Txt variant="monoSmall" tone="soft">
                  {pct(w.share)}
                </Txt>
              </Txt>
            </View>
          );
        })}
      </View>
    </Panel>
  );
}

const styles = StyleSheet.create({
  bar: { flexDirection: 'row', height: 8, marginTop: 14 },
  key: { flexDirection: 'row', flexWrap: 'wrap', marginTop: 10 },
  keyItem: { flexDirection: 'row', alignItems: 'center', marginRight: 14, marginBottom: 6 },
  swatch: { width: 8, height: 8, borderWidth: StyleSheet.hairlineWidth, marginRight: 6 },
});
