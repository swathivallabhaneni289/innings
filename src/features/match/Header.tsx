import { StyleSheet, useWindowDimensions, View } from 'react-native';

import { Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam, useNow } from '@/data';
import { COMPETITIONS } from '@/domain/constants';
import type { Match, Team } from '@/domain/types';
import { clock, dayLabel, recordedDay } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';
import { codeFontSize, serialFor, TicketFrame } from './TicketFrame';

const VS_W = 36; // the small "vs" between the two codes

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/** "2 Apr 2011" from a classic's 'YYYY-MM-DD'. */
const proseDate = (iso: string) => {
  const [y, m, d] = iso.split('-').map(Number);
  return `${d} ${MONTHS[m - 1]} ${y}`;
};

/**
 * Match Page masthead, printed as the match ticket on its competition's card stock: fine print, both team codes in
 * stadium type with "vs" between and the full names small beneath, and a tear-off stub with the venue, the day and a
 * serial. Classics add a Vault stamp and their headline.
 */
export function Header({ match }: { match: Match }) {
  const t = useTheme();
  const now = useNow(60_000);
  const { width } = useWindowDimensions();
  const [a, b] = match.teamIds.map(getTeam);
  // Each code gets half the ticket body: the screen less its 20 px gutters and the ticket's 16 px padding, less the "vs".
  const codeSize = codeFontSize(Math.max(a.short.length, b.short.length), (width - 40 - 32 - VS_W) / 2, 64);
  const classic = match.classic;
  const competition = classic?.competitionName ?? COMPETITIONS.find((c) => c.id === match.competitionId)?.name ?? match.competitionId;
  // Classics and archive matches have no real start time on record, only the day.
  const day = recordedDay(match);
  const when = day ? proseDate(day) : `${dayLabel(match.startTime, now)}, ${clock(match.startTime)}`;

  return (
    <View>
      <TicketFrame
        side="bottom"
        stock={stockFor(match, t)}
        style={{ marginTop: 4 }}
        stub={
          <View style={styles.stub}>
            <View style={{ flex: 1, marginRight: 12 }}>
              <Txt variant="h3">
                {match.venue.name}, {match.venue.city}
              </Txt>
              <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
                {match.format}
                {match.gender === 'women' ? ' · Women' : ''} · {when}
              </Txt>
            </View>
            <View style={styles.serialCol}>
              <Txt variant="label" tone="red" style={styles.serial}>
                {serialFor(match)}
              </Txt>
              <Barcode seed={match.id} width={58} height={22} />
            </View>
          </View>
        }>
        {classic ? (
          <View style={[styles.vault, { borderColor: t.red }]}>
            <Txt variant="label" tone="red" style={{ fontSize: 11 }}>
              From the Vault · {classic.era}
            </Txt>
          </View>
        ) : null}
        <Txt variant="label">
          Admit one · {competition} · {match.stage}
        </Txt>

        <View style={{ marginTop: 14 }} accessible accessibilityRole="header" accessibilityLabel={`${a.name} versus ${b.name}`}>
          <View style={styles.row}>
            <Code team={a} size={codeSize} />
            <Txt variant="bodySmall" style={styles.vs}>
              vs
            </Txt>
            <Code team={b} size={codeSize} end />
          </View>
          <View style={[styles.row, { alignItems: 'flex-start', marginTop: 8 }]}>
            <Txt variant="bodySmall" tone="soft" style={styles.name}>
              {a.name}
            </Txt>
            <View style={{ width: VS_W }} />
            <Txt variant="bodySmall" tone="soft" style={[styles.name, { textAlign: 'right' }]}>
              {b.name}
            </Txt>
          </View>
        </View>
      </TicketFrame>

      {classic ? (
        <View style={{ marginTop: 18 }}>
          <Txt variant="review" style={{ fontSize: 20, lineHeight: 28 }}>
            “{classic.headline}”
          </Txt>
          <Txt variant="mono" style={{ marginTop: 8 }}>
            {classic.scoreline}
          </Txt>
        </View>
      ) : null}
    </View>
  );
}

/** A team code in stadium type with a thin kit-colour bar under it; `end` sets it flush right (the away side). */
function Code({ team, size, end }: { team: Team; size: number; end?: boolean }) {
  return (
    <View style={[styles.codeCol, end ? { alignItems: 'flex-end' } : null]}>
      <View>
        <Txt variant="display" numberOfLines={1} style={{ fontSize: size, lineHeight: size + 2 }}>
          {team.short}
        </Txt>
        <View style={[styles.kit, { backgroundColor: team.colour }]} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  vault: { alignSelf: 'flex-start', borderWidth: 1.5, paddingHorizontal: 8, paddingVertical: 2, marginBottom: 12, transform: [{ rotate: '-2deg' }] },
  row: { flexDirection: 'row', alignItems: 'center' },
  codeCol: { flex: 1, alignItems: 'flex-start' },
  kit: { height: 3, marginTop: 1 },
  vs: { width: VS_W, textAlign: 'center', fontFamily: fonts.label },
  name: { flex: 1, fontFamily: fonts.ui },
  stub: { flexDirection: 'row', alignItems: 'center' },
  serialCol: { alignItems: 'flex-end' },
  serial: { fontSize: 10, letterSpacing: 0.4, marginBottom: 5 },
});
