import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { BallRating, Button, Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam } from '@/data';
import type { RatedMatch } from '@/domain/api';
import { ticketDate } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';
import { Perforation } from './Perforation';

const PAD = 14;

/**
 * A collection entry printed in full on archive ticket stock: fine print, the codes, the hook; below the tear line,
 * the scoreline, result, venue, fan rating and the two vault actions.
 */
export function ClassicRow({
  item,
  pinned,
  busy,
  message,
  onTogglePin,
}: {
  item: RatedMatch;
  pinned: boolean;
  busy: boolean;
  message?: string;
  onTogglePin: () => void;
}) {
  const t = useTheme();
  const { match } = item;
  const c = match.classic;
  const [a, b] = match.teamIds.map(getTeam);
  return (
    <View style={styles.wrap}>
      <View style={[styles.ticket, { backgroundColor: stockFor(match, t) }]}>
        <Pressable onPress={() => router.push(`/match/${match.id}`)} accessibilityRole="button">
          {({ pressed }) => (
            <View style={{ opacity: pressed ? 0.7 : 1 }}>
              <Txt variant="label" numberOfLines={1}>
                {ticketDate(match)} · {match.format} · {match.stage}
              </Txt>
              <View style={styles.teams}>
                <Txt variant="display" numberOfLines={1} style={styles.code}>
                  {a.short}
                </Txt>
                <Txt variant="bodySmall" style={styles.vs}>
                  vs
                </Txt>
                <Txt variant="display" numberOfLines={1} style={[styles.code, { flexShrink: 1 }]}>
                  {b.short}
                </Txt>
              </View>
              <Txt variant="h2" style={{ marginTop: 6 }}>
                {c?.headline ?? match.result?.text ?? match.stage}
              </Txt>

              <View style={{ marginVertical: 12 }}>
                <Perforation inset={PAD} />
              </View>

              {c?.scoreline ? <Txt variant="mono">{c.scoreline}</Txt> : null}
              {match.result?.text ? (
                <Txt variant="bodySmall" style={{ marginTop: 4 }}>
                  {match.result.text}
                </Txt>
              ) : null}
              <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
                {match.venue.name}, {match.venue.city}
              </Txt>
              <View style={styles.rating}>
                <BallRating value={item.ratingAvg} size={14} gap={2} />
                <Txt variant="mono" style={{ marginLeft: 8 }}>
                  {item.ratingCount > 0 ? item.ratingAvg.toFixed(1) : '–'}
                </Txt>
                <Txt variant="monoSmall" tone="soft" style={{ marginLeft: 6, flex: 1 }} numberOfLines={1}>
                  {item.ratingCount.toLocaleString()} {item.ratingCount === 1 ? 'fan' : 'fans'}
                </Txt>
                <Barcode seed={match.id} width={44} height={18} />
              </View>
            </View>
          )}
        </Pressable>

        <View style={styles.actions}>
          <Button
            label="I watched this"
            size="sm"
            style={styles.action}
            onPress={() => router.push({ pathname: '/log', params: { matchId: match.id } })}
          />
          <Button
            label={pinned ? 'Pinned to Top 5' : 'Pin to Top 5'}
            kind={pinned ? 'ghost' : 'outline'}
            size="sm"
            loading={busy}
            style={styles.action}
            icon={<Feather name={pinned ? 'check' : 'bookmark'} size={13} color={pinned ? t.red : t.ink} />}
            onPress={onTogglePin}
          />
        </View>
      </View>
      {message ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 8, fontFamily: fonts.ui }}>
          {message}
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { marginBottom: 16 },
  ticket: { padding: PAD },
  teams: { flexDirection: 'row', alignItems: 'baseline', marginTop: 8 },
  code: { fontSize: 34, lineHeight: 34 },
  vs: { marginHorizontal: 8, fontFamily: fonts.label },
  rating: { flexDirection: 'row', alignItems: 'center', marginTop: 10 },
  actions: { flexDirection: 'row', alignItems: 'center', marginTop: 14, gap: 10 },
  action: { minHeight: 44 },
});
