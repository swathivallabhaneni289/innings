import { useState } from 'react';
import { View } from 'react-native';

import { Empty, Panel, QueryState, Screen, Txt } from '@/components';
import { useFriends } from '@/data';
import { FriendRow } from '@/features/profile/FriendRow';

export default function FriendsScreen() {
  const q = useFriends();
  const [pulling, setPulling] = useState(false);

  const refresh = async () => {
    setPulling(true);
    try {
      await q.refetch();
    } finally {
      setPulling(false);
    }
  };

  return (
    <Screen back title="Friends" onRefresh={refresh} refreshing={pulling}>
      <QueryState q={q} loadingLabel="Checking the dressing room…">
        {({ friends }) =>
          friends.length === 0 ? (
            <Panel>
              <Empty title="Empty pavilion" body="No fans to follow yet. Check back after the next match." />
            </Panel>
          ) : (
            <View>
              <Txt variant="body" tone="soft">
                Follow fans to see their stubs in your feed and race them on the Bragging Board.
              </Txt>
              <Txt variant="h3" style={{ marginTop: 28, marginBottom: 10 }}>
                Following {friends.filter((f) => f.following).length} of {friends.length}
              </Txt>
              <Panel style={{ paddingVertical: 2 }}>
                {friends.map((f, i) => (
                  <FriendRow key={f.user.id} user={f.user} following={f.following} last={i === friends.length - 1} />
                ))}
              </Panel>
            </View>
          )
        }
      </QueryState>
    </Screen>
  );
}
