import { router } from 'expo-router';
import { Pressable, ScrollView, StyleSheet, useWindowDimensions, View } from 'react-native';

import { CricketBall, Empty, LiveTag, TicketStub, Txt } from '@/components';
import { Barcode } from '@/components/TicketStub';
import { getTeam, useNow } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { InningsScore, Match } from '@/domain/types';
import { clock, countdown, inningsLine, isLocked } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';
import { TeamCodes } from './TeamCodes';

const HEIGHT = 152;
const STUB = 72;
const PAD = 14; // TicketStub's main padding

/** "On now": live matches first, then the rest of today's fixtures, each printed as a ticket on its competition's stock. */
export function OnNowStrip({ matches }: { matches: Match[] }) {
  const { width } = useWindowDimensions();
  if (matches.length === 0) {
    return <Empty title="No play today" body="Nothing live or on today's fixture list. The Vault is always open." />;
  }
  // One ticket fills the row; several leave the next one peeking in from the right.
  const ticketW = matches.length === 1 ? width - 40 : Math.min(320, width - 84);
  return (
    <ScrollView
      key={matches.map((m) => m.id).join(',')}
      horizontal
      showsHorizontalScrollIndicator={false}
      style={styles.strip}
      contentContainerStyle={styles.stripContent}
      snapToInterval={ticketW + 12}
      decelerationRate="fast">
      {matches.map((m) => (
        <OnNowTicket key={m.id} match={m} width={ticketW} />
      ))}
    </ScrollView>
  );
}

const runs = (i: InningsScore) => `${i.runs}${i.wickets >= 10 ? '' : `/${i.wickets}`}`;

function OnNowTicket({ match: m, width }: { match: Match; width: number }) {
  const t = useTheme();
  const now = useNow(1_000);
  const live = m.state === 'live';
  const done = m.state === 'completed';
  const locked = isLocked(m, now);
  const canPredict = !live && !done && !locked;
  const comp = competitionShort(m.competitionId);
  const codesW = width - STUB - PAD * 2;

  const batting = live ? m.live?.battingTeamId : undefined;
  const battingInnings = batting ? (m.live?.innings ?? []).filter((i) => i.teamId === batting) : [];
  const finalLine = done
    ? m.teamIds
        .map((id) => {
          const inn = (m.result?.innings ?? []).filter((i) => i.teamId === id);
          return inn.length ? `${getTeam(id).short} ${inn.map(runs).join(' & ')}` : null;
        })
        .filter(Boolean)
        .join('  ·  ')
    : '';

  return (
    <Pressable
      onPress={() => router.push('/match/' + m.id)}
      accessibilityRole="button"
      style={({ pressed }) => [{ marginRight: 12, opacity: pressed ? 0.88 : 1 }]}>
      <TicketStub
        stock={stockFor(m, t)}
        height={HEIGHT}
        sideSize={STUB}
        style={{ width }}
        main={
          <View style={styles.main}>
            <View style={styles.row}>
              <Txt variant="label" numberOfLines={1} style={styles.fine}>
                {live ? `Admit one · ${m.stage}` : `${comp} · ${m.stage}${done ? '' : ` · ${clock(m.startTime)}`}`}
              </Txt>
              {live ? <LiveTag /> : done ? <Txt variant="label" tone="soft" style={{ fontSize: 10 }}>Result</Txt> : null}
            </View>

            <TeamCodes teamIds={m.teamIds} width={codesW} max={42} />

            <View>
              {live ? (
                <>
                  {battingInnings.length ? (
                    <View style={styles.row}>
                      <CricketBall size={9} />
                      <Txt variant="mono" numberOfLines={1} style={{ fontSize: 12.5, marginLeft: 5, flexShrink: 1 }}>
                        {getTeam(batting!).short} {battingInnings.map(inningsLine).join(' & ')}
                      </Txt>
                    </View>
                  ) : null}
                  <Txt variant="bodySmall" numberOfLines={2} style={{ fontFamily: fonts.ui, marginTop: 2 }}>
                    {m.live?.statusText ?? 'In play'}
                  </Txt>
                </>
              ) : done ? (
                <>
                  {finalLine ? (
                    <Txt variant="mono" numberOfLines={1} style={{ fontSize: 12.5 }}>
                      {finalLine}
                    </Txt>
                  ) : null}
                  <Txt variant="bodySmall" numberOfLines={2} style={{ fontFamily: fonts.ui, marginTop: 2 }}>
                    {m.result?.text ?? 'Match complete'}
                  </Txt>
                </>
              ) : (
                <>
                  <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
                    {m.venue.name}, {m.venue.city}
                  </Txt>
                  <Txt variant="bodySmall" numberOfLines={1} style={{ fontFamily: fonts.ui, marginTop: 2 }}>
                    {locked ? (
                      'Waiting on the first ball'
                    ) : (
                      <>
                        Locks in{' '}
                        <Txt variant="mono" style={{ fontSize: 13 }}>
                          {countdown(m.startTime, now)}
                        </Txt>
                      </>
                    )}
                  </Txt>
                </>
              )}
            </View>
          </View>
        }
        side={
          canPredict ? (
            <View style={styles.predictWrap} accessibilityLabel="Make your call">
              <Txt style={[styles.predict, { color: t.red }]}>PREDICT</Txt>
            </View>
          ) : (
            <View style={styles.stub}>
              <Txt variant="label" tone="red" style={{ fontSize: 9.5, letterSpacing: 0.6 }}>
                {m.format}
              </Txt>
              <Barcode seed={m.id} width={40} height={50} />
              {done ? (
                <Txt variant="label" tone="soft" style={{ fontSize: 8.5, letterSpacing: 0.4 }}>
                  Full time
                </Txt>
              ) : live || locked ? (
                <Txt variant="label" tone="soft" style={{ fontSize: 8.5, letterSpacing: 0.4 }}>
                  Locked 🔒
                </Txt>
              ) : (
                <View />
              )}
            </View>
          )
        }
      />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  strip: { marginHorizontal: -20 },
  stripContent: { paddingHorizontal: 20 },
  main: { flex: 1, justifyContent: 'space-between' },
  row: { flexDirection: 'row', alignItems: 'center' },
  fine: { fontSize: 10, flex: 1, marginRight: 8 },
  stub: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  predictWrap: { flex: 1, alignSelf: 'stretch', alignItems: 'center', justifyContent: 'center' },
  predict: { width: 120, textAlign: 'center', fontFamily: fonts.heading, fontSize: 14, lineHeight: 18, letterSpacing: 2.2, transform: [{ rotate: '-90deg' }] },
});
