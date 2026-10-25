import { StyleSheet, View } from 'react-native';

import { Chip, TeamBadge, Txt } from '@/components';
import { getTeam } from '@/data';
import { FAV_TEAMS } from '@/features/profile/constants';
import { fonts, useTheme } from '@/theme';

/**
 * Favourite-team chips (international by name, franchises by code) and a "Backing" line with the code over its
 * kit colour. Shared by onboarding and Settings. `saved` is the team on file: one that isn't on the fixed list
 * still gets a chip, and a blank one gets none.
 */
export function TeamChips({ value, onChange, saved }: { value: string; onChange: (teamId: string) => void; saved?: string }) {
  const t = useTheme();
  const listed = [...FAV_TEAMS.international, ...FAV_TEAMS.franchise] as string[];
  const international: string[] = saved && !listed.includes(saved) ? [saved, ...FAV_TEAMS.international] : [...FAV_TEAMS.international];

  return (
    <View>
      <Txt variant="bodySmall" tone="soft" style={styles.groupLabel}>
        International
      </Txt>
      <View style={styles.chips}>
        {international.map((id) => (
          <View key={id} style={styles.chipCell}>
            <Chip label={getTeam(id).name} selected={value === id} onPress={() => onChange(id)} />
          </View>
        ))}
      </View>
      <Txt variant="bodySmall" tone="soft" style={[styles.groupLabel, { marginTop: 6 }]}>
        Franchise
      </Txt>
      <View style={styles.chips}>
        {FAV_TEAMS.franchise.map((id) => (
          <View key={id} style={styles.chipCell}>
            <Chip label={getTeam(id).short} selected={value === id} onPress={() => onChange(id)} />
          </View>
        ))}
      </View>
      <View style={[styles.backing, { borderTopColor: t.rule }]}>
        <Txt variant="bodySmall" tone="soft">
          Backing
        </Txt>
        {value ? (
          <>
            <View style={{ marginLeft: 10 }}>
              <TeamBadge teamId={value} size="sm" />
            </View>
            <Txt variant="bodySmall" numberOfLines={1} style={{ marginLeft: 8, flexShrink: 1, fontFamily: fonts.uiBold }}>
              {getTeam(value).name}
            </Txt>
          </>
        ) : (
          <Txt variant="bodySmall" tone="faint" style={{ marginLeft: 8 }}>
            nobody yet
          </Txt>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  groupLabel: { fontFamily: fonts.label, marginBottom: 8 },
  chips: { flexDirection: 'row', flexWrap: 'wrap', marginBottom: 6 },
  chipCell: { marginBottom: 8 },
  backing: { flexDirection: 'row', alignItems: 'center', marginTop: 8, paddingTop: 12, borderTopWidth: StyleSheet.hairlineWidth },
});
