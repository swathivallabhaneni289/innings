import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { View } from 'react-native';

import { Button, Panel, Txt } from '@/components';
import type { Match } from '@/domain/types';
import { useTheme } from '@/theme';

/** The way into a match's chat room, worded for where the match is. Vault classics have no chat. */
export function ChatEntry({ match }: { match: Match }) {
  const t = useTheme();
  if (match.kind === 'classic') return null;
  const live = match.state === 'live';
  const line = live
    ? 'Talk it through with other fans while it’s on.'
    : match.state === 'upcoming'
      ? 'Talk through the build-up with other fans. It opens a day before the first ball.'
      : 'Have your say on how it went. It stays open for a day after the match.';
  return (
    <Panel style={{ marginTop: 16 }}>
      <Txt variant="h3">Match chat</Txt>
      <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2, marginBottom: 12 }}>
        {line}
      </Txt>
      <View>
        <Button
          label={live ? 'Join the chat' : 'Open the chat'}
          kind={live ? 'primary' : 'outline'}
          icon={<Feather name="message-circle" size={17} color={live ? t.onInk : t.ink} />}
          onPress={() => router.push(`/match/chat/${match.id}`)}
          style={{ minHeight: 48 }}
        />
      </View>
    </Panel>
  );
}
