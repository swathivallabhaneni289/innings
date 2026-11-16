import * as Haptics from 'expo-haptics';
import { router } from 'expo-router';
import { useState } from 'react';
import { Alert, Pressable, ScrollView, StyleSheet, useWindowDimensions, View } from 'react-native';

import { Empty, MatchTicket, Panel, Txt } from '@/components';
import { MAX_TOP5 } from '@/domain/constants';
import type { Match } from '@/domain/types';
import { fonts, useTheme } from '@/theme';
import { useTop5 } from './top5';

const GUTTER = 20; // Screen's horizontal padding: the pager bleeds to the screen edges
const GAP = 12;

/** The fan's pinned greatest matches as large ticket stubs, one per page, numbered 1–5. */
export function Top5Pager({ top5, ratings }: { top5: Match[]; ratings: Record<string, number> }) {
  const t = useTheme();
  const { width } = useWindowDimensions();
  const pageW = width - GUTTER * 2;
  const [page, setPage] = useState(0);
  const pin = useTop5(top5.map((m) => m.id));

  if (top5.length === 0) {
    return (
      <Panel style={{ paddingVertical: 4 }}>
        <Empty
          title="No Top 5 yet"
          body="Tap “Pin to Top 5” on any match, or on the games in a collection below. Five tickets, your call."
        />
      </Panel>
    );
  }

  const confirmUnpin = (m: Match) => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    Alert.alert('Unpin from Top 5?', m.classic?.headline ?? m.stage, [
      { text: 'Keep it', style: 'cancel' },
      { text: 'Unpin', style: 'destructive', onPress: () => pin.unpin(m.id) },
    ]);
  };

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
          <View key={m.id} style={{ width: pageW, marginRight: i < top5.length - 1 ? GAP : 0 }}>
            <View style={styles.head}>
              <Txt variant="display" style={styles.num}>
                {i + 1}
              </Txt>
              <Txt variant="monoSmall" tone="faint" style={{ flex: 1, marginLeft: 6 }}>
                of {MAX_TOP5}
              </Txt>
              <Pressable
                onPress={() => pin.unpin(m.id)}
                hitSlop={14}
                disabled={pin.busyId === m.id}
                accessibilityRole="button"
                accessibilityLabel={`Unpin number ${i + 1}`}>
                <Txt variant="bodySmall" tone={pin.busyId === m.id ? 'faint' : 'soft'} style={{ fontFamily: fonts.label }}>
                  {pin.busyId === m.id ? 'Unpinning…' : 'Unpin'}
                </Txt>
              </Pressable>
            </View>
            <Pressable
              onPress={() => router.push(`/match/${m.id}`)}
              onLongPress={() => confirmUnpin(m)}
              accessibilityRole="button"
              accessibilityHint="Long-press to unpin">
              <MatchTicket match={m} size="lg" rating={ratings[m.id] ?? null} caption={m.classic?.headline} />
            </Pressable>
            {pin.message?.matchId === m.id ? (
              <Txt variant="bodySmall" tone="red" style={{ marginTop: 6 }}>
                {pin.message.text}
              </Txt>
            ) : null}
          </View>
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
  dots: { flexDirection: 'row', justifyContent: 'center', marginTop: 12 },
  dot: { width: 14, height: 3, marginHorizontal: 3 },
});
