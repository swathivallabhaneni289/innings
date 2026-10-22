import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { BallRating, TicketStub, Txt } from '@/components';
import { getTeam } from '@/data';
import type { RatedMatch } from '@/domain/api';
import { fonts, stockFor, useTheme } from '@/theme';
import { matchYear } from './top5';

export const CLASSIC_CARD_WIDTH = 200;
const HEIGHT = 224;
const STUB = 42;

/** A classic on the collection shelf, printed as a small ticket: the year, the codes, the hook; the fan rating on the stub. */
export function ClassicCard({ item }: { item: RatedMatch }) {
  const t = useTheme();
  const { match } = item;
  const [a, b] = match.teamIds.map(getTeam);
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={`${a.name} versus ${b.name}, ${matchYear(match)}. ${match.classic?.headline ?? ''}`}
      onPress={() => router.push(`/match/${match.id}`)}
      style={({ pressed }) => [styles.cell, { opacity: pressed ? 0.85 : 1 }]}>
      <TicketStub
        vertical
        stock={stockFor(match, t)}
        height={HEIGHT}
        sideSize={STUB}
        main={
          <View style={{ flex: 1 }}>
            <View style={styles.yearRow}>
              <Txt variant="display" style={styles.year}>
                {matchYear(match)}
              </Txt>
              <Txt variant="label" tone="soft" numberOfLines={2} style={styles.fine}>
                {match.format} · {match.stage}
              </Txt>
            </View>
            <View style={styles.teams}>
              <Txt variant="display" numberOfLines={1} style={styles.code}>
                {a.short}
              </Txt>
              <Txt variant="bodySmall" style={styles.vs}>
                vs
              </Txt>
              <Txt variant="display" numberOfLines={1} style={[styles.code, { flexShrink: 1 }]}>
                {b.short}
              </Txt>
            </View>
            <Txt variant="bodySmall" numberOfLines={3} style={{ marginTop: 6 }}>
              {match.classic?.headline ?? match.result?.text ?? match.stage}
            </Txt>
          </View>
        }
        side={
          <View style={styles.stub}>
            <BallRating value={item.ratingAvg} size={11} gap={2} />
            <Txt variant="monoSmall">{item.ratingCount > 0 ? item.ratingAvg.toFixed(1) : '–'}</Txt>
          </View>
        }
      />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  cell: { width: CLASSIC_CARD_WIDTH, marginRight: 12 },
  yearRow: { flexDirection: 'row', alignItems: 'flex-end', justifyContent: 'space-between' },
  year: { fontSize: 40, lineHeight: 40 },
  fine: { fontSize: 9, marginLeft: 8, marginBottom: 5, flexShrink: 1 },
  teams: { flexDirection: 'row', alignItems: 'baseline', marginTop: 4 },
  code: { fontSize: 24, lineHeight: 26 },
  vs: { marginHorizontal: 6, fontFamily: fonts.label, fontSize: 12 },
  stub: { flex: 1, flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', alignSelf: 'stretch' },
});
