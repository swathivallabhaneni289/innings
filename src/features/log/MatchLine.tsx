import { Feather } from '@expo/vector-icons';
import { Pressable, StyleSheet, View } from 'react-native';

import { LiveTag, TicketStub, Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { Match } from '@/domain/types';
import { dayLabel, recordedDay, stubNo as fmtStubNo } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

// Short competition names ("World Cup", "INTL", "IPL"): a classic's full name ("ICC Men's T20 World Cup 2022")
// clipped on these slim tickets, and the stub already prints its full date.
const competitionOf = (m: Match) => competitionShort(m.competitionId);

/** "Today", "Sat 24 Oct"; classics and archive matches print their recorded date, "2 Apr 2011". */
function dateOf(m: Match, now: number) {
  const day = recordedDay(m);
  if (day) {
    const [y, mo, d] = day.split('-').map(Number);
    return `${d} ${MONTHS[mo - 1]} ${y}`;
  }
  return dayLabel(m.startTime, now);
}

/** Fine-print state on the ticket: red only for Live. */
function StateTag({ m }: { m: Match }) {
  if (m.state === 'live') return <LiveTag />;
  const [label, tone] =
    m.kind === 'classic' ? (['Classic', 'ink'] as const) : m.state === 'upcoming' ? (['Not started', 'faint'] as const) : (['Ended', 'soft'] as const);
  return (
    <Txt variant="label" tone={tone} style={{ fontSize: 10 }}>
      {label}
    </Txt>
  );
}

/**
 * A match printed as a slim ticket on its competition's stock: fine print and state on top, the team codes in
 * stadium type, one line of status; the stub carries the date and a barcode (or the serial when re-logging).
 */
function SlimTicket({
  m,
  now,
  size,
  serial,
  selected,
}: {
  m: Match;
  now: number;
  size: 'row' | 'line';
  serial?: string;
  selected?: boolean;
}) {
  const t = useTheme();
  const [a, b] = m.teamIds.map(getTeam);
  const line = size === 'line';
  const code = line ? 32 : 28;
  const sub = (m.state === 'live' ? m.live?.statusText : m.result?.text ?? m.classic?.headline) || `${m.venue.name}, ${m.venue.city}`;
  const fine = line ? `${competitionOf(m)} · ${m.stage} · ${m.venue.name}` : `${competitionOf(m)} · ${m.stage}`;
  return (
    <TicketStub
      stock={stockFor(m, t)}
      height={line ? 112 : 100}
      sideSize={78}
      main={
        <View style={styles.main}>
          <View style={styles.fineRow}>
            <Txt variant="label" numberOfLines={line ? 2 : 1} style={styles.fine}>
              {fine}
            </Txt>
            {line && m.state !== 'live' ? null : <StateTag m={m} />}
          </View>
          <View style={styles.teams}>
            <Txt variant="display" numberOfLines={1} style={{ fontSize: code, lineHeight: code, flexShrink: 1 }}>
              {a.short}
            </Txt>
            <Txt style={{ fontFamily: fonts.label, fontSize: 12, marginHorizontal: 7 }}>vs</Txt>
            <Txt variant="display" numberOfLines={1} style={{ fontSize: code, lineHeight: code, flexShrink: 1 }}>
              {b.short}
            </Txt>
          </View>
          <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
            {sub}
          </Txt>
        </View>
      }
      side={
        <View style={styles.stub}>
          {serial ? (
            <Txt variant="label" tone="red" numberOfLines={1} style={{ fontSize: 9.5, letterSpacing: 0.4 }}>
              {serial}
            </Txt>
          ) : null}
          <Txt
            numberOfLines={1}
            adjustsFontSizeToFit
            minimumFontScale={0.8}
            align="center"
            style={{ fontFamily: fonts.uiBold, fontSize: 12, lineHeight: 15, alignSelf: 'stretch' }}>
            {dateOf(m, now)}
          </Txt>
          {selected ? (
            <Feather name="check" size={22} color={t.ink} accessibilityLabel="Selected" />
          ) : (
            <Barcode seed={m.id} width={40} height={line ? 30 : 28} />
          )}
        </View>
      }
    />
  );
}

/** A match in the picker, printed as a slim ticket. Not-started matches can't be logged yet. */
export function MatchRow({
  m,
  now,
  selected,
  onPress,
}: {
  m: Match;
  now: number;
  selected?: boolean;
  onPress: () => void;
}) {
  const disabled = m.state === 'upcoming';
  const [a, b] = m.teamIds.map(getTeam);
  return (
    <Pressable
      onPress={onPress}
      disabled={disabled}
      accessibilityRole="button"
      accessibilityLabel={`${a.name} versus ${b.name}, ${m.stage}`}
      accessibilityState={{ selected, disabled }}
      style={({ pressed }) => [styles.row, { opacity: disabled ? 0.45 : pressed ? 0.75 : 1 }]}>
      <SlimTicket m={m} now={now} size="row" selected={selected} />
    </Pressable>
  );
}

/** The chosen match at the top of the later steps, as a slim ticket. Re-logs print the existing stub number. */
export function MatchLine({ m, now, stubNo }: { m: Match; now: number; stubNo?: number }) {
  return <SlimTicket m={m} now={now} size="line" serial={stubNo ? fmtStubNo(stubNo) : undefined} />;
}

const styles = StyleSheet.create({
  row: { marginBottom: 10 },
  main: { flex: 1, justifyContent: 'space-between' },
  fineRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  fine: { fontSize: 10, flexShrink: 1, marginRight: 8 },
  teams: { flexDirection: 'row', alignItems: 'baseline' },
  stub: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
});
