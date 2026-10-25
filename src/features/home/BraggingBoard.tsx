import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Avatar, Button, Empty, Panel, Txt } from '@/components';
import type { BoardRow } from '@/domain/api';
import { fonts, useTheme } from '@/theme';

const SHOW = 5;

/** This week's top predictors among the people I follow. "You" is always on the board. */
export function BraggingBoard({ rows }: { rows: BoardRow[] }) {
  const t = useTheme();
  // A new fan's board is just "You · 0": a table of one reads like a fault, so say what fills it instead.
  if (!rows.some((r) => !r.isMe)) {
    return (
      <Empty
        title="Just you on the board"
        body="Follow a few fans and race them for bragging points. Make a call before the first ball to get on the scoresheet."
        action={<Button label="Find friends" kind="outline" size="sm" onPress={() => router.push('/friends')} style={{ minHeight: 44 }} />}
      />
    );
  }
  if (rows.every((r) => r.points === 0)) {
    return <Empty title="No points on the board" body="Make a call before the first ball. Points land here once matches finish." />;
  }
  // Ties share a rank.
  const ranked = rows.map((r) => ({ ...r, rank: 1 + rows.filter((o) => o.points > r.points).length }));
  const shown = ranked.slice(0, SHOW);
  const me = ranked.find((r) => r.isMe);
  if (me && !shown.includes(me)) shown.push(me);
  return (
    <Panel style={{ paddingVertical: 6, paddingHorizontal: 0 }}>
      <View style={styles.head}>
        <Txt variant="bodySmall" tone="soft">
          Points from calls, last 7 days
        </Txt>
      </View>
      {shown.map((r, i) => (
        <Pressable
          key={r.user.id}
          onPress={() => router.push('/user/' + r.user.id)}
          accessibilityRole="button"
          accessibilityLabel={`${r.isMe ? 'You' : r.user.name}, rank ${r.rank}, ${r.points} points`}
          style={({ pressed }) => [
            styles.row,
            { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule },
            r.isMe ? { backgroundColor: t.paper } : null,
            pressed ? { opacity: 0.8 } : null,
          ]}>
          <Txt variant="mono" tone={r.rank === 1 ? 'ink' : 'faint'} style={{ width: 30, fontFamily: r.rank === 1 ? fonts.monoBold : fonts.monoMedium }}>
            {String(r.rank).padStart(2, '0')}
          </Txt>
          <Avatar user={r.user} size={28} />
          <Txt variant="body" numberOfLines={1} style={{ flex: 1, marginLeft: 10, fontFamily: r.isMe || r.rank === 1 ? fonts.uiBold : fonts.ui }}>
            {r.isMe ? 'You' : r.user.name}
          </Txt>
          <Txt variant="score" style={{ fontSize: 24, lineHeight: 28, minWidth: 40, textAlign: 'right' }}>
            {r.points}
          </Txt>
          <Txt variant="label" tone="faint" style={{ fontSize: 9, marginLeft: 4, width: 22 }}>
            pts
          </Txt>
        </Pressable>
      ))}
    </Panel>
  );
}

const styles = StyleSheet.create({
  head: { paddingHorizontal: 14, paddingTop: 6, paddingBottom: 8 },
  row: { flexDirection: 'row', alignItems: 'center', minHeight: 48, paddingHorizontal: 14 },
});
