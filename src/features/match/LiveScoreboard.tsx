import { StyleSheet, View } from 'react-native';

import { Avatar, CricketBall, LiveTag, Panel, Rule, TeamBadge, Txt } from '@/components';
import { getTeam } from '@/data';
import type { InningsScore, Match, User } from '@/domain/types';
import { inningsLine } from '@/lib/format';
import { fonts, useTheme } from '@/theme';

/**
 * Both sides on scoreboard lines: the team code in stadium type over its kit-colour bar, the full name, the score.
 * `battingTeamId` marks the side in (live); `winnerId` marks the winner (result).
 * Renders nothing when there are no innings to show (some Vault classics).
 */
export function InningsTable({
  match,
  innings,
  battingTeamId,
  winnerId,
  live,
}: {
  match: Match;
  innings: InningsScore[];
  battingTeamId?: string;
  winnerId?: string | null;
  live?: boolean;
}) {
  const t = useTheme();
  if (innings.length === 0 && !live) return null;
  return (
    <View>
      {match.teamIds.map((id, i) => {
        const team = getTeam(id);
        const mine = innings.filter((s) => s.teamId === id);
        const batting = battingTeamId === id;
        const won = winnerId === id;
        return (
          <View key={id} style={[styles.line, i > 0 ? { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule } : null]}>
            <View style={styles.code}>
              <TeamBadge teamId={id} size="md" />
            </View>
            <View style={styles.name}>
              <Txt variant="body" tone={battingTeamId && !batting ? 'soft' : 'ink'} numberOfLines={2} style={styles.teamName}>
                {team.name}
              </Txt>
              {batting ? (
                <View style={styles.marker} accessibilityLabel="batting">
                  <CricketBall size={11} />
                </View>
              ) : null}
              {won ? (
                <Txt variant="label" tone="green" style={styles.won}>
                  Won
                </Txt>
              ) : null}
            </View>
            {mine.length > 0 ? (
              <Txt variant="mono" tone={!battingTeamId || batting ? 'ink' : 'soft'} style={styles.score}>
                {mine.map(inningsLine).join(' & ')}
              </Txt>
            ) : (
              <Txt variant="bodySmall" tone="soft" style={styles.score}>
                {live ? 'Yet to bat' : '—'}
              </Txt>
            )}
          </View>
        );
      })}
    </View>
  );
}

/** DURING: the live scoreboard panel, refreshed with the page every 20 s. No band on top: the LIVE tag is the accent. */
export function LiveScoreboard({ match }: { match: Match }) {
  const live = match.live;
  return (
    <Panel style={{ marginTop: 16 }}>
      <View style={styles.head}>
        <LiveTag />
        <Txt variant="bodySmall" tone="faint">
          Updates every 20 s
        </Txt>
      </View>
      <InningsTable match={match} innings={live?.innings ?? []} battingTeamId={live?.battingTeamId} live />
      {live?.statusText ? (
        <>
          <Rule style={{ marginTop: 4, marginBottom: 12 }} />
          <Txt variant="body" style={{ fontFamily: fonts.uiBold }}>
            {live.statusText}
          </Txt>
        </>
      ) : null}
    </Panel>
  );
}

/** AFTER: the result line over both innings. */
export function ResultCard({ match }: { match: Match }) {
  const r = match.result;
  if (!r) return null;
  const innings = r.innings; // may be empty for some classics: the table then renders nothing
  return (
    <Panel style={{ marginTop: 16 }}>
      <Txt variant="label" tone="soft" style={{ fontSize: 10, marginBottom: 4 }}>
        Result
      </Txt>
      <Txt variant="h2" style={{ marginBottom: innings.length > 0 ? 8 : 0 }}>
        {r.text}
      </Txt>
      <InningsTable match={match} innings={innings} winnerId={r.winnerId} />
    </Panel>
  );
}

/** Friends who are watching (live) or said "I'm watching" (before): avatars with first names. */
export function FriendsWatching({ users, emptyText }: { users: User[]; emptyText?: string }) {
  if (users.length === 0) {
    return emptyText ? (
      <Txt variant="bodySmall" tone="soft">
        {emptyText}
      </Txt>
    ) : null;
  }
  return (
    <View style={styles.friends}>
      {users.map((u) => (
        <View key={u.id} style={styles.friend}>
          <Avatar user={u} size={36} />
          <Txt variant="bodySmall" numberOfLines={1} style={styles.friendName}>
            {u.name.split(/\s+/)[0]}
          </Txt>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  line: { flexDirection: 'row', alignItems: 'center', paddingVertical: 10 },
  code: { width: 58 },
  name: { flex: 1, flexDirection: 'row', alignItems: 'center', marginRight: 10 },
  teamName: { flexShrink: 1, fontFamily: fonts.ui, fontSize: 14, lineHeight: 19 },
  marker: { marginLeft: 8 },
  won: { fontSize: 10, marginLeft: 8 },
  score: { textAlign: 'right', flexShrink: 0 },
  head: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 4 },
  friends: { flexDirection: 'row', flexWrap: 'wrap' },
  friend: { width: 64, alignItems: 'center', marginRight: 6, marginBottom: 12 },
  friendName: { fontFamily: fonts.ui, fontSize: 12, lineHeight: 16, marginTop: 4, maxWidth: 62 },
});
