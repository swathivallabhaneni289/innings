import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Button, Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam } from '@/data';
import type { VaultResponse } from '@/domain/api';
import { fonts, stockFor, useTheme } from '@/theme';
import { Perforation } from './Perforation';
import { anniversary, competitionNoYear, matchYear } from './top5';

type OnThisDayData = NonNullable<VaultResponse['onThisDay']>;

const PAD = 16;

/** The classic played on today's date in history, printed on the archive's ticket stock, with a nudge to log it. */
export function OnThisDay({ data }: { data: OnThisDayData }) {
  const t = useTheme();
  const { match, yearsAgo, exact } = data;
  const c = match.classic;
  const [a, b] = match.teamIds.map(getTeam);
  const years = yearsAgo === 1 ? '1 year' : `${yearsAgo} years`;
  const when = exact ? `${years} ago today` : `${years} ago on ${anniversary(match)}`;
  return (
    <Pressable onPress={() => router.push(`/match/${match.id}`)} accessibilityRole="button">
      {({ pressed }) => (
        <View style={[styles.ticket, { backgroundColor: stockFor(match, t), opacity: pressed ? 0.88 : 1 }]}>
          <Txt variant="label" numberOfLines={2}>
            Admit one · {match.stage}
            {c ? ` · ${competitionNoYear(c.competitionName)}` : ''}
          </Txt>

          <View style={styles.yearRow}>
            <Txt variant="display" style={styles.year}>
              {matchYear(match)}
            </Txt>
            <Txt variant="bodySmall" style={styles.when}>
              {when}
            </Txt>
          </View>

          <Txt variant="h2" style={{ marginTop: 10 }}>
            {c?.headline ?? match.result?.text ?? match.stage}
          </Txt>

          <View style={styles.teams}>
            <Txt variant="display" style={styles.code}>
              {a.short}
            </Txt>
            <Txt variant="bodySmall" style={styles.vs}>
              vs
            </Txt>
            <Txt variant="display" style={styles.code}>
              {b.short}
            </Txt>
          </View>

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

          <View style={styles.stub}>
            <Button
              label="I watched this — log it"
              size="sm"
              style={{ minHeight: 44 }}
              onPress={() => router.push({ pathname: '/log', params: { matchId: match.id } })}
            />
            <Barcode seed={match.id} width={64} height={30} />
          </View>
        </View>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  ticket: { padding: PAD },
  yearRow: { flexDirection: 'row', alignItems: 'flex-end', marginTop: 10 },
  year: { fontSize: 64, lineHeight: 64 },
  when: { flex: 1, marginLeft: 12, marginBottom: 6, fontFamily: fonts.uiBold },
  teams: { flexDirection: 'row', alignItems: 'baseline', marginTop: 8 },
  code: { fontSize: 30, lineHeight: 30 },
  vs: { marginHorizontal: 8, fontFamily: fonts.label },
  stub: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginTop: 14 },
});
