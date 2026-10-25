import { StyleSheet, View } from 'react-native';

import { Avatar, TeamBadge, Txt } from '@/components';
import type { User } from '@/domain/types';
import { FollowButton } from '@/features/profile/FollowButton';
import { useTheme } from '@/theme';

/**
 * One fan to follow during sign-up: avatar, name, @handle, bio and team code, with a Follow toggle.
 * Unlike FriendRow it doesn't open the profile (the app's screens open once sign-up is done).
 */
export function FanRow({ user, following, last = false }: { user: User; following: boolean; last?: boolean }) {
  const t = useTheme();
  return (
    <View style={[styles.row, { borderBottomColor: t.rule, borderBottomWidth: last ? 0 : StyleSheet.hairlineWidth }]}>
      <Avatar user={user} size={44} />
      <View style={styles.body}>
        <Txt variant="h3" numberOfLines={2}>
          {user.name}
        </Txt>
        <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
          @{user.handle}
        </Txt>
        {user.bio ? (
          <Txt variant="bodySmall" numberOfLines={2} style={{ marginTop: 4 }}>
            “{user.bio}”
          </Txt>
        ) : null}
        {user.favouriteTeamId ? (
          <View style={{ flexDirection: 'row', marginTop: 8 }}>
            <TeamBadge teamId={user.favouriteTeamId} size="sm" />
          </View>
        ) : null}
      </View>
      <FollowButton userId={user.id} following={following} size="sm" style={styles.follow} />
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'flex-start', paddingVertical: 14 },
  body: { flex: 1, marginLeft: 12, marginRight: 10 },
  follow: { alignItems: 'flex-end', maxWidth: 130 },
});
