import { router } from 'expo-router';
import { useState } from 'react';
import { Pressable, ScrollView, StyleSheet, useWindowDimensions, View } from 'react-native';

import { Button, Empty, MatchTicket, Panel, Txt } from '@/components';
import { matchTitle } from '@/data';
import type { Match } from '@/domain/types';
import { fonts, useTheme } from '@/theme';

const GUTTER = 20; // Screen's horizontal padding: the strip bleeds to the screen edges
const GAP = 12;
const PEEK = 28; // how much of the next ticket shows, so it reads as a strip

/**
 * Read-only Top 5 (pinning lives in the Vault): large numbered ticket stubs in a snapping horizontal strip.
 * `ratings` holds the profile owner's own rating per match id, when they logged it.
 */
export function Top5Strip({
  top5,
  ratings,
  isMe,
  ownerFirstName,
}: {
  top5: Match[];
  ratings: Record<string, number>;
  isMe: boolean;
  ownerFirstName: string;
}) {
  const t = useTheme();
  const { width } = useWindowDimensions();
  const [page, setPage] = useState(0);
  const pageW = width - GUTTER * 2 - (top5.length > 1 ? PEEK : 0);

  if (top5.length === 0) {
    return (
      <Panel style={{ paddingVertical: 4 }}>
        {isMe ? (
          <Empty
            title="No Top 5 yet"
            body="The five matches you'd rewatch on any rain-delayed day. Pin them in the Vault."
            action={<Button label="Open the Vault" kind="outline" size="sm" onPress={() => router.navigate('/vault')} style={{ minHeight: 44 }} />}
          />
        ) : (
          <Empty title="No Top 5 pinned" body={`${ownerFirstName} hasn't picked their all-time greats yet.`} />
        )}
      </Panel>
    );
  }

  return (
    <View>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        snapToInterval={pageW + GAP}
        decelerationRate="fast"
        disableIntervalMomentum
        style={{ marginHorizontal: -GUTTER }}
        contentContainerStyle={{ paddingHorizontal: GUTTER }}
        onMomentumScrollEnd={(e) => setPage(Math.round(e.nativeEvent.contentOffset.x / (pageW + GAP)))}>
        {top5.map((m, i) => (
          <Pressable
            key={m.id}
            onPress={() => router.push('/match/' + m.id)}
            accessibilityRole="button"
            accessibilityLabel={`Number ${i + 1}: ${matchTitle(m)}`}
            style={{ width: pageW, marginRight: i < top5.length - 1 ? GAP : 0 }}>
            <View style={styles.head}>
              <Txt variant="display" style={styles.num}>
                {i + 1}
              </Txt>
              <Txt variant="bodySmall" tone="soft" numberOfLines={1} style={styles.headLabel}>
                {m.classic ? m.classic.competitionName : `${matchTitle(m)} · ${m.stage}`}
              </Txt>
            </View>
            <MatchTicket match={m} size="lg" rating={ratings[m.id] ?? null} caption={m.classic?.headline} />
          </Pressable>
        ))}
      </ScrollView>
      {top5.length > 1 ? (
        <View style={styles.dots}>
          {top5.map((m, i) => (
            <View key={m.id} style={[styles.dot, { backgroundColor: i === Math.min(page, top5.length - 1) ? t.ink : t.rule }]} />
          ))}
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  head: { flexDirection: 'row', alignItems: 'baseline', marginBottom: 6 },
  num: { fontSize: 34, lineHeight: 36 },
  headLabel: { flex: 1, marginLeft: 10, fontFamily: fonts.ui },
  dots: { flexDirection: 'row', justifyContent: 'center', marginTop: 12 },
  dot: { width: 14, height: 3, marginHorizontal: 3 },
});
