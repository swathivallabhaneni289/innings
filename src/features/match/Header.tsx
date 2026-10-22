import { StyleSheet, View } from 'react-native';

import { Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam, useNow } from '@/data';
import { COMPETITIONS } from '@/domain/constants';
import type { Match } from '@/domain/types';
import { clock, dayLabel } from '@/lib/format';
import { stockFor, useTheme } from '@/theme';
import { serialFor, TicketFrame } from './TicketFrame';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/** "2 Apr 2011" from a classic's 'YYYY-MM-DD'. */
const proseDate = (iso: string) => {
  const [y, m, d] = iso.split('-').map(Number);
  return `${d} ${MONTHS[m - 1]} ${y}`;
};

/**
 * Match Page masthead, printed as the match ticket on its competition's card stock: fine print, both teams set big,
 * and a tear-off stub with the venue, the day and a serial. Classics add a Vault stamp and their headline.
 */
export function Header({ match }: { match: Match }) {
  const t = useTheme();
  const now = useNow(60_000);
  const [a, b] = match.teamIds.map(getTeam);
  const classic = match.classic;
  const competition = classic?.competitionName ?? COMPETITIONS.find((c) => c.id === match.competitionId)?.name ?? match.competitionId;
  // Classics have no real start time on record, only the day.
  const when = classic ? proseDate(classic.date) : `${dayLabel(match.startTime, now)}, ${clock(match.startTime)}`;

  return (
    <View>
      <TicketFrame
        side="bottom"
        stock={stockFor(match, t)}
        style={{ marginTop: 4 }}
        stub={
          <View style={styles.stub}>
            <View style={{ flex: 1, marginRight: 12 }}>
              <Txt variant="h3" numberOfLines={2}>
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
        <Txt variant="label" numberOfLines={2}>
          Admit one · {competition} · {match.stage}
        </Txt>

        <View style={{ marginTop: 12 }}>
          <TeamLine name={a.name} colour={a.colour} />
          <Txt variant="label" tone="soft" style={styles.versus}>
            versus
          </Txt>
          <TeamLine name={b.name} colour={b.colour} />
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

function TeamLine({ name, colour }: { name: string; colour: string }) {
  return (
    <View style={styles.team}>
      <View style={[styles.stripe, { backgroundColor: colour }]} />
      <Txt variant="display" numberOfLines={2} style={styles.teamName}>
        {name}
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  vault: { alignSelf: 'flex-start', borderWidth: 1.5, paddingHorizontal: 8, paddingVertical: 2, marginBottom: 12, transform: [{ rotate: '-2deg' }] },
  team: { flexDirection: 'row', alignItems: 'center' },
  stripe: { width: 5, alignSelf: 'stretch', marginRight: 12, marginVertical: 6 },
  teamName: { flex: 1, fontSize: 34, lineHeight: 40, textTransform: 'uppercase' },
  versus: { marginLeft: 17, marginVertical: 2 },
  stub: { flexDirection: 'row', alignItems: 'center' },
  serialCol: { alignItems: 'flex-end' },
  serial: { fontSize: 10, letterSpacing: 0.4, marginBottom: 5 },
});
