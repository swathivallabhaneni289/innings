import { router } from 'expo-router';
import { Pressable, StyleSheet, useWindowDimensions, View } from 'react-native';

import { Panel, Rule, SplitBar, Txt } from '@/components';
import { useNow } from '@/data';
import type { PredictionSplit } from '@/domain/api';
import { competitionShort } from '@/domain/constants';
import type { Match, Prediction } from '@/domain/types';
import { clock, countdown, dayLabel, isLocked, relative } from '@/lib/format';
import { PredictionPicker } from '@/features/predict/PredictionPicker';
import { fonts } from '@/theme';
import { TeamCodes } from './TeamCodes';

/** The next big match: three quick picks, and how your friends have called it. */
export function YourCallCard({ match, split, mine }: { match: Match; split: PredictionSplit | null; mine: Prediction | null }) {
  const now = useNow(5_000);
  const { width } = useWindowDimensions();
  const locked = isLocked(match, now);
  return (
    <Panel style={{ padding: 16 }}>
      <Pressable onPress={() => router.push('/match/' + match.id)} accessibilityRole="button">
        <Txt variant="label" tone="soft" numberOfLines={1} style={{ fontSize: 10 }}>
          Next big match · {competitionShort(match.competitionId)} · {match.stage}
        </Txt>
        <TeamCodes teamIds={match.teamIds} width={width - 40 - 32} max={40} style={{ marginTop: 6 }} />
        <Txt variant="bodySmall" tone="soft" numberOfLines={1} style={{ marginTop: 4 }}>
          {match.venue.name}, {match.venue.city}
        </Txt>
        <View style={[styles.row, { marginTop: 2 }]}>
          <Txt variant="bodySmall" style={{ fontFamily: fonts.ui }}>
            {dayLabel(match.startTime, now)}, {clock(match.startTime)}
          </Txt>
          <LockCountdown match={match} />
        </View>
      </Pressable>

      <Rule style={{ marginTop: 14, marginBottom: 2 }} />
      <PredictionPicker compact match={match} mine={mine} locked={locked} />

      {split ? (
        <View style={{ marginTop: 18 }}>
          <Txt variant="bodySmall" style={{ fontFamily: fonts.uiBold, marginBottom: 6 }}>
            Friends&apos; call
            <Txt variant="bodySmall" tone="soft">
              {split.friendsTotal > 0 ? `  ·  ${split.friendsTotal} picked` : "  ·  nobody's called it yet"}
            </Txt>
          </Txt>
          <SplitBar teamIds={match.teamIds} shares={split.friendShares} />
        </View>
      ) : null}
    </Panel>
  );
}

/** Ticks every second on its own, so the picker above doesn't re-render with it. */
function LockCountdown({ match }: { match: Match }) {
  const now = useNow(1_000);
  const locked = isLocked(match, now);
  const ms = Date.parse(match.startTime) - now;
  if (locked) {
    return (
      <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
        Locked 🔒
      </Txt>
    );
  }
  if (ms > 86_400_000) {
    return (
      <Txt variant="bodySmall" style={{ fontFamily: fonts.ui }}>
        Locks {relative(match.startTime, now)}
      </Txt>
    );
  }
  return (
    <Txt variant="bodySmall" style={{ fontFamily: fonts.ui }}>
      Locks in{' '}
      <Txt variant="mono" style={{ fontSize: 13 }}>
        {countdown(match.startTime, now)}
      </Txt>
    </Txt>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
});
