import { StyleSheet, View } from 'react-native';

import { Avatar, Panel, SectionHeader, SplitBar, Txt } from '@/components';
import { useNow, type MatchPageResponse, type PredictionSplit } from '@/data';
import type { Match, Prediction, User } from '@/domain/types';
import { PredictionPicker } from '@/features/predict/PredictionPicker';
import { countdown, isLocked } from '@/lib/format';
import { useTheme } from '@/theme';

/** BEFORE: countdown to the first ball, your call, friends' split (with faces) and the whole crowd's split. */
export function PredictionSection({ data }: { data: MatchPageResponse }) {
  const now = useNow(1000);
  const { match, split, mine } = data;
  const locked = data.locked || isLocked(match, now);
  return (
    <View>
      <Countdown match={match} now={now} locked={locked} />

      <SectionHeader title="Your call" />
      <PredictionPicker match={match} mine={mine.prediction} split={split} locked={locked} />

      <SectionHeader title="Friends' predictions" />
      <FriendsSplit match={match} split={split} />
      <View style={{ marginTop: 22 }}>
        <SplitBar
          teamIds={match.teamIds}
          shares={split.shares}
          label={split.total > 0 ? `All fans · ${split.total} ${split.total === 1 ? 'call' : 'calls'}` : 'All fans · no calls yet'}
        />
      </View>
    </View>
  );
}

/** DURING: your picks, sealed at the first ball (the picker's read-only view also covers "no prediction made"). */
export function YourPicks({ match, mine }: { match: Match; mine: Prediction | null }) {
  return <PredictionPicker match={match} mine={mine} locked />;
}

function Countdown({ match, now, locked }: { match: Match; now: number; locked: boolean }) {
  const secs = Math.max(0, Math.floor((Date.parse(match.startTime) - now) / 1000));
  const days = Math.floor(secs / 86_400);
  const value = days >= 1 ? `${days}d ${Math.floor((secs % 86_400) / 3600)}h` : countdown(match.startTime, now);
  return (
    <Panel style={{ marginTop: 16 }}>
      {locked ? (
        <>
          <Txt variant="h2">Locked 🔒</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
            The first ball is down. Picks are sealed until the result.
          </Txt>
        </>
      ) : (
        <View style={styles.countRow}>
          <View style={{ flex: 1, marginRight: 12 }}>
            <Txt variant="h3">First ball in</Txt>
            <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
              Predictions lock at the first ball.
            </Txt>
          </View>
          <Txt variant="stat" accessibilityLabel={`First ball in ${value}`}>
            {value}
          </Txt>
        </View>
      )}
    </Panel>
  );
}

function FriendsSplit({ match, split }: { match: Match; split: PredictionSplit }) {
  if (split.friendsTotal === 0) {
    return (
      <Txt variant="bodySmall" tone="soft">
        None of your friends have called it yet. Get yours in first.
      </Txt>
    );
  }
  const [a, b] = match.teamIds;
  const side = (teamId: string) => split.friendPicks.filter((p) => p.teamId === teamId).map((p) => p.user);
  return (
    <View>
      <SplitBar
        teamIds={match.teamIds}
        shares={split.friendShares}
        label={`${split.friendsTotal} ${split.friendsTotal === 1 ? 'friend has' : 'friends have'} called it`}
      />
      <View style={styles.sides}>
        <AvatarStack users={side(a)} />
        <AvatarStack users={side(b)} alignEnd />
      </View>
    </View>
  );
}

const MAX_FACES = 5;

function AvatarStack({ users, alignEnd }: { users: User[]; alignEnd?: boolean }) {
  const t = useTheme();
  const shown = users.slice(0, MAX_FACES);
  const extra = users.length - shown.length;
  return (
    <View style={[styles.stack, alignEnd ? { justifyContent: 'flex-end' } : null]}>
      {shown.map((u, i) => (
        <View key={u.id} style={[styles.face, { borderColor: t.paper, marginLeft: i === 0 ? 0 : -8 }]}>
          <Avatar user={u} size={26} />
        </View>
      ))}
      {extra > 0 ? (
        <Txt variant="monoSmall" tone="soft" style={{ marginLeft: 6 }}>
          +{extra}
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  countRow: { flexDirection: 'row', alignItems: 'center' },
  sides: { flexDirection: 'row', justifyContent: 'space-between', marginTop: 10 },
  stack: { flex: 1, flexDirection: 'row', alignItems: 'center' },
  face: { borderWidth: 2, borderRadius: 15 },
});
