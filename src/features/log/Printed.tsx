import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useEffect, useState } from 'react';
import { StyleSheet, View } from 'react-native';
import Animated, { Easing, useAnimatedStyle, useSharedValue, withTiming } from 'react-native-reanimated';

import { Button, MatchTicket, Panel, Rule, SpinningBall, Txt } from '@/components';
import { playerName, type FeedItem } from '@/data';
import type { useShareStory } from '@/features/share/useShareStory';
import { stubNo as fmtStubNo } from '@/lib/format';
import { useTheme } from '@/theme';
import { placeLine } from './draft';
import { RatingLine } from './RatingLine';

const PRINT_MS = 520;
const STAMP_AT = PRINT_MS + 180;

/**
 * After Save: the ticket feeds out of the printer, the LOGGED stamp slams down, then Share story / Done.
 * The story card itself is mounted off-screen by the flow (outside any scroll view) and captured via `story`.
 */
export function Printed({
  item,
  relogged,
  story,
  onDone,
}: {
  item: FeedItem;
  relogged: boolean;
  story: ReturnType<typeof useShareStory>;
  onDone: () => void;
}) {
  const t = useTheme();
  const { log, match } = item;
  const [stamped, setStamped] = useState(false);
  const feed = useSharedValue(0);

  useEffect(() => {
    feed.value = withTiming(1, { duration: PRINT_MS, easing: Easing.out(Easing.cubic) });
    const stamp = setTimeout(() => setStamped(true), STAMP_AT);
    const buzz = setTimeout(() => Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success), STAMP_AT + 170);
    return () => {
      clearTimeout(stamp);
      clearTimeout(buzz);
    };
  }, [feed]);

  const printing = useAnimatedStyle(() => ({ opacity: feed.value, transform: [{ translateY: (1 - feed.value) * -36 }] }));
  const review = log.review?.trim();
  const rated = Object.keys(log.playerRatings).length;
  const hasFacts = !!log.fanPotmId || rated > 0 || !!review;

  return (
    <View>
      <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
        <Txt variant="mono" tone="red" style={{ fontSize: 12.5 }}>
          {fmtStubNo(log.stubNo)}
        </Txt>
        {relogged ? ' reprinted with your new take.' : ' is in your collection.'}
      </Txt>

      <View style={[styles.slot, { backgroundColor: t.ink }]} />
      <Animated.View style={[{ marginTop: 6 }, printing]}>
        <MatchTicket
          match={match}
          size="lg"
          rating={log.rating}
          stubNo={log.stubNo}
          instantClassic={log.instantClassic}
          caption={placeLine(log)}
          stamped={stamped}
        />
      </Animated.View>

      <View style={styles.ratingRow}>
        <SpinningBall spinKey={log.id} size={20} />
        <RatingLine rating={log.rating} style={{ marginLeft: 10 }} />
      </View>

      {hasFacts ? (
        <Panel style={{ marginTop: 16 }}>
          {log.fanPotmId ? (
            <View style={styles.factRow}>
              <Feather name="award" size={16} color={t.red} />
              <Txt variant="bodySmall" tone="soft" style={{ marginLeft: 10, flex: 1 }}>
                Your player of the match: <Txt variant="h3">{playerName(log.fanPotmId)}</Txt>
              </Txt>
            </View>
          ) : null}
          {rated > 0 ? (
            <View style={[styles.factRow, log.fanPotmId ? { marginTop: 10 } : null]}>
              <Feather name="users" size={16} color={t.inkSoft} />
              <Txt variant="bodySmall" tone="soft" style={{ marginLeft: 10 }}>
                {rated} player {rated === 1 ? 'rating' : 'ratings'} filed
              </Txt>
            </View>
          ) : null}
          {review ? (
            <>
              {log.fanPotmId || rated > 0 ? <Rule style={{ marginVertical: 12 }} /> : null}
              <Txt variant="review" numberOfLines={3}>
                “{review}”
              </Txt>
            </>
          ) : null}
        </Panel>
      ) : null}

      <View style={styles.actions}>
        <Button
          label="Share story"
          size="lg"
          icon={<Feather name="share" size={16} color={t.onInk} />}
          loading={story.sharing}
          onPress={story.share}
          style={{ flex: 1 }}
        />
        <Button label="Done" kind="outline" size="lg" onPress={onDone} style={{ flex: 1, marginLeft: 10 }} />
      </View>
      {story.error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
          {story.error}
        </Txt>
      ) : (
        <Txt variant="bodySmall" tone="faint" style={{ marginTop: 8 }}>
          A 9:16 card of your ticket, ready for your story.
        </Txt>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  slot: { height: 3, marginTop: 14, marginHorizontal: -4 },
  ratingRow: { flexDirection: 'row', alignItems: 'center', marginTop: 18 },
  factRow: { flexDirection: 'row', alignItems: 'center' },
  actions: { flexDirection: 'row', marginTop: 24 },
});
