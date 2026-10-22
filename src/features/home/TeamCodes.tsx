import { View, type ViewStyle } from 'react-native';

import { Txt } from '@/components';
import { getTeam } from '@/data';
import { fonts } from '@/theme';

const CHAR_EM = 0.56; // Big Shoulders Black, average cap width (generous, so codes like "RCB-W" never clip)
const JOIN_W = 40; // the small word between the codes, with its margins

/** Font size that fits both codes and the joining word into `width`, clamped to [min, max]. */
export function codeSize(teamIds: [string, string], width: number, max: number, min = 22): number {
  const chars = getTeam(teamIds[0]).short.length + getTeam(teamIds[1]).short.length;
  return Math.round(Math.max(min, Math.min(max, (width - JOIN_W) / (chars * CHAR_EM))));
}

/**
 * The two team codes printed in stadium type with a small "vs" between them, sized to fit the ticket.
 * Pass `winnerId` to print it as a result: "SRH beat KKR", the beaten side faded.
 */
export function TeamCodes({
  teamIds,
  width,
  max = 42,
  min,
  winnerId,
  style,
}: {
  teamIds: [string, string];
  width: number;
  max?: number;
  min?: number;
  winnerId?: string | null;
  style?: ViewStyle;
}) {
  const won = winnerId && teamIds.includes(winnerId) ? winnerId : null;
  const order: [string, string] = won ? [won, teamIds[0] === won ? teamIds[1] : teamIds[0]] : teamIds;
  const [a, b] = order.map(getTeam);
  const fs = codeSize(teamIds, width, max, min);
  return (
    <View style={[{ flexDirection: 'row', alignItems: 'baseline' }, style]} accessibilityLabel={`${a.name} ${won ? 'beat' : 'versus'} ${b.name}`}>
      <Txt variant="display" numberOfLines={1} style={{ fontSize: fs, lineHeight: fs + 2, flexShrink: 1 }}>
        {a.short}
      </Txt>
      <Txt style={{ fontFamily: fonts.label, fontSize: 12.5, lineHeight: 16, marginHorizontal: 8 }}>{won ? 'beat' : 'vs'}</Txt>
      <Txt variant="display" numberOfLines={1} style={{ fontSize: fs, lineHeight: fs + 2, flexShrink: 1, opacity: won ? 0.45 : 1 }}>
        {b.short}
      </Txt>
    </View>
  );
}
