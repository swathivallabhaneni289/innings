import { router } from 'expo-router';
import { Pressable, View, type ViewStyle } from 'react-native';

import { Avatar, Txt } from '@/components';
import { useFriends, useMe, useUserProfile } from '@/data';
import type { User } from '@/domain/types';
import { useTheme } from '@/theme';

/**
 * Users we can name without a request: me, the people I follow, plus any passed in (e.g. a log's author).
 * Replies and "watched with" only carry user ids, so anyone else is looked up by `useUser`.
 */
export function useKnownUsers(extra: User[] = []) {
  const me = useMe();
  const friends = useFriends();
  const map = new Map<string, User>();
  friends.data?.friends.forEach((f) => map.set(f.user.id, f.user));
  if (me.data) map.set(me.data.id, me.data);
  extra.forEach((u) => map.set(u.id, u));
  return { map, ready: !friends.isLoading };
}

export type KnownUsers = ReturnType<typeof useKnownUsers>;

/** A user by id: from the known set, otherwise their profile (cached per user). */
export function useUser(id: string, known: KnownUsers): User | undefined {
  const k = known.map.get(id);
  const q = useUserProfile(!k && known.ready ? id : undefined);
  return k ?? q.data?.user;
}

/** Avatar + name for a user id, linking to their profile. */
export function UserChip({ id, known, style }: { id: string; known: KnownUsers; style?: ViewStyle }) {
  const t = useTheme();
  const user = useUser(id, known);
  return (
    <Pressable onPress={() => router.push('/user/' + id)} hitSlop={6} style={[{ flexDirection: 'row', alignItems: 'center' }, style]}>
      {user ? <Avatar user={user} size={22} /> : <View style={{ width: 22, height: 22, borderRadius: 11, backgroundColor: t.rule }} />}
      <Txt variant="bodySmall" style={{ marginLeft: 6 }} numberOfLines={1}>
        {user?.name ?? '…'}
      </Txt>
    </Pressable>
  );
}
