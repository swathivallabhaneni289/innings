import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Avatar, Button, TeamBadge, Txt } from '@/components';
import { getTeam, type UserResponse } from '@/data';
import { fonts, useTheme } from '@/theme';
import { listLine } from '@/features/onboarding/teams';
import { FollowButton } from './FollowButton';

/** Photo, name, @handle, the teams they back and fan bio; Friends + Edit on my own profile, Follow on anyone else's. */
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
          <View style={styles.teamRow} accessible accessibilityLabel={`Backs ${listLine(user.teamIds.map((id) => getTeam(id).name))}`}>
            {user.teamIds.map((id) => (
              <View key={id} style={styles.teamBadge}>
                <TeamBadge teamId={id} size="sm" />
              </View>
            ))}
            {user.teamIds.length === 1 ? (
              <Txt variant="bodySmall" tone="soft" numberOfLines={2} style={styles.teamName}>
                {team.name}
              </Txt>
            ) : null}
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
  teamRow: { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', gap: 10, marginTop: 8 },
  teamBadge: { paddingTop: 2 },
  teamName: { flexShrink: 1, fontFamily: fonts.ui },
  bio: { marginTop: 16 },
  actions: { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', gap: 10, marginTop: 16 },
  action: { minHeight: 44 },
});
