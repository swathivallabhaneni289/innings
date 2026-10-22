import { router } from 'expo-router';
import { useCallback, useState } from 'react';
import { View } from 'react-native';

import { Button, Empty, QueryState, Screen, SectionHeader } from '@/components';
import { useHome, useUnreadCount } from '@/data';
import type { FeedItem } from '@/domain/api';
import { FeedCard } from '@/features/feed/FeedCard';
import { BraggingBoard } from '@/features/home/BraggingBoard';
import { HomeHeader } from '@/features/home/HomeHeader';
import { JustFinished } from '@/features/home/JustFinished';
import { OnNowStrip } from '@/features/home/OnNowStrip';
import { YourCallCard } from '@/features/home/YourCallCard';

const FEED_PAGE = 8;

export default function HomeScreen() {
  const unread = useUnreadCount();
  const home = useHome();
  const { refetch } = home;
  // Own flag so the spinner shows for a pull only, not for the 30 s background refresh.
  const [pulling, setPulling] = useState(false);
  const onRefresh = useCallback(async () => {
    setPulling(true);
    try {
      await refetch();
    } finally {
      setPulling(false);
    }
  }, [refetch]);

  return (
    <Screen header={<HomeHeader unread={unread} />} onRefresh={onRefresh} refreshing={pulling}>
      <QueryState q={home} loadingLabel="Opening the turnstiles…">
        {(d) => (
          <>
            <SectionHeader title="On now" style={{ marginTop: 16 }} />
            <OnNowStrip matches={d.onNow} />

            {d.yourCall ? (
              <>
                <SectionHeader title="Your call" />
                <YourCallCard match={d.yourCall} split={d.yourCallSplit} mine={d.yourCallMine} />
              </>
            ) : null}

            <SectionHeader title="Friends' stubs" action="Friends" onAction={() => router.push('/friends')} />
            <FriendsFeed items={d.feed} />

            <SectionHeader title="Just finished" />
            <JustFinished items={d.justFinished} />

            <SectionHeader title="Bragging board" />
            <BraggingBoard rows={d.board} />
          </>
        )}
      </QueryState>
    </Screen>
  );
}

function FriendsFeed({ items }: { items: FeedItem[] }) {
  const [all, setAll] = useState(false);
  if (items.length === 0) {
    return (
      <Empty
        title="Quiet in the stands"
        body="When the friends you follow log a match, their stubs land here."
        action={<Button label="Find friends" kind="outline" size="sm" onPress={() => router.push('/friends')} />}
      />
    );
  }
  const shown = all ? items : items.slice(0, FEED_PAGE);
  return (
    <View>
      {shown.map((it, i) => (
        <View key={it.log.id} style={i > 0 ? { marginTop: 12 } : null}>
          <FeedCard item={it} />
        </View>
      ))}
      {!all && items.length > FEED_PAGE ? (
        <Button label={`More from the stands · ${items.length - FEED_PAGE}`} kind="outline" size="sm" onPress={() => setAll(true)} style={{ marginTop: 14, alignSelf: 'center' }} />
      ) : null}
    </View>
  );
}
