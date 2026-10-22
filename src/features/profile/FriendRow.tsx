import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Avatar, TeamBadge, Txt } from '@/components';
import { getTeam } from '@/data';
import type { User } from '@/domain/types';
import { useTheme } from '@/theme';
import { FollowButton } from './FollowButton';

/**
 * One fan in the friends list: avatar, name, bio, favourite team and a Follow toggle. Tap → their profile.
 * Rows sit in a white card; `last` drops the hairline under the final row.
 */
export function FriendRow({ user, following, last = false }: { user: User; following: boolean; last?: boolean }) {
  const t = useTheme();
  return (
    <Pressable
      onPress={() => router.push('/user/' + user.id)}
      accessibilityRole="button"
      accessibilityLabel={`${user.name}'s profile. Backs ${getTeam(user.favouriteTeamId).name}`}
      style={({ pressed }) => [
        styles.row,
        { borderBottomColor: t.rule, borderBottomWidth: last ? 0 : StyleSheet.hairlineWidth, opacity: pressed ? 0.7 : 1 },
      ]}>
      <Avatar user={user} size={44} />
      <View style={styles.body}>
        <Txt variant="h3" numberOfLines={2}>
          {user.name}
        </Txt>
        <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
          @{user.handle}
        </Txt>
        {user.bio ? (
          <Txt variant="bodySmall" numberOfLines={3} style={styles.bio}>
            “{user.bio}”
          </Txt>
        ) : null}
        {/* The code with its team-colour bar is enough here; full names ("Royal Challengers…") clipped beside the button. */}
        <View style={styles.team}>
          <TeamBadge teamId={user.favouriteTeamId} size="sm" />
        </View>
      </View>
      <FollowButton userId={user.id} following={following} size="sm" style={styles.follow} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'flex-start', paddingVertical: 14 },
  body: { flex: 1, marginLeft: 12, marginRight: 10 },
  bio: { marginTop: 4 },
  team: { flexDirection: 'row', alignItems: 'center', marginTop: 8 },
  follow: { alignItems: 'flex-end', maxWidth: 130 },
});
