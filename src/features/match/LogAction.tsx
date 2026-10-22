import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { Pressable, View } from 'react-native';

import { Button, MatchTicket } from '@/components';
import type { Match, MatchLog } from '@/domain/types';
import { useTheme } from '@/theme';

/** AFTER: "Log this match", or — once logged — your ticket stub with an "Edit your log" button. */
export function LogAction({ match, log }: { match: Match; log: MatchLog | null }) {
  const t = useTheme();
  const openLog = () => router.push({ pathname: '/log', params: { matchId: match.id } });
  if (log) {
    return (
      <View>
        <Pressable onPress={() => router.push(`/stub/${log.id}`)} accessibilityRole="button" accessibilityLabel="Open your ticket stub">
          <MatchTicket match={match} rating={log.rating} stubNo={log.stubNo} instantClassic={log.instantClassic} />
        </Pressable>
        <Button
          label="Edit your log"
          kind="outline"
          icon={<Feather name="edit-3" size={15} color={t.ink} />}
          onPress={openLog}
          style={{ marginTop: 12, minHeight: 50 }}
        />
      </View>
    );
  }
  return <Button label="Log this match" icon={<Feather name="plus" size={17} color={t.onInk} />} onPress={openLog} style={{ minHeight: 50 }} />;
}
