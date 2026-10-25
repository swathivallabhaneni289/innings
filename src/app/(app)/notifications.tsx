import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { useCallback, useEffect, useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import Svg, { Line, Path } from 'react-native-svg';

import { Empty, QueryState, Screen, Txt } from '@/components';
import { useNotifications, useNow, usePrefs, type NotificationsResponse } from '@/data';
import { relative } from '@/lib/format';
import { useTheme } from '@/theme';

type Item = NotificationsResponse['items'][number];

const ICONS: Record<string, keyof typeof Feather.glyphMap> = {
  lock: 'clock',
  pts: 'award',
  reply: 'message-square',
  like: 'heart',
};

export default function NotificationsScreen() {
  const t = useTheme();
  const q = useNotifications();
  const now = useNow(60_000);
  const markSeen = usePrefs((s) => s.markNotificationsSeen);
  useEffect(() => {
    if (q.data) markSeen();
  }, [q.data, markSeen]);
  const { refetch } = q;
  const [pulling, setPulling] = useState(false);
  const onRefresh = useCallback(async () => {
    setPulling(true);
    try {
      await refetch();
    } finally {
      setPulling(false);
    }
  }, [refetch]);

  return (
    <Screen back title="Notifications" onRefresh={onRefresh} refreshing={pulling}>
      <QueryState q={q}>
        {(d) =>
          d.items.length === 0 ? (
            <Empty title="No new deliveries" body="When friends log a match, or reply to and like your stubs, you'll hear it here first." />
          ) : (
            <View style={[styles.card, { backgroundColor: t.card }]}>
              {d.items.map((n, i) => (
                <NotificationRow key={n.id} item={n} now={now} first={i === 0} />
              ))}
            </View>
          )
        }
      </QueryState>
    </Screen>
  );
}

/** A small ticket outline: a friend logged a match. */
function TicketGlyph({ color }: { color: string }) {
  return (
    <Svg width={20} height={20} viewBox="0 0 24 24" fill="none">
      <Path d="M3 6h18v4a2 2 0 0 0 0 4v4H3v-4a2 2 0 0 0 0-4z" stroke={color} strokeWidth={1.8} strokeLinejoin="round" />
      <Line x1={15.5} y1={8} x2={15.5} y2={16} stroke={color} strokeWidth={1.6} strokeDasharray="1.2 2.4" strokeLinecap="round" />
    </Svg>
  );
}

function NotificationRow({ item, now, first }: { item: Item; now: number; first: boolean }) {
  const t = useTheme();
  const kind = item.id.split(':')[0];
  const color = kind === 'pts' ? t.green : kind === 'lock' ? t.red : t.ink;
  const target = item.logId ? '/stub/' + item.logId : item.matchId ? '/match/' + item.matchId : null;
  return (
    <Pressable
      disabled={!target}
      onPress={() => target && router.push(target)}
      accessibilityRole="button"
      style={({ pressed }) => [styles.row, !first ? { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule } : null, { opacity: pressed ? 0.7 : 1 }]}>
      <View style={[styles.icon, { backgroundColor: kind === 'pts' ? t.greenSoft : t.paper }]}>
        {kind === 'log' ? <TicketGlyph color={color} /> : <Feather name={ICONS[kind] ?? 'bell'} size={16} color={color} />}
      </View>
      <View style={{ flex: 1, marginLeft: 12 }}>
        <Txt variant="body" style={{ fontSize: 14.5, lineHeight: 20 }}>
          {item.text}
        </Txt>
        <Txt variant="bodySmall" tone="faint" style={{ marginTop: 2 }}>
          {relative(item.createdAt, now)}
        </Txt>
      </View>
      {target ? <Feather name="chevron-right" size={16} color={t.inkFaint} style={{ marginLeft: 8 }} /> : null}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: { paddingHorizontal: 14, marginTop: 8 },
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 13, minHeight: 56 },
  icon: { width: 34, height: 34, alignItems: 'center', justifyContent: 'center' },
});
