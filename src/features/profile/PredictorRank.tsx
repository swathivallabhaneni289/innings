import { router } from 'expo-router';
import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { Button, Panel, Rule, Txt } from '@/components';
import type { FanStats } from '@/data';
import type { BadgeDef } from '@/domain/types';
import { fonts, useTheme } from '@/theme';
import { thousands } from './constants';

const TILTS = ['-4deg', '3deg', '-2deg', '5deg', '-3deg'];

/** Rank among the fan's circle, the current correct-winner streak, and earned badges as red rubber stamps. */
export function PredictorRank({ stats, isMe, ownerFirstName }: { stats: FanStats; isMe: boolean; ownerFirstName: string }) {
  const [open, setOpen] = useState<BadgeDef['id'] | null>(null);
  const solo = stats.rankOf <= 1;
  // Following fans but no call settled yet: "#9 of 9" would read as a verdict on someone who hasn't played.
  const noCalls = !solo && stats.points === 0 && stats.accuracy == null;
  const openBadge = stats.badges.find((b) => b.id === open);

  return (
    <Panel>
      {solo || noCalls ? (
        <View>
          <Txt variant="h2">Unranked</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
            {noCalls
              ? isMe
                ? `No calls settled yet. Call a winner before the first ball to join the table of ${stats.rankOf}.`
                : `${ownerFirstName} hasn't had a call settled yet.`
              : isMe
                ? 'Rank is counted among the fans you follow. Follow a few to start a rivalry.'
                : `${ownerFirstName} isn't following anyone yet, so there's no table to top.`}
          </Txt>
          {isMe && solo ? (
            <View style={{ flexDirection: 'row', marginTop: 12 }}>
              <Button label="Find friends" kind="outline" size="sm" onPress={() => router.push('/friends')} style={{ minHeight: 44 }} />
            </View>
          ) : null}
        </View>
      ) : (
        <View style={styles.rankRow}>
          <Txt variant="display" style={styles.rank}>
            #{stats.rank}
          </Txt>
          <View style={{ flex: 1, marginLeft: 14 }}>
            <Txt variant="h3">
              of {stats.rankOf} {isMe ? 'among friends' : `in ${ownerFirstName}'s circle`}
            </Txt>
            <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
              {thousands(stats.points)} bragging points
            </Txt>
          </View>
        </View>
      )}

      <Rule style={{ marginVertical: 14 }} />
      <View style={styles.streakRow}>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label }}>
          Current streak
        </Txt>
        <Txt variant="bodySmall" tone={stats.streak > 0 ? 'ink' : 'faint'} style={{ marginLeft: 10, flex: 1, fontFamily: fonts.uiBold }}>
          {stats.streak > 0 ? `${stats.streak} ${stats.streak === 1 ? 'winner' : 'winners'} in a row` : 'None going'}
        </Txt>
      </View>

      <Txt variant="h3" style={{ marginTop: 18 }}>
        Badges
      </Txt>
      {stats.badges.length === 0 ? (
        <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
          {isMe ? 'No stamps on your ticket yet. Call a winner to earn your first.' : 'No stamps on the ticket yet.'}
        </Txt>
      ) : (
        <>
          <View style={styles.badges}>
            {stats.badges.map((b, i) => (
              <BadgeStamp key={b.id} badge={b} tilt={TILTS[i % TILTS.length]} selected={open === b.id} onPress={() => setOpen(open === b.id ? null : b.id)} />
            ))}
          </View>
          {openBadge ? (
            <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
              <Txt variant="bodySmall" style={{ fontFamily: fonts.uiBold }}>
                {openBadge.title}.
              </Txt>{' '}
              {openBadge.description}
            </Txt>
          ) : (
            <Txt variant="bodySmall" tone="faint" style={{ marginTop: 2 }}>
              Tap a stamp to see how it was earned.
            </Txt>
          )}
        </>
      )}
    </Panel>
  );
}

/** A badge inked like the rubber stamps on the tickets: one red border, heavy caps, slightly askew. Tap for what it means. */
function BadgeStamp({ badge, tilt, selected, onPress }: { badge: BadgeDef; tilt: string; selected: boolean; onPress: () => void }) {
  const t = useTheme();
  return (
    <Pressable
      onPress={onPress}
      hitSlop={{ top: 10, bottom: 10, left: 4, right: 4 }}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      accessibilityLabel={`${badge.title}. ${badge.description}`}
      style={[styles.stamp, { borderColor: t.red, backgroundColor: selected ? t.redSoft : 'transparent', transform: [{ rotate: tilt }] }]}>
      <Txt tone="red" style={styles.stampText}>
        {badge.title}
      </Txt>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  rankRow: { flexDirection: 'row', alignItems: 'center' },
  rank: { fontSize: 56, lineHeight: 58 },
  streakRow: { flexDirection: 'row', alignItems: 'baseline' },
  badges: { flexDirection: 'row', flexWrap: 'wrap', marginTop: 10, marginHorizontal: -6, paddingVertical: 4 },
  stamp: { borderWidth: 2, paddingHorizontal: 7, paddingVertical: 3, margin: 6, opacity: 0.92 },
  stampText: { fontFamily: fonts.heading, fontSize: 11, lineHeight: 14, letterSpacing: 1.1, textTransform: 'uppercase' },
});
