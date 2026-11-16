import * as Haptics from 'expo-haptics';
import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { TeamBadge, Txt } from '@/components';
import { getTeam } from '@/data';
import { FAV_TEAMS } from '@/features/profile/constants';
import { fonts, useTheme } from '@/theme';
import { listLine, makeMain, MAX_TEAMS, toggleTeam } from './teams';

/**
 * Every team the fan backs, up to five: international sides by name, IPL sides by code. A picked chip carries its
 * order like a ticket serial, and the first is the main team, the one on the fan card. The "Backing" line shows the
 * picks as badges; tapping one makes it the main team. Shared by sign-up and Settings. `saved` is the list on file: a
 * team that isn't on the fixed lists still gets a chip.
 */
export function TeamChips({ value, onChange, saved = [] }: { value: string[]; onChange: (teamIds: string[]) => void; saved?: string[] }) {
  const t = useTheme();
  const [full, setFull] = useState(false);
  const listed = [...FAV_TEAMS.international, ...FAV_TEAMS.franchise] as string[];
  const international = [...saved.filter((id) => !listed.includes(id)), ...FAV_TEAMS.international];

  const toggle = (id: string) => {
    const next = toggleTeam(value, id);
    setFull(next.full);
    if (next.full) {
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Warning);
      return;
    }
    Haptics.selectionAsync();
    onChange(next.ids);
  };

  const chip = (id: string, label: string) => {
    const at = value.indexOf(id);
    const picked = at >= 0;
    return (
      <Pressable
        key={id}
        onPress={() => toggle(id)}
        accessibilityRole="button"
        accessibilityState={{ selected: picked }}
        accessibilityLabel={picked ? `${getTeam(id).name}, team ${at + 1}${at === 0 ? ', main team' : ''}` : getTeam(id).name}
        style={({ pressed }) => [styles.chip, { borderColor: picked ? t.ink : t.rule, backgroundColor: picked ? t.ink : t.card, opacity: pressed ? 0.75 : 1 }]}>
        {picked ? (
          <Txt variant="mono" style={[styles.serial, { color: t.red }]}>
            {at + 1}
          </Txt>
        ) : null}
        <Txt style={[styles.chipLabel, { color: picked ? t.onInk : t.ink }]}>{label}</Txt>
      </Pressable>
    );
  };

  return (
    <View>
      <Txt variant="bodySmall" tone="soft" style={{ marginBottom: 12 }}>
        Pick every side you back, up to {MAX_TEAMS}. Your first pick is your main team.
      </Txt>
      <Txt variant="bodySmall" tone="soft" style={styles.groupLabel}>
        International
      </Txt>
      <View style={styles.chips}>{international.map((id) => chip(id, getTeam(id).name))}</View>
      <Txt variant="bodySmall" tone="soft" style={[styles.groupLabel, { marginTop: 6 }]}>
        IPL
      </Txt>
      <View style={styles.chips}>{FAV_TEAMS.franchise.map((id) => chip(id, getTeam(id).short))}</View>

      <View style={[styles.backing, { borderTopColor: t.rule }]} accessibilityLiveRegion="polite">
        <Txt variant="bodySmall" tone="soft" style={{ marginRight: 4 }}>
          Backing
        </Txt>
        {value.length === 0 ? (
          <Txt variant="bodySmall" tone="faint" style={{ marginLeft: 6 }}>
            nobody yet
          </Txt>
        ) : (
          <View style={styles.badges} accessibilityLabel={`Backing ${listLine(value.map((id) => getTeam(id).name))}`}>
            {value.map((id, i) => (
              <Pressable
                key={id}
                onPress={() => {
                  if (i === 0) return;
                  Haptics.selectionAsync();
                  onChange(makeMain(value, id));
                }}
                disabled={i === 0}
                hitSlop={6}
                accessibilityRole="button"
                accessibilityLabel={i === 0 ? `${getTeam(id).name}, main team` : `Make ${getTeam(id).name} your main team`}
                style={styles.badge}>
                <TeamBadge teamId={id} size="sm" />
                {i === 0 ? (
                  <Txt variant="label" tone="red" style={styles.main}>
                    Main
                  </Txt>
                ) : null}
              </Pressable>
            ))}
          </View>
        )}
      </View>
      {full ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
          That&apos;s {MAX_TEAMS}. Tap one you&apos;ve picked to make room.
        </Txt>
      ) : value.length > 1 ? (
        <Txt variant="monoSmall" tone="faint" style={{ marginTop: 8 }}>
          Tap a badge to make it your main team.
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  groupLabel: { fontFamily: fonts.label, marginBottom: 8 },
  chips: { flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginBottom: 8 },
  chip: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, paddingHorizontal: 12, minHeight: 40 },
  chipLabel: { fontFamily: fonts.label, fontSize: 13 },
  serial: { fontSize: 12, marginRight: 7 },
  backing: { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', marginTop: 8, paddingTop: 12, borderTopWidth: StyleSheet.hairlineWidth },
  badges: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'flex-end', gap: 14, marginLeft: 8, flex: 1 },
  badge: { flexDirection: 'row', alignItems: 'flex-end' },
  main: { marginLeft: 5, marginBottom: 2, fontSize: 9 },
});
