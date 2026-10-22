import { useState, type ReactNode } from 'react';
import { StyleSheet, View, type ViewStyle } from 'react-native';
import Svg, { Line, Path } from 'react-native-svg';

import { getTeam } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { Match } from '@/domain/types';
import { ticketDate } from '@/lib/format';
import { stockFor, useTheme } from '@/theme';
import { BallRating } from './BallRating';
import { Stamp } from './Motion';
import { Txt } from './Txt';

const NOTCH = 7;

/** A square-cut ticket with two notches at the perforation (horizontal: stub on the right; vertical: stub at the bottom). */
function ticketPath(w: number, h: number, split: number, vertical: boolean) {
  if (!vertical) {
    return (
      `M 0 0 L ${split - NOTCH} 0 A ${NOTCH} ${NOTCH} 0 0 0 ${split + NOTCH} 0 L ${w} 0 L ${w} ${h} ` +
      `L ${split + NOTCH} ${h} A ${NOTCH} ${NOTCH} 0 0 0 ${split - NOTCH} ${h} L 0 ${h} Z`
    );
  }
  return (
    `M 0 0 L ${w} 0 L ${w} ${split - NOTCH} A ${NOTCH} ${NOTCH} 0 0 0 ${w} ${split + NOTCH} L ${w} ${h} ` +
    `L 0 ${h} L 0 ${split + NOTCH} A ${NOTCH} ${NOTCH} 0 0 0 0 ${split - NOTCH} Z`
  );
}

/** Deterministic barcode bars from a seed string, so a ticket always prints the same code. */
export function Barcode({ seed, width, height, color }: { seed: string; width: number; height: number; color?: string }) {
  const t = useTheme();
  let h = 2166136261;
  for (let i = 0; i < seed.length; i++) h = Math.imul(h ^ seed.charCodeAt(i), 16777619);
  // One path for all bars: a barcode is drawn as a single shape, which keeps long ticket lists fast.
  let d = '';
  let x = 0;
  while (x < width) {
    h = Math.imul(h ^ (h >>> 13), 1274126177);
    const bw = 1 + (Math.abs(h) % 3);
    const gap = 1 + (Math.abs(h >>> 7) % 2);
    if (x + bw > width) break;
    d += `M${x} 0h${bw}v${height}h${-bw}Z`;
    x += bw + gap;
  }
  return (
    <Svg width={width} height={height} accessibilityElementsHidden>
      <Path d={d} fill={color ?? t.ink} />
    </Svg>
  );
}

/**
 * The signature element: a match ticket on coloured card stock with a perforated, notched stub.
 * `main` fills the body, `side` the tear-off stub. Horizontal tickets put the stub on the right; vertical ones
 * (grids) put it at the bottom. Pass `stock` for the card colour (default: white card).
 */
export function TicketStub({
  main,
  side,
  height,
  vertical = false,
  sideSize,
  style,
  overlay,
  stock,
}: {
  main: ReactNode;
  side?: ReactNode;
  height: number;
  vertical?: boolean;
  sideSize?: number;
  style?: ViewStyle;
  overlay?: ReactNode;
  stock?: string;
}) {
  const t = useTheme();
  const [w, setW] = useState(0);
  const sideLen = sideSize ?? (vertical ? Math.round(height * 0.26) : Math.max(70, Math.round(w * 0.22)));
  const split = vertical ? height - sideLen : w - sideLen;
  return (
    <View style={[{ height }, style]} onLayout={(e) => setW(Math.round(e.nativeEvent.layout.width))}>
      {w > 0 ? (
        <Svg width={w} height={height} style={StyleSheet.absoluteFill}>
          <Path d={ticketPath(w, height, split, vertical)} fill={stock ?? t.card} />
          {vertical ? (
            <Line x1={NOTCH + 4} y1={split} x2={w - NOTCH - 4} y2={split} stroke={t.cardEdge} strokeWidth={1.6} strokeDasharray="1.5 4" strokeLinecap="round" />
          ) : (
            <Line x1={split} y1={NOTCH + 4} x2={split} y2={height - NOTCH - 4} stroke={t.cardEdge} strokeWidth={1.6} strokeDasharray="1.5 4" strokeLinecap="round" />
          )}
        </Svg>
      ) : null}
      <View style={vertical ? styles.colLayout : styles.rowLayout}>
        <View style={vertical ? { height: split, padding: 12 } : { width: Math.max(0, split), padding: 14 }}>{main}</View>
        {side ? (
          <View style={[vertical ? { height: sideLen, paddingHorizontal: 12, paddingVertical: 8 } : { width: sideLen, paddingVertical: 12, paddingHorizontal: 8 }, styles.side]}>
            {side}
          </View>
        ) : null}
      </View>
      {overlay ? (
        <View pointerEvents="none" style={[StyleSheet.absoluteFill, styles.overlay]}>
          {overlay}
        </View>
      ) : null}
    </View>
  );
}

/**
 * A match printed as a ticket on its competition's card stock: fine print, the team codes, venue, date,
 * a ball rating and the stub serial with a barcode. `rating` is the fan's own rating on a logged stub,
 * or the average fan rating.
 */
export function MatchTicket({
  match,
  rating,
  stubNo,
  size = 'md',
  stamped = false,
  instantClassic = false,
  caption,
  style,
}: {
  match: Match;
  rating?: number | null;
  stubNo?: number;
  size?: 'lg' | 'md' | 'sm';
  stamped?: boolean;
  instantClassic?: boolean;
  caption?: string; // e.g. "📍 Hostel room, with 6 people"
  style?: ViewStyle;
}) {
  const t = useTheme();
  const [a, b] = match.teamIds.map(getTeam);
  const stock = stockFor(match, t);
  const comp = match.classic ? match.classic.competitionName : competitionShort(match.competitionId);
  const serial = stubNo ? `No. ${String(stubNo).padStart(4, '0')}` : match.format;

  if (size === 'sm') {
    return (
      <TicketStub
        vertical
        stock={stock}
        height={172}
        style={style}
        main={
          <View style={{ flex: 1, justifyContent: 'space-between' }}>
            <Txt variant="label" numberOfLines={1} style={{ fontSize: 9 }}>
              {ticketDate(match)}
            </Txt>
            <View>
              <Txt variant="display" numberOfLines={1} style={{ fontSize: 30, lineHeight: 30 }}>
                {a.short}
              </Txt>
              <Txt variant="label" tone="soft" style={{ fontSize: 9 }}>
                versus
              </Txt>
              <Txt variant="display" numberOfLines={1} style={{ fontSize: 30, lineHeight: 30 }}>
                {b.short}
              </Txt>
            </View>
            {rating != null ? <BallRating value={rating} size={11} gap={2} /> : <View />}
          </View>
        }
        side={
          <View style={styles.sideRow}>
            <Txt variant="label" tone="red" style={{ fontSize: 9, letterSpacing: 0.4 }}>
              {serial}
            </Txt>
            {instantClassic ? <Txt style={{ fontSize: 11 }}>❤️</Txt> : <Barcode seed={match.id} width={34} height={14} />}
          </View>
        }
        overlay={stamped ? <Stamp visible size="sm" /> : null}
      />
    );
  }

  const lg = size === 'lg';
  const codeSize = lg ? 54 : 42;
  return (
    <TicketStub
      stock={stock}
      height={lg ? 196 : 158}
      style={style}
      main={
        <View style={{ flex: 1, justifyContent: 'space-between' }}>
          <Txt variant="label" numberOfLines={1}>
            Admit one · {match.stage} · {comp}
          </Txt>
          <View>
            <View style={styles.teams}>
              <Txt variant="display" numberOfLines={1} style={{ fontSize: codeSize, lineHeight: codeSize * 0.98 }}>
                {a.short}
              </Txt>
              <Txt variant="h3" style={{ marginHorizontal: 8 }}>
                vs
              </Txt>
              <Txt variant="display" numberOfLines={1} style={{ fontSize: codeSize, lineHeight: codeSize * 0.98 }}>
                {b.short}
              </Txt>
            </View>
            <Txt variant="bodySmall" tone="soft" numberOfLines={1} style={{ marginTop: 2 }}>
              {caption ?? `${match.venue.name}, ${match.venue.city}`}
            </Txt>
          </View>
          <View style={styles.footer}>
            <Txt variant="monoSmall" numberOfLines={1} style={{ flexShrink: 1 }}>
              {ticketDate(match)}
            </Txt>
            {rating != null ? <BallRating value={rating} size={lg ? 15 : 12} gap={2} /> : null}
          </View>
        </View>
      }
      side={
        <View style={styles.sideCol}>
          <Txt variant="label" tone="red" style={{ fontSize: 9.5, letterSpacing: 0.4 }}>
            {serial}
          </Txt>
          {instantClassic ? <Txt style={{ fontSize: 16 }}>❤️</Txt> : null}
          <Barcode seed={match.id + (stubNo ?? '')} width={lg ? 48 : 42} height={lg ? 64 : 52} />
        </View>
      }
      overlay={stamped ? <Stamp visible size={lg ? 'lg' : 'md'} /> : null}
    />
  );
}

const styles = StyleSheet.create({
  rowLayout: { flex: 1, flexDirection: 'row' },
  colLayout: { flex: 1 },
  side: { justifyContent: 'center', alignItems: 'center' },
  sideRow: { flex: 1, flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', alignSelf: 'stretch' },
  sideCol: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  teams: { flexDirection: 'row', alignItems: 'baseline' },
  footer: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 8 },
  overlay: { alignItems: 'center', justifyContent: 'center' },
});
