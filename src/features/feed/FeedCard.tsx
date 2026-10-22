// CONTRACT (owned by the Home agent): props below are fixed; Match Page reviews and Profile may import it.
import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Avatar, Txt } from '@/components';
import { getTeam, playerName, useNow } from '@/data';
import type { FeedItem } from '@/domain/api';
import { watchPlaceLabel } from '@/domain/constants';
import type { MatchLog } from '@/domain/types';
import { relative } from '@/lib/format';
import { fonts, useTheme } from '@/theme';
import { StubActions } from './StubActions';

export interface FeedCardProps {
  item: FeedItem;
  showMatch?: boolean; // false inside the Match Page, where the match is already known
  onPress?: () => void; // default: open /stub/[id]
}

/** "Hostel room, with 6 people" — where a stub was watched. */
export function placeLine(log: MatchLog): string {
  const place = log.placeNote?.trim() || watchPlaceLabel(log.watchedAt);
  const n = log.companions ?? log.watchedWith.length;
  return n > 0 ? `${place}, with ${n} ${n === 1 ? 'person' : 'people'}` : place;
}

/** The friend's rating inked on as a round rubber stamp: the number of balls, slightly askew. */
export function RatingStamp({ rating, size = 56 }: { rating: number; size?: number }) {
  const t = useTheme();
  return (
    <View
      accessibilityLabel={`${rating} out of 5 balls`}
      style={[styles.stamp, { width: size, height: size, borderRadius: size / 2, borderColor: t.red }]}>
      <Txt style={{ fontFamily: fonts.condensed, fontSize: size * 0.4, lineHeight: size * 0.42, color: t.red }}>{String(rating)}</Txt>
      <Txt style={{ fontFamily: fonts.monoMedium, fontSize: 7, lineHeight: 9, letterSpacing: 0.8, color: t.red }}>BALLS</Txt>
    </View>
  );
}

/** A friend's logged stub: who, which match, their rating as a stamp, the review, where they watched, and like / reply / "I was there too". */
export function FeedCard({ item, showMatch = true, onPress }: FeedCardProps) {
  const t = useTheme();
  const now = useNow(60_000);
  const { log, user, match } = item;
  const open = onPress ?? (() => router.push('/stub/' + log.id));
  const sub = showMatch
    ? `${getTeam(match.teamIds[0]).short} vs ${getTeam(match.teamIds[1]).short} · ${match.venue.name}`
    : `@${user.handle}`;
  return (
    <Pressable onPress={open} accessibilityRole="button" style={({ pressed }) => [styles.card, { backgroundColor: t.card, opacity: pressed ? 0.9 : 1 }]}>
      <View style={styles.head}>
        <Pressable onPress={() => router.push('/user/' + user.id)} hitSlop={6} style={styles.who} accessibilityRole="button" accessibilityLabel={user.name}>
          <Avatar user={user} size={34} />
          <View style={{ marginLeft: 10, flex: 1 }}>
            <Txt variant="bodySmall" numberOfLines={1} style={{ fontFamily: fonts.uiBold, fontSize: 14.5, lineHeight: 19 }}>
              {user.name}
              <Txt variant="bodySmall" tone="faint">
                {'  ·  '}
                {relative(log.createdAt, now)}
              </Txt>
            </Txt>
            <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
              {sub}
            </Txt>
          </View>
        </Pressable>
        <View style={styles.stampWrap}>
          <RatingStamp rating={log.rating} />
          {log.instantClassic ? (
            <Txt style={styles.heart} accessibilityLabel="Instant classic">
              ❤️
            </Txt>
          ) : null}
        </View>
      </View>

      {log.review ? (
        <Txt variant="body" numberOfLines={4} style={{ marginTop: 10 }}>
          {log.review}
        </Txt>
      ) : null}

      <Txt variant="bodySmall" tone="soft" numberOfLines={2} style={{ marginTop: log.review ? 8 : 10 }}>
        📍 {placeLine(log)}
        {log.fanPotmId ? (
          <>
            {'  ·  '}Fan POTM <Txt variant="bodySmall" style={{ fontFamily: fonts.ui }}>{playerName(log.fanPotmId)}</Txt>
          </>
        ) : null}
      </Txt>

      <View style={[styles.actions, { borderTopColor: t.rule }]}>
        <StubActions item={item} onReply={() => router.push('/stub/' + log.id)} />
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: { padding: 14, paddingBottom: 8 },
  head: { flexDirection: 'row', alignItems: 'center' },
  who: { flex: 1, flexDirection: 'row', alignItems: 'center', minHeight: 44 },
  stampWrap: { marginLeft: 10, marginRight: 2 },
  stamp: { borderWidth: 2, alignItems: 'center', justifyContent: 'center', transform: [{ rotate: '-10deg' }] },
  heart: { position: 'absolute', top: -6, left: -8, fontSize: 14 },
  actions: { marginTop: 12, paddingTop: 4, borderTopWidth: StyleSheet.hairlineWidth },
});
