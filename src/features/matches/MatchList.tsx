import { StyleSheet, View } from 'react-native';

import { Txt } from '@/components';
import type { MatchesResponse } from '@/data';
import { useNow } from '@/data';
import { useTheme } from '@/theme';
import { groupMatches } from './group';
import { MatchCard } from './MatchCard';

/** The fixture list, day by day, each match a ticket. Renders nothing for an empty list (the screen shows <Empty>). */
export function MatchList({ data }: { data: MatchesResponse }) {
  const now = useNow(60_000);
  const groups = groupMatches(data.matches, now);
  return (
    <View>
      {groups.map((g) => (
        // 16 here + the last ticket's 12 below it = 28 between days.
        <View key={g.key} style={{ marginTop: 16 }}>
          <DayHeader title={g.title} date={g.date} live={g.live} count={g.matches.length} />
          {g.matches.map((m) => (
            <MatchCard key={m.id} match={m} rating={data.fanRatings[m.id]} now={now} />
          ))}
        </View>
      ))}
    </View>
  );
}

function DayHeader({ title, date, live, count }: { title: string; date?: string; live?: boolean; count: number }) {
  const t = useTheme();
  return (
    <View style={styles.row}>
      <View style={styles.title}>
        {live ? <View style={[styles.dot, { backgroundColor: t.red }]} /> : null}
        <Txt variant="h3" style={{ fontSize: 16 }}>
          {title}
        </Txt>
        {date ? (
          <Txt variant="bodySmall" tone="soft" style={{ marginLeft: 8 }}>
            {date}
          </Txt>
        ) : null}
      </View>
      <Txt variant="monoSmall" tone="faint" style={{ fontSize: 11 }}>
        {count} {count === 1 ? 'match' : 'matches'}
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 },
  title: { flexDirection: 'row', alignItems: 'baseline', flexShrink: 1 },
  dot: { width: 7, height: 7, borderRadius: 4, marginRight: 8, alignSelf: 'center' },
});
