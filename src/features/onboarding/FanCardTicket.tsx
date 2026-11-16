import { StyleSheet, View, type ViewStyle } from 'react-native';

import { Avatar, Barcode, TeamBadge, TicketStub, Txt } from '@/components';
import { fonts, useTheme } from '@/theme';

/** A stable four-digit fan number from the user id, printed on the stub like a serial. */
function fanNo(seed: string) {
  let h = 2166136261;
  for (let i = 0; i < seed.length; i++) h = Math.imul(h ^ seed.charCodeAt(i), 16777619);
  return String(Math.abs(h) % 10000).padStart(4, '0');
}

/**
 * The fan card being filled in, printed as a ticket: photo or initials, name, @handle, the team's code over its
 * kit colour, and the year they joined. Updates as the fan types.
 */
export function FanCardTicket({
  userId,
  name,
  handle,
  teamIds,
  avatarColour,
  photoUri,
  memberSince,
  style,
}: {
  userId: string;
  name: string;
  handle: string;
  teamIds: string[];
  avatarColour: string;
  photoUri?: string;
  memberSince: number;
  style?: ViewStyle;
}) {
  const t = useTheme();
  const shownName = name.trim();
  return (
    <TicketStub
      stock={t.stock.international}
      height={176}
      style={style}
      main={
        <View
          style={styles.main}
          accessible
          accessibilityLabel={`Fan card preview: ${shownName || 'no name yet'}, ${handle ? `@${handle}` : 'no username yet'}, member since ${memberSince}`}>
          <Txt variant="label" numberOfLines={1}>
            Fan card · Innings
          </Txt>
          <View style={styles.id}>
            <Avatar user={{ name: shownName || '·', avatarColour, photoUri: photoUri || undefined }} size={46} />
            <View style={{ flex: 1, marginLeft: 12 }}>
              <Txt variant="h2" tone={shownName ? 'ink' : 'faint'} numberOfLines={1}>
                {shownName || 'Your name'}
              </Txt>
              <Txt variant="mono" tone={handle ? 'soft' : 'faint'} numberOfLines={1}>
                @{handle || 'username'}
              </Txt>
            </View>
          </View>
          <View style={styles.footer}>
            {teamIds.length > 0 ? (
              <View style={styles.teams}>
                <TeamBadge teamId={teamIds[0]} size="md" />
                {teamIds.slice(1, 3).map((id) => (
                  <TeamBadge key={id} teamId={id} size="sm" />
                ))}
                {teamIds.length > 3 ? (
                  <Txt variant="monoSmall" tone="soft">
                    +{teamIds.length - 3}
                  </Txt>
                ) : null}
              </View>
            ) : (
              <Txt variant="bodySmall" tone="faint" style={{ fontFamily: fonts.ui }}>
                Pick a team
              </Txt>
            )}
            <Txt variant="monoSmall" numberOfLines={1} style={{ flexShrink: 1 }}>
              Member since {memberSince}
            </Txt>
          </View>
        </View>
      }
      side={
        <View style={styles.side}>
          <Txt variant="label" tone="red" style={styles.serial}>
            No. {fanNo(userId)}
          </Txt>
          <Barcode seed={handle || userId} width={42} height={58} />
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  main: { flex: 1, justifyContent: 'space-between' },
  id: { flexDirection: 'row', alignItems: 'center' },
  footer: { flexDirection: 'row', alignItems: 'flex-end', justifyContent: 'space-between', gap: 8 },
  teams: { flexDirection: 'row', alignItems: 'flex-end', gap: 8 },
  side: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  serial: { fontSize: 9.5, letterSpacing: 0.4 },
});
