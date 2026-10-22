import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Avatar, Empty, Panel, Txt } from '@/components';
import { getTeam, playerName, type PredictionResultRow } from '@/data';
import { fonts, useTheme } from '@/theme';

const surname = (id: string) => {
  const parts = playerName(id).split(/\s+/);
  return parts.length > 1 ? parts.slice(1).join(' ') : parts[0];
};

const multiplier = (m: number) => `×${Number.isInteger(m) ? m : m.toFixed(1)}`;

/** AFTER: who called it — three picks marked ✓/✗ (green only for a correct pick), points and the streak multiplier. */
export function PredictionResults({ results, meId }: { results: PredictionResultRow[]; meId?: string }) {
  if (results.length === 0) {
    return <Empty title="No calls on the books" body="Nobody in your circle predicted this one." />;
  }
  return (
    <Panel style={{ paddingVertical: 4 }}>
      {results.map((r, i) => (
        <ResultRow key={r.user.id} row={r} isMe={r.user.id === meId} last={i === results.length - 1} />
      ))}
    </Panel>
  );
}

function ResultRow({ row, isMe, last }: { row: PredictionResultRow; isMe: boolean; last: boolean }) {
  const t = useTheme();
  const { user, prediction: p, outcome: o } = row;
  return (
    <Pressable
      onPress={() => router.push(`/user/${user.id}`)}
      accessibilityRole="button"
      style={({ pressed }) => [styles.row, { borderBottomColor: last ? 'transparent' : t.rule, opacity: pressed ? 0.7 : 1 }]}>
      <Avatar user={user} size={34} />
      <View style={styles.body}>
        <View style={styles.nameRow}>
          <Txt variant="h3" numberOfLines={1} style={{ flexShrink: 1, fontSize: 15 }}>
            {isMe ? 'You' : user.name}
          </Txt>
          {o.upset ? (
            <Txt variant="label" tone="soft" style={styles.tag}>
              Upset +10
            </Txt>
          ) : null}
        </View>
        <View style={styles.picks}>
          <Pick label={getTeam(p.winnerId).short} correct={o.winnerCorrect} settled={o.settled} />
          <Pick label={surname(p.topRunScorerId)} correct={o.topScorerCorrect} settled={o.settled} />
          <Pick label={surname(p.topWicketTakerId)} correct={o.topWicketCorrect} settled={o.settled} />
        </View>
      </View>
      <View style={styles.points}>
        {o.settled ? (
          <Txt variant="mono" tone={o.points > 0 ? 'ink' : 'faint'} style={{ fontSize: 16, lineHeight: 21 }}>
            {o.points > 0 ? `+${o.points}` : '0'}
          </Txt>
        ) : (
          <Txt variant="bodySmall" tone="faint">
            Pending
          </Txt>
        )}
        {o.multiplier > 1 ? (
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 1 }}>
            <Txt variant="monoSmall" tone="soft">
              {multiplier(o.multiplier)}
            </Txt>{' '}
            streak
          </Txt>
        ) : null}
      </View>
    </Pressable>
  );
}

function Pick({ label, correct, settled }: { label: string; correct: boolean; settled: boolean }) {
  const right = settled && correct;
  return (
    <View style={styles.pick}>
      <Txt
        variant="bodySmall"
        tone={right ? 'green' : settled ? 'faint' : 'soft'}
        numberOfLines={1}
        style={{ flexShrink: 1, fontFamily: right ? fonts.uiBold : fonts.serif }}>
        {label}
      </Txt>
      {settled ? (
        <Txt variant="bodySmall" tone={correct ? 'green' : 'faint'} style={{ marginLeft: 3 }} accessibilityLabel={correct ? 'correct' : 'wrong'}>
          {correct ? '✓' : '✗'}
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 11, borderBottomWidth: StyleSheet.hairlineWidth },
  body: { flex: 1, marginLeft: 12, marginRight: 8 },
  nameRow: { flexDirection: 'row', alignItems: 'center' },
  tag: { fontSize: 9.5, marginLeft: 8 },
  picks: { flexDirection: 'row', marginTop: 4 },
  pick: { flexDirection: 'row', alignItems: 'center', marginRight: 12, maxWidth: '36%' },
  points: { alignItems: 'flex-end', minWidth: 52 },
});
