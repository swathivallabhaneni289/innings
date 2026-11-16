import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { Empty, Txt } from '@/components';
import type { FanSummary } from '@/data';
import { FeedCard } from '@/features/feed/FeedCard';
import { fonts, useTheme } from '@/theme';

type Tab = 'popular' | 'friends';

/** AFTER: reviews in two tabs — the most-liked from every fan, and your circle's. */
export function Reviews({ fans }: { fans: FanSummary }) {
  const t = useTheme();
  const [tab, setTab] = useState<Tab>('popular');
  const list = tab === 'popular' ? fans.reviews : fans.friendReviews;
  const tabs: { id: Tab; label: string; count: number }[] = [
    { id: 'popular', label: 'Popular', count: fans.reviews.length },
    { id: 'friends', label: 'Friends', count: fans.friendReviews.length },
  ];
  return (
    <View>
      <View style={[styles.tabs, { borderBottomColor: t.rule }]}>
        {tabs.map((x) => {
          const active = x.id === tab;
          return (
            <Pressable
              key={x.id}
              onPress={() => setTab(x.id)}
              accessibilityRole="tab"
              accessibilityState={{ selected: active }}
              style={[styles.tab, { borderBottomColor: active ? t.ink : 'transparent' }]}>
              <Txt variant="body" tone={active ? 'ink' : 'soft'} style={{ fontFamily: active ? fonts.uiBold : fonts.ui }}>
                {x.label}
              </Txt>
              <Txt variant="monoSmall" tone={active ? 'soft' : 'faint'} style={{ marginLeft: 6 }}>
                {x.count}
              </Txt>
            </Pressable>
          );
        })}
      </View>
      {list.length === 0 ? (
        tab === 'popular' ? (
          <Empty title="No reviews yet" body="Nobody has had their say yet. Write the first one above." />
        ) : (
          <Empty title="Quiet in the dressing room" body="None of your friends have reviewed this one yet." />
        )
      ) : (
        list.map((item) => (
          <View key={item.log.id} style={{ marginBottom: 12 }}>
            <FeedCard item={item} showMatch={false} />
          </View>
        ))
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  tabs: { flexDirection: 'row', borderBottomWidth: StyleSheet.hairlineWidth, marginBottom: 14 },
  tab: { flexDirection: 'row', alignItems: 'center', minHeight: 44, marginRight: 22, borderBottomWidth: 2, marginBottom: -1 },
});
