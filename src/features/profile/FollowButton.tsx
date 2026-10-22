import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useEffect, useState } from 'react';
import { View, type ViewStyle } from 'react-native';

import { Button, Txt } from '@/components';
import { useActions } from '@/data';
import { useTheme } from '@/theme';

/**
 * Follow / Following toggle. Keeps its own busy state so only the tapped button spins
 * (useActions().pending is shared by every action), and shows a failed toggle inline.
 */
export function FollowButton({
  userId,
  following,
  size = 'md',
  style,
}: {
  userId: string;
  following: boolean;
  size?: 'sm' | 'md';
  style?: ViewStyle;
}) {
  const t = useTheme();
  const { toggleFollow } = useActions();
  const [on, setOn] = useState(following);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => setOn(following), [following]);

  const press = async () => {
    Haptics.selectionAsync();
    setBusy(true);
    setError(null);
    try {
      const res = await toggleFollow(userId);
      setOn(res.following);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't update. Try again.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <View style={style}>
      <Button
        label={on ? 'Following' : 'Follow'}
        kind={on ? 'outline' : 'primary'}
        size={size}
        style={{ minHeight: 44 }}
        loading={busy}
        onPress={press}
        icon={busy ? undefined : <Feather name={on ? 'check' : 'user-plus'} size={size === 'sm' ? 12 : 14} color={on ? t.ink : t.onInk} />}
      />
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 4, maxWidth: 220 }}>
          {error}
        </Txt>
      ) : null}
    </View>
  );
}
