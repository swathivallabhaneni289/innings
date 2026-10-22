import { StyleSheet, View } from 'react-native';

import { Empty, Panel, Rule, Txt } from '@/components';
import type { FanStats } from '@/data';
import { WATCH_PLACES } from '@/domain/constants';
import { fonts, useTheme } from '@/theme';

/** Where the fan watched: one inked bar per spot, then the grounds they've been to. */
export function WatchBreakdown({ stats, isMe }: { stats: FanStats; isMe: boolean }) {
  const t = useTheme();
  const max = Math.max(1, ...stats.watch.map((w) => w.count));

  return (
    <Panel>
      {stats.watch.length === 0 ? (
        <Empty
          title="No grounds, no sofas yet"
          body={isMe ? 'Log a match and say where you watched it.' : 'No watching spots logged yet.'}
        />
      ) : (
        stats.watch.map((w) => {
          const place = WATCH_PLACES.find((p) => p.id === w.place);
          return (
            <View key={w.place} style={styles.row}>
              <Txt variant="bodySmall" numberOfLines={1} style={styles.label}>
                {place?.label ?? w.place}
              </Txt>
              <View style={[styles.track, { backgroundColor: t.rule }]}>
                <View style={{ width: `${(w.count / max) * 100}%`, height: '100%', backgroundColor: t.ink }} />
              </View>
              <Txt variant="mono" align="right" style={styles.count}>
                {w.count}
              </Txt>
            </View>
          );
        })
      )}

      <Rule style={{ marginTop: 14, marginBottom: 12 }} />
      <View style={styles.stadiumHead}>
        <Txt variant="h3" style={{ flex: 1 }}>
          Grounds visited
        </Txt>
        <Txt variant="mono">{stats.stadiums.length}</Txt>
      </View>
      {stats.stadiums.length === 0 ? (
        <Txt variant="bodySmall" tone="soft" style={{ marginTop: 6 }}>
          {isMe ? 'Not been to the ground yet. The stands are calling.' : 'Not been to the ground yet.'}
        </Txt>
      ) : (
        stats.stadiums.map((v, i) => (
          <View
            key={v.id}
            style={[styles.venue, { borderTopColor: t.rule, borderTopWidth: i === 0 ? 0 : StyleSheet.hairlineWidth }]}>
            <Txt variant="bodySmall" numberOfLines={1} style={{ flex: 1, fontFamily: fonts.ui }}>
              {v.name}
            </Txt>
            <Txt variant="bodySmall" tone="soft" numberOfLines={1} style={{ marginLeft: 8 }}>
              {v.city}
            </Txt>
          </View>
        ))
      )}
    </Panel>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 5 },
  label: { width: 104 },
  track: { flex: 1, height: 8, marginHorizontal: 8 },
  count: { width: 28 },
  stadiumHead: { flexDirection: 'row', alignItems: 'baseline', marginBottom: 4 },
  venue: { flexDirection: 'row', alignItems: 'baseline', paddingVertical: 8 },
});
