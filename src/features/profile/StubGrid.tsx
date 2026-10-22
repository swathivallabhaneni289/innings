import { router } from 'expo-router';
import { Pressable, StyleSheet, useWindowDimensions, View } from 'react-native';

import { Button, Empty, MatchTicket, Panel } from '@/components';
import { matchTitle, type FeedItem } from '@/data';

const GUTTER = 20;
const GAP = 12;

/** Every logged match as a small vertical ticket stub, two to a row → the stub page. */
export function StubGrid({ logs, isMe, ownerFirstName }: { logs: FeedItem[]; isMe: boolean; ownerFirstName: string }) {
  const { width } = useWindowDimensions();
  const cellW = Math.floor((width - GUTTER * 2 - GAP) / 2);

  if (logs.length === 0) {
    return (
      <Panel style={{ paddingVertical: 4 }}>
        {isMe ? (
          <Empty
            title="No stubs yet"
            body="Log a match and its ticket gets printed here."
            action={<Button label="Log a match" size="sm" onPress={() => router.push('/log')} style={{ minHeight: 44 }} />}
          />
        ) : (
          <Empty title="No stubs yet" body={`${ownerFirstName} hasn't logged a match yet. Quiet season.`} />
        )}
      </Panel>
    );
  }

  return (
    <View style={styles.grid}>
      {logs.map((item) => (
        <Pressable
          key={item.log.id}
          onPress={() => router.push('/stub/' + item.log.id)}
          accessibilityRole="button"
          accessibilityLabel={`${matchTitle(item.match)}, rated ${item.log.rating} balls`}
          style={({ pressed }) => ({ width: cellW, marginBottom: GAP + 2, opacity: pressed ? 0.85 : 1 })}>
          <MatchTicket match={item.match} size="sm" rating={item.log.rating} stubNo={item.log.stubNo} instantClassic={item.log.instantClassic} />
        </Pressable>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  grid: { flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between' },
});
