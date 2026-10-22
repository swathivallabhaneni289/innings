import { StyleSheet, View } from 'react-native';

import { CricketBall, Empty, Panel, Txt } from '@/components';
import { getPlayer, getTeam, type FanSummary, type PlayerAverage } from '@/data';
import type { Match } from '@/domain/types';
import { fonts, useTheme } from '@/theme';

/**
 * AFTER: Fans' Player of the Match (highest average with at least 3 ratings) beside the official pick,
 * a red "Fans disagree" rubber stamp when they differ, then every rated player with an average out of 10.
 */
export function PlayerRatings({ match, fans }: { match: Match; fans: FanSummary }) {
  const players = fans.players.filter((p) => getPlayer(p.playerId)); // classics may rate players outside the catalog
  const fansPotm = fans.fansPotmId ? players.find((p) => p.playerId === fans.fansPotmId) : undefined;
  const officialId = fans.officialPotmId ?? match.result?.officialPotmId ?? null;
  const official = getPlayer(officialId);
  const officialName = official?.name ?? match.classic?.potmName ?? null;
  const officialAvg = officialId ? players.find((p) => p.playerId === officialId)?.avg : undefined;

  if (players.length === 0) {
    // Nothing rated (often a Vault classic): keep the official pick, skip the empty crown and table.
    return (
      <View>
        {officialName ? (
          <Panel style={{ marginBottom: 12 }}>
            <Txt variant="label" tone="soft" style={styles.small}>
              Official player of the match
            </Txt>
            <Txt variant="h3" style={{ marginTop: 4, fontSize: 16 }}>
              {officialName}
            </Txt>
          </Panel>
        ) : null}
        {match.kind === 'classic' ? (
          <Txt variant="bodySmall" tone="soft">
            No fan player ratings for this one.
          </Txt>
        ) : (
          <Empty title="No player ratings yet" body="Rate the XIs when you log it and the table fills up here." />
        )}
      </View>
    );
  }

  return (
    <View>
      <Panel>
        <View style={styles.potmHead}>
          <Txt variant="h3">Player of the match</Txt>
          {fans.fansDisagree ? <RubberStamp text="Fans disagree" /> : null}
        </View>
        <View style={styles.potm}>
          <View style={styles.potmCol}>
            <Txt variant="label" tone="red" style={styles.small}>
              Fans{fansPotm ? ` · ${fansPotm.avg.toFixed(1)}` : ''}
            </Txt>
            {fansPotm ? (
              <>
                <Txt variant="h3" numberOfLines={2} style={styles.potmName}>
                  {getPlayer(fansPotm.playerId)?.name}
                </Txt>
                <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
                  from {fansPotm.count} {fansPotm.count === 1 ? 'fan' : 'fans'}
                </Txt>
              </>
            ) : (
              <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
                Not crowned yet. Needs at least 3 fan ratings.
              </Txt>
            )}
          </View>
          <View style={styles.potmCol}>
            <Txt variant="label" tone="soft" style={styles.small}>
              Official{officialAvg != null ? ` · ${officialAvg.toFixed(1)}` : ''}
            </Txt>
            <Txt variant="h3" numberOfLines={2} style={styles.potmName}>
              {officialName ?? 'Not awarded'}
            </Txt>
            {official ? (
              <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
                {getTeam(official.teamId).short} · {official.role}
              </Txt>
            ) : null}
          </View>
        </View>
      </Panel>

      <Panel style={{ marginTop: 12, paddingTop: 10, paddingBottom: 6 }}>
        <TableHead />
        {players.map((p, i) => (
          <PlayerRow key={p.playerId} rank={i + 1} row={p} crowned={p.playerId === fansPotm?.playerId} last={i === players.length - 1} />
        ))}
      </Panel>
    </View>
  );
}

/** A red rubber stamp, inked at a slight angle (the C mockup's "FANS DISAGREE"). */
function RubberStamp({ text }: { text: string }) {
  const t = useTheme();
  return (
    <View accessibilityRole="text" accessibilityLabel={text} style={[styles.stamp, { borderColor: t.red }]}>
      <Txt tone="red" style={styles.stampText}>
        {text}
      </Txt>
    </View>
  );
}

function TableHead() {
  const t = useTheme();
  return (
    <View style={[styles.row, { borderBottomColor: t.rule, paddingVertical: 6 }]}>
      <Txt variant="label" tone="faint" style={[styles.small, styles.rank]}>
        #
      </Txt>
      <Txt variant="label" tone="faint" style={[styles.small, { flex: 1 }]}>
        Player
      </Txt>
      <Txt variant="label" tone="faint" style={[styles.small, { width: 64 + 40, textAlign: 'right' }]}>
        Avg /10
      </Txt>
      <Txt variant="label" tone="faint" style={[styles.small, styles.count]}>
        Fans
      </Txt>
    </View>
  );
}

function PlayerRow({ rank, row, crowned, last }: { rank: number; row: PlayerAverage; crowned: boolean; last: boolean }) {
  const t = useTheme();
  const player = getPlayer(row.playerId);
  return (
    <View style={[styles.row, { borderBottomColor: last ? 'transparent' : t.rule }]}>
      <Txt variant="monoSmall" tone="faint" style={styles.rank}>
        {String(rank).padStart(2, '0')}
      </Txt>
      <View style={styles.name}>
        <Txt variant="bodySmall" numberOfLines={1} style={{ flexShrink: 1, fontFamily: crowned ? fonts.uiBold : fonts.serif }}>
          {player?.name}
        </Txt>
        <Txt variant="monoSmall" tone="faint" style={{ fontSize: 10, marginLeft: 6 }}>
          {player ? getTeam(player.teamId).short : ''}
        </Txt>
        {crowned ? (
          <View style={{ marginLeft: 6 }} accessibilityLabel="Fans' player of the match">
            <CricketBall size={10} />
          </View>
        ) : null}
      </View>
      <View style={[styles.track, { backgroundColor: t.rule }]}>
        <View style={{ width: `${Math.min(100, Math.max(0, row.avg * 10))}%`, height: '100%', backgroundColor: t.ink }} />
      </View>
      <Txt variant="mono" style={styles.avg}>
        {row.avg.toFixed(1)}
      </Txt>
      <Txt variant="monoSmall" tone="faint" style={styles.count}>
        {row.count}
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  potmHead: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', minHeight: 28 },
  potm: { flexDirection: 'row', marginTop: 10 },
  potmCol: { flex: 1, marginRight: 12 },
  potmName: { fontSize: 16, lineHeight: 21, marginTop: 3 },
  small: { fontSize: 10 },
  stamp: { borderWidth: 2, paddingHorizontal: 6, paddingVertical: 2, marginLeft: 8, transform: [{ rotate: '6deg' }] },
  stampText: { fontFamily: fonts.heading, fontSize: 11, lineHeight: 15, letterSpacing: 1.1, textTransform: 'uppercase' },
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 8, borderBottomWidth: StyleSheet.hairlineWidth },
  rank: { width: 24 },
  name: { flex: 1, flexDirection: 'row', alignItems: 'center', marginRight: 8 },
  track: { width: 64, height: 4 },
  avg: { width: 40, textAlign: 'right' },
  count: { width: 34, textAlign: 'right' },
});
