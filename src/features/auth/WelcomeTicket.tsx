import { StyleSheet, View, type ViewStyle } from 'react-native';

import { Barcode, TicketStub, Txt } from '@/components';
import { ticketDate } from '@/lib/format';
import { useTheme } from '@/theme';

/** The ticket on the Welcome screen, printed like a match stub: admit one, today's date, a red serial, a barcode. */
export function WelcomeTicket({ style }: { style?: ViewStyle }) {
  const t = useTheme();
  const today = ticketDate({ startTime: new Date().toISOString() });
  return (
    <TicketStub
      stock={t.stock.test}
      height={204}
      style={style}
      main={
        <View style={styles.main} accessible accessibilityLabel="Admit one. Innings. Season ticket.">
          <Txt variant="label" numberOfLines={1}>
            Admit one · Innings
          </Txt>
          <View>
            <Txt variant="display" numberOfLines={1} adjustsFontSizeToFit style={styles.big}>
              SEASON
            </Txt>
            <Txt variant="display" numberOfLines={1} adjustsFontSizeToFit style={styles.big}>
              TICKET
            </Txt>
            <Txt variant="bodySmall" tone="soft" numberOfLines={1} style={{ marginTop: 4 }}>
              Any ground, any sofa, any hostel room
            </Txt>
          </View>
          <View style={styles.footer}>
            <Txt variant="monoSmall" numberOfLines={1}>
              {today}
            </Txt>
            <Txt variant="monoSmall" tone="soft" numberOfLines={1}>
              Seat: yours
            </Txt>
          </View>
        </View>
      }
      side={
        <View style={styles.side}>
          <Txt variant="label" tone="red" style={styles.serial}>
            No. 0001
          </Txt>
          <Barcode seed="innings-admit-one" width={46} height={70} />
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  main: { flex: 1, justifyContent: 'space-between' },
  big: { fontSize: 50, lineHeight: 48 },
  footer: { flexDirection: 'row', justifyContent: 'space-between', gap: 8 },
  side: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  serial: { fontSize: 9.5, letterSpacing: 0.4 },
});
