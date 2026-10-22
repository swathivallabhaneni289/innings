import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { BallRating, Txt } from '@/components';
import { getTeam } from '@/data';
import type { RatedMatch } from '@/domain/api';
import { competitionShort } from '@/domain/constants';
import { fonts, useTheme } from '@/theme';
import { matchYear } from './top5';

/**
 * One line of the all-time fan rankings: the rank in big stadium type, the match in plain type,
 * the fans' average in mono beside the balls. Rows sit in a white card; `last` drops the hairline under the final row.
 */
export function RankRow({ rank, item, compact = false, last = false }: { rank: number; item: RatedMatch; compact?: boolean; last?: boolean }) {
  const t = useTheme();
  const { match } = item;
  const [a, b] = match.teamIds.map(getTeam);
  const competition = match.classic?.competitionName ?? competitionShort(match.competitionId);
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={`Number ${rank}: ${a.name} versus ${b.name}, ${matchYear(match)}`}
      onPress={() => router.push(`/match/${match.id}`)}
      style={({ pressed }) => [
        styles.row,
        { borderBottomColor: t.rule, borderBottomWidth: last ? 0 : StyleSheet.hairlineWidth, opacity: pressed ? 0.7 : 1 },
      ]}>
      <Txt variant="display" style={styles.rank} numberOfLines={1} adjustsFontSizeToFit>
        {rank}
      </Txt>
      <View style={{ flex: 1 }}>
        <View style={styles.titleRow}>
          <Txt variant="h3" numberOfLines={1} style={{ flex: 1 }}>
            {a.short}
            <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label }}>
              {'  vs  '}
            </Txt>
            {b.short}
          </Txt>
          <Txt variant="mono" tone="soft" style={{ marginLeft: 8 }}>
            {matchYear(match)}
          </Txt>
        </View>
        {!compact && match.classic?.headline ? (
          <Txt variant="bodySmall" numberOfLines={2} style={{ marginTop: 2 }}>
            {match.classic.headline}
          </Txt>
        ) : null}
        <Txt variant="bodySmall" tone="soft" numberOfLines={1} style={{ marginTop: 1 }}>
          {match.stage} · {competition}
        </Txt>
        <View style={styles.ratingRow}>
          <BallRating value={item.ratingAvg} size={13} gap={2} />
          <Txt variant="mono" style={{ marginLeft: 8 }}>
            {item.ratingCount > 0 ? item.ratingAvg.toFixed(2) : '–'}
          </Txt>
          <Txt variant="monoSmall" tone="faint" style={{ marginLeft: 6 }}>
            {item.ratingCount.toLocaleString()} {item.ratingCount === 1 ? 'fan' : 'fans'}
          </Txt>
        </View>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'flex-start', paddingVertical: 14 },
  rank: { width: 48, fontSize: 44, lineHeight: 44, marginTop: -2 },
  titleRow: { flexDirection: 'row', alignItems: 'baseline' },
  ratingRow: { flexDirection: 'row', alignItems: 'center', marginTop: 8 },
});
