import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Avatar, Button, TeamBadge, Txt } from '@/components';
import { getTeam, type UserResponse } from '@/data';
import { fonts, useTheme } from '@/theme';
import { FollowButton } from './FollowButton';

/** Photo, name, @handle, favourite team and fan bio; Friends + Edit on my own profile, Follow on anyone else's. */
export function ProfileHeader({ data }: { data: UserResponse }) {
  const t = useTheme();
  const { user, isMe, following, stats } = data;
  const team = getTeam(user.favouriteTeamId);
  const followingCount = Math.max(0, stats.rankOf - 1);

  return (
    <View>
      <View style={styles.top}>
        <Avatar user={user} size={72} />
        <View style={styles.idBlock}>
          <Txt variant="h1" numberOfLines={2}>
            {user.name}
          </Txt>
          <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
            @{user.handle}
          </Txt>
          <View style={styles.teamRow}>
            <TeamBadge teamId={user.favouriteTeamId} size="sm" />
            <Txt variant="bodySmall" tone="soft" numberOfLines={2} style={styles.teamName}>
              {team.name}
            </Txt>
          </View>
        </View>
      </View>

      {user.bio ? (
        <Txt variant="review" style={styles.bio}>
          “{user.bio}”
        </Txt>
      ) : isMe ? (
        <Pressable onPress={() => router.push('/settings')} hitSlop={8} accessibilityRole="button">
          <Txt variant="review" tone="faint" style={styles.bio}>
            Add a one-line fan bio…
          </Txt>
        </Pressable>
      ) : null}

      <View style={styles.actions}>
        {isMe ? (
          <>
            <Button
              label={`${followingCount} following`}
              kind="outline"
              size="sm"
              icon={<Feather name="users" size={14} color={t.ink} />}
              onPress={() => router.push('/friends')}
              style={styles.action}
            />
            <Button
              label="Edit profile"
              kind="outline"
              size="sm"
              icon={<Feather name="edit-2" size={13} color={t.ink} />}
              onPress={() => router.push('/settings')}
              style={styles.action}
            />
          </>
        ) : (
          <FollowButton userId={user.id} following={following} />
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  top: { flexDirection: 'row', alignItems: 'center', marginTop: 4 },
  idBlock: { flex: 1, marginLeft: 16 },
  teamRow: { flexDirection: 'row', alignItems: 'center', marginTop: 8 },
  teamName: { marginLeft: 8, flexShrink: 1, fontFamily: fonts.ui },
  bio: { marginTop: 16 },
  actions: { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', gap: 10, marginTop: 16 },
  action: { minHeight: 44 },
});
