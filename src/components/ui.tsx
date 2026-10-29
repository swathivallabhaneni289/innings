// Small design-system pieces: buttons, chips, section headers, rules, avatars, team badges, split bars, stats.
import { useState, type ReactNode } from 'react';
import { ActivityIndicator, Image, Pressable, StyleSheet, View, type ViewStyle } from 'react-native';

import { getTeam } from '@/data';
import type { User } from '@/domain/types';
import { pct } from '@/lib/format';
import { fonts, useTheme } from '@/theme';
import { Txt } from './Txt';

// ---------- Button ----------

/** Primary = solid ink (red is kept for the Log ball, stamps and live); outline; ghost = red text; success = green. */
export function Button({
  label,
  onPress,
  kind = 'primary',
  size = 'md',
  icon,
  disabled,
  loading,
  style,
  maxFontSizeMultiplier,
}: {
  label: string;
  onPress?: () => void;
  kind?: 'primary' | 'outline' | 'ghost' | 'success';
  size?: 'sm' | 'md' | 'lg';
  icon?: ReactNode;
  disabled?: boolean;
  loading?: boolean;
  style?: ViewStyle;
  /** Caps how far the label grows with the system text size, for a screen that has to fit without scrolling. */
  maxFontSizeMultiplier?: number;
}) {
  const t = useTheme();
  const bg = kind === 'primary' ? t.ink : kind === 'success' ? t.green : 'transparent';
  const fg = kind === 'primary' ? t.onInk : kind === 'success' ? t.onRed : kind === 'outline' ? t.ink : t.red;
  const pad = { sm: [7, 12], md: [12, 18], lg: [16, 22] }[size];
  return (
    <Pressable
      accessibilityRole="button"
      disabled={disabled || loading}
      onPress={onPress}
      style={({ pressed }) => [
        styles.button,
        {
          backgroundColor: bg,
          borderColor: kind === 'outline' ? t.ink : 'transparent',
          paddingVertical: pad[0],
          paddingHorizontal: kind === 'ghost' ? 4 : pad[1],
          opacity: disabled ? 0.35 : pressed ? 0.78 : 1,
        },
        style,
      ]}>
      {loading ? <ActivityIndicator size="small" color={fg} /> : icon}
      <Txt
        maxFontSizeMultiplier={maxFontSizeMultiplier}
        style={{ fontFamily: fonts.heading, color: fg, fontSize: size === 'lg' ? 16 : size === 'sm' ? 13 : 15, marginLeft: icon || loading ? 7 : 0 }}>
        {label}
      </Txt>
    </Pressable>
  );
}

// ---------- Chip ----------

/** Filter chip: square-cut, outlined; selected = solid ink. */
export function Chip({
  label,
  selected,
  onPress,
  live,
  style,
}: {
  label: string;
  selected?: boolean;
  onPress?: () => void;
  live?: boolean;
  style?: ViewStyle;
}) {
  const t = useTheme();
  return (
    <Pressable
      onPress={onPress}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      style={[styles.chip, { borderColor: selected ? t.ink : t.rule, backgroundColor: selected ? t.ink : t.card }, style]}>
      {live ? <View style={[styles.dot, { backgroundColor: t.red }]} /> : null}
      <Txt style={{ fontFamily: fonts.label, fontSize: 13, color: selected ? t.onInk : t.ink }}>{label}</Txt>
    </Pressable>
  );
}

// ---------- Section header ----------

/** Section title in bold grotesk, with an optional red action on the right. */
export function SectionHeader({ title, action, onAction, style }: { title: string; action?: string; onAction?: () => void; style?: ViewStyle }) {
  return (
    <View style={[styles.headerRow, { marginTop: 28, marginBottom: 12 }, style]}>
      <Txt variant="h2">{title}</Txt>
      {action ? (
        <Pressable onPress={onAction} hitSlop={12}>
          <Txt style={{ fontFamily: fonts.uiBold, fontSize: 13 }} tone="red">
            {action}
          </Txt>
        </Pressable>
      ) : null}
    </View>
  );
}

export function Rule({ kind = 'hair', style }: { kind?: 'hair' | 'double' | 'perforated'; style?: ViewStyle }) {
  const t = useTheme();
  if (kind === 'double') return <View style={[{ height: 2, backgroundColor: t.ink }, style]} />;
  if (kind === 'perforated') return <View style={[{ borderTopWidth: 1.6, borderStyle: 'dotted', borderColor: t.cardEdge }, style]} />;
  return <View style={[{ height: StyleSheet.hairlineWidth, backgroundColor: t.rule }, style]} />;
}

// ---------- Avatar ----------

export function Avatar({ user, size = 32 }: { user: Pick<User, 'name' | 'avatarColour' | 'photoUri'>; size?: number }) {
  const [failed, setFailed] = useState(false);
  const initials = user.name
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();
  if (user.photoUri && !failed)
    return <Image source={{ uri: user.photoUri }} onError={() => setFailed(true)} style={{ width: size, height: size, borderRadius: size / 2 }} />;
  return (
    <View style={{ width: size, height: size, borderRadius: size / 2, backgroundColor: user.avatarColour, alignItems: 'center', justifyContent: 'center' }}>
      {/* The initials are sized to the circle, so they don't grow with the system text size. */}
      <Txt allowFontScaling={false} style={{ fontFamily: fonts.uiBold, fontSize: size * 0.36, color: '#FFFFFF' }}>
        {initials}
      </Txt>
    </View>
  );
}

// ---------- Team badge ----------

/** Team code printed in stadium type, with a kit-colour band under it. */
export function TeamBadge({ teamId, size = 'md' }: { teamId: string; size?: 'sm' | 'md' | 'lg' }) {
  const team = getTeam(teamId);
  const fs = { sm: 15, md: 18, lg: 24 }[size];
  return (
    <View style={{ alignSelf: 'flex-start' }}>
      <Txt variant="score" style={{ fontSize: fs, lineHeight: fs * 1.05 }}>
        {team.short}
      </Txt>
      <View style={{ height: 3, backgroundColor: team.colour, marginTop: 1 }} />
    </View>
  );
}

// ---------- Split bar ----------

/**
 * Two-team share bar, e.g. friends' predictions. Shares are 0..1 keyed by team id.
 * Ink and faint ink only (red is kept for live and stamps); each team code carries its kit-colour underline.
 */
export function SplitBar({ teamIds, shares, label }: { teamIds: [string, string]; shares: Record<string, number>; label?: string }) {
  const t = useTheme();
  const a = shares[teamIds[0]] ?? 0;
  const b = shares[teamIds[1]] ?? 0;
  const total = a + b;
  const left = total > 0 ? a / total : 0.5;
  const leftPct = total > 0 ? pct(left) : '–';
  const rightPct = total > 0 ? pct(1 - left) : '–';
  return (
    <View accessible accessibilityLabel={`${label ? `${label}. ` : ''}${getTeam(teamIds[0]).short} ${leftPct}, ${getTeam(teamIds[1]).short} ${rightPct}`}>
      {label ? (
        <Txt variant="bodySmall" tone="soft" style={{ marginBottom: 6 }}>
          {label}
        </Txt>
      ) : null}
      <View style={styles.splitLabels}>
        <View style={styles.splitSide}>
          <View>
            <TeamBadge teamId={teamIds[0]} size="sm" />
          </View>
          <Txt variant="mono" style={{ marginLeft: 8, marginBottom: 4 }}>
            {leftPct}
          </Txt>
        </View>
        <View style={styles.splitSide}>
          <Txt variant="mono" style={{ marginRight: 8, marginBottom: 4 }}>
            {rightPct}
          </Txt>
          <View>
            <TeamBadge teamId={teamIds[1]} size="sm" />
          </View>
        </View>
      </View>
      {total > 0 ? (
        <View style={styles.splitTrack}>
          <View style={{ flex: Math.max(left, 0.001), backgroundColor: t.ink }} />
          <View style={{ width: 3 }} />
          <View style={{ flex: Math.max(1 - left, 0.001), backgroundColor: t.inkFaint }} />
        </View>
      ) : (
        // No picks yet: an empty track, so it doesn't read as a 50/50 split.
        <View style={[styles.splitTrack, { borderWidth: 1, borderColor: t.rule }]} />
      )}
    </View>
  );
}

// ---------- Stat ----------

/** Big printed number over a fine-print label. `extra` sits beside the number (e.g. a ball rating). */
export function Stat({ value, label, tone, extra }: { value: string; label: string; tone?: 'ink' | 'green' | 'red'; extra?: ReactNode }) {
  return (
    <View style={{ alignItems: 'flex-start' }}>
      <View style={{ flexDirection: 'row', alignItems: 'center' }}>
        <Txt variant="stat" tone={tone ?? 'ink'}>
          {value}
        </Txt>
        {extra ? <View style={{ marginLeft: 6 }}>{extra}</View> : null}
      </View>
      <Txt variant="label" tone="soft" style={{ fontSize: 9.5, marginTop: 2 }}>
        {label}
      </Txt>
    </View>
  );
}

export function LiveTag() {
  const t = useTheme();
  return (
    <View style={styles.live}>
      <View style={[styles.dot, { backgroundColor: t.red }]} />
      <Txt variant="label" tone="red" style={{ fontSize: 10 }}>
        Live
      </Txt>
    </View>
  );
}

/** A flat white card. `accent` prints a thin coloured band along the top edge. */
export function Panel({ children, style, accent }: { children: ReactNode; style?: ViewStyle; accent?: 'red' | 'green' }) {
  const t = useTheme();
  return (
    <View
      style={[
        styles.panel,
        { backgroundColor: t.card },
        accent ? { borderTopWidth: 3, borderTopColor: accent === 'red' ? t.red : t.green } : null,
        style,
      ]}>
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  button: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center', borderWidth: 1.5 },
  chip: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, paddingHorizontal: 14, minHeight: 40, marginRight: 8 },
  dot: { width: 6, height: 6, borderRadius: 3, marginRight: 6 },
  headerRow: { flexDirection: 'row', alignItems: 'baseline', justifyContent: 'space-between' },
  splitLabels: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: 8 },
  splitSide: { flexDirection: 'row', alignItems: 'center' },
  splitTrack: { flexDirection: 'row', height: 8 },
  live: { flexDirection: 'row', alignItems: 'center' },
  panel: { padding: 16 },
});
