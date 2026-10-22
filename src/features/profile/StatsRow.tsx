import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { BallRating, Rule, Stat } from '@/components';
import type { FanStats } from '@/data';
import { pct } from '@/lib/format';
import { useTheme } from '@/theme';
import { thousands } from './constants';

/** Scoreboard row between rules: matches logged · average rating given · prediction accuracy · bragging points. */
export function StatsRow({ stats }: { stats: FanStats }) {
  const avg = stats.avgRatingGiven;
  return (
    <View style={{ marginTop: 22 }}>
      <Rule kind="double" />
      <View style={styles.row}>
        <Cell first>
          <Stat value={String(stats.logged)} label="Logged" />
        </Cell>
        <Cell>
          <Stat value={avg == null ? '–' : avg.toFixed(1)} label="Avg given" />
          {avg != null ? (
            <View style={{ marginTop: 5 }}>
              <BallRating value={Math.round(avg * 2) / 2} size={9} gap={1} />
            </View>
          ) : null}
        </Cell>
        <Cell>
          <Stat value={stats.accuracy == null ? '–' : pct(stats.accuracy)} label="Accuracy" />
        </Cell>
        <Cell>
          <Stat value={thousands(stats.points)} label="Points" />
        </Cell>
      </View>
      <Rule kind="double" />
    </View>
  );
}

function Cell({ children, first }: { children: ReactNode; first?: boolean }) {
  const t = useTheme();
  return <View style={[styles.cell, first ? styles.first : { borderLeftWidth: StyleSheet.hairlineWidth, borderLeftColor: t.ink }]}>{children}</View>;
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', paddingVertical: 12 },
  cell: { flex: 1, paddingLeft: 10, paddingRight: 4 },
  first: { paddingLeft: 0 },
});
