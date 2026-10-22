import { router } from 'expo-router';
import { Pressable, StyleSheet, useWindowDimensions, View } from 'react-native';

import { BallRating, LiveTag, Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { InningsScore, Match } from '@/domain/types';
import { codeFontSize, serialFor, TicketFrame } from '@/features/match/TicketFrame';
import { clock, inningsLine, relative } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';

/** "IPL · T20 · Match 21" — without saying "The Hundred" twice. */
const kicker = (m: Match) => [...new Set([competitionShort(m.competitionId), m.format, m.stage])].join(' · ');

const STUB_W = 76; // TicketFrame's default stub
const VS_W = 34; // "vs" with its margins

/** One line per side, in batting order: "ENG 412 (132.1)", "SL 81/2 (31)"; a second Test innings joins with "&". */
function scoreLines(innings: InningsScore[]): string[] {
  const order: string[] = [];
  for (const i of innings) if (!order.includes(i.teamId)) order.push(i.teamId);
  return order.map((id) => `${getTeam(id).short} ${innings.filter((i) => i.teamId === id).map(inningsLine).join(' & ')}`);
}

/** One fixture on the Matches list, printed as a ticket on its competition's card stock. */
export function MatchCard({ match, rating, now }: { match: Match; rating?: { avg: number; count: number }; now: number }) {
  const t = useTheme();
  const { width } = useWindowDimensions();
  const [a, b] = match.teamIds.map(getTeam);
  const live = match.state === 'live';
  // Ticket body = screen less 20 px gutters, the stub, its perforation and 14 px padding each side; "AUS-W vs IND-W" must fit.
  const codeSize = codeFontSize(a.short.length + b.short.length, width - 40 - STUB_W - 2 - 28 - VS_W, 40);
  return (
    <Pressable
      onPress={() => router.push(`/match/${match.id}`)}
      accessibilityRole="button"
      accessibilityLabel={`${a.name} versus ${b.name}, ${match.stage}`}
      style={({ pressed }) => [styles.card, { opacity: pressed ? 0.75 : 1 }]}>
      <TicketFrame
        stock={stockFor(match, t)}
        stub={
          <>
            <Txt variant="label" tone="red" style={styles.serial}>
              {serialFor(match)}
            </Txt>
            <Barcode seed={match.id} width={44} height={40} />
          </>
        }>
        <View style={styles.top}>
          <Txt variant="label" numberOfLines={2} style={styles.fine}>
            {kicker(match)}
          </Txt>
          {live ? (
            <LiveTag />
          ) : match.state === 'upcoming' ? (
            <Txt variant="label" style={styles.fineRight}>
              {clock(match.startTime)}
            </Txt>
          ) : (
            <Txt variant="label" tone="soft" style={styles.fineRight}>
              Result
            </Txt>
          )}
        </View>

        <View style={styles.teams}>
          <Code short={a.short} colour={a.colour} size={codeSize} />
          <Txt variant="bodySmall" style={styles.vs}>
            vs
          </Txt>
          <Code short={b.short} colour={b.colour} size={codeSize} />
        </View>
        <Txt variant="bodySmall" tone="soft" numberOfLines={2} style={{ marginTop: 6 }}>
          {match.venue.name}, {match.venue.city}
        </Txt>

        <View style={{ marginTop: 10 }}>
          <CardState match={match} rating={rating} now={now} />
        </View>
      </TicketFrame>
    </Pressable>
  );
}

/** A team code in stadium type with a thin kit-colour bar under it. */
function Code({ short, colour, size }: { short: string; colour: string; size: number }) {
  return (
    <View>
      <Txt variant="display" numberOfLines={1} style={{ fontSize: size, lineHeight: size + 2 }}>
        {short}
      </Txt>
      <View style={[styles.kit, { backgroundColor: colour }]} />
    </View>
  );
}

function CardState({ match, rating, now }: { match: Match; rating?: { avg: number; count: number }; now: number }) {
  if (match.state === 'live') {
    const lines = scoreLines(match.live?.innings ?? []);
    return (
      <View>
        {lines.length > 0 ? (
          lines.map((line) => (
            <Txt key={line} variant="mono" numberOfLines={1}>
              {line}
            </Txt>
          ))
        ) : (
          <Txt variant="bodySmall" tone="soft">
            Toss done, first ball coming up
          </Txt>
        )}
        {match.live?.statusText ? (
          <Txt variant="bodySmall" numberOfLines={2} style={styles.status}>
            {match.live.statusText}
          </Txt>
        ) : null}
      </View>
    );
  }
  if (match.state === 'upcoming') {
    return (
      <Txt variant="bodySmall" tone="soft">
        First ball {relative(match.startTime, now)}
      </Txt>
    );
  }
  const lines = scoreLines(match.result?.innings ?? []);
  return (
    <View>
      {lines.map((line) => (
        <Txt key={line} variant="mono" numberOfLines={1}>
          {line}
        </Txt>
      ))}
      <Txt variant="bodySmall" numberOfLines={2} style={[styles.status, lines.length === 0 ? { marginTop: 0 } : null]}>
        {match.result?.text ?? 'Result to follow'}
      </Txt>
      <View style={styles.rating}>
        {rating && rating.count > 0 ? (
          <>
            <BallRating value={rating.avg} size={13} gap={2} />
            <Txt variant="mono" style={{ marginLeft: 8 }}>
              {rating.avg.toFixed(1)}
            </Txt>
            <Txt variant="bodySmall" tone="soft" style={{ marginLeft: 6 }}>
              · {rating.count} {rating.count === 1 ? 'rating' : 'ratings'}
            </Txt>
          </>
        ) : (
          <Txt variant="bodySmall" tone="soft">
            No fan ratings yet
          </Txt>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: { marginBottom: 12 },
  top: { flexDirection: 'row', alignItems: 'flex-start', justifyContent: 'space-between' },
  fine: { fontSize: 10, flexShrink: 1, marginRight: 8 },
  fineRight: { fontSize: 10 },
  teams: { flexDirection: 'row', alignItems: 'center', marginTop: 10 },
  kit: { height: 3, marginTop: 1 },
  vs: { width: VS_W, textAlign: 'center', fontFamily: fonts.label },
  status: { fontFamily: fonts.uiBold, marginTop: 6 },
  rating: { flexDirection: 'row', alignItems: 'center', marginTop: 8 },
  serial: { fontSize: 9.5, letterSpacing: 0.4 },
});
