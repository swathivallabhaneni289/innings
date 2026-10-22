import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { useMemo, useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { Panel, QueryState, Screen, SectionHeader, Txt } from '@/components';
import { useUserProfile, type UserResponse } from '@/data';
import { useTheme } from '@/theme';
import { firstName } from './constants';
import { PredictorRank } from './PredictorRank';
import { ProfileHeader } from './ProfileHeader';
import { StatsRow } from './StatsRow';
import { StubGrid } from './StubGrid';
import { Top5Strip } from './Top5Strip';
import { WatchBreakdown } from './WatchBreakdown';

/**
 * A fan's profile: header, Top 5 tickets, stats scoreboard, stub collection, where they watched,
 * predictor rank and badges. `id` is "me" for the Profile tab, or a user id for /user/[id].
 */
export function ProfileView({ id }: { id: string }) {
  const t = useTheme();
  const q = useUserProfile(id);
  const [pulling, setPulling] = useState(false);
  const isMe = q.data?.isMe ?? id === 'me';

  const refresh = async () => {
    setPulling(true);
    try {
      await q.refetch();
    } finally {
      setPulling(false);
    }
  };

  return (
    <Screen
      back={id !== 'me'}
      right={
        isMe ? (
          <Pressable onPress={() => router.push('/settings')} hitSlop={12} accessibilityRole="button" accessibilityLabel="Settings">
            <Feather name="settings" size={21} color={t.ink} />
          </Pressable>
        ) : undefined
      }
      onRefresh={refresh}
      refreshing={pulling}>
      <QueryState q={q} loadingLabel="Fetching the tickets…">
        {(data) => <ProfileBody data={data} />}
      </QueryState>
    </Screen>
  );
}

function ProfileBody({ data }: { data: UserResponse }) {
  const { user, isMe, stats, top5 } = data;
  const owner = firstName(user.name);
  const logs = useMemo(() => [...data.logs].sort((a, b) => Date.parse(b.log.createdAt) - Date.parse(a.log.createdAt)), [data.logs]);
  const ratings = useMemo(() => Object.fromEntries(data.logs.map((i) => [i.match.id, i.log.rating])), [data.logs]);

  return (
    <View>
      <ProfileHeader data={data} />

      <SectionHeader title="Top 5 matches" />
      <Top5Strip top5={top5} ratings={ratings} isMe={isMe} ownerFirstName={owner} />

      <StatsRow stats={stats} />

      <SectionHeader title={logs.length ? `Stub collection · ${logs.length}` : 'Stub collection'} />
      <StubGrid logs={logs} isMe={isMe} ownerFirstName={owner} />

      <SectionHeader title={isMe ? "Where I've watched" : `Where ${owner} has watched`} />
      <WatchBreakdown stats={stats} isMe={isMe} />

      <SectionHeader title="Predictor rank" />
      <PredictorRank stats={stats} isMe={isMe} ownerFirstName={owner} />

      {isMe ? (
        <Panel style={styles.wrapped}>
          <Txt variant="h3">Season Wrapped</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
            Your season on one ticket: every match, every rating, every call. It lands when the season ends.
          </Txt>
        </Panel>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrapped: { marginTop: 28 },
});
