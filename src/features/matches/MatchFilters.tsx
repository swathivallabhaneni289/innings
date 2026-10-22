import { ScrollView, StyleSheet, View } from 'react-native';

import { Chip } from '@/components';
import type { MatchFilter } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { CompetitionId } from '@/domain/types';
import { useTheme } from '@/theme';

export type StateFilter = NonNullable<MatchFilter['state']>;

export const STATE_CHIPS: { id: StateFilter; label: string }[] = [
  { id: 'live', label: 'Live' },
  { id: 'upcoming', label: 'Upcoming' },
  { id: 'recent', label: 'Recent' },
];

export interface ScopeChip {
  id: string;
  label: string;
  phrase: string; // used in empty states: "Nothing live in the IPL"
  filter: Pick<MatchFilter, 'competition' | 'team'>;
}

const comp = (id: CompetitionId, phrase: string): ScopeChip => ({
  id,
  label: competitionShort(id),
  phrase,
  filter: { competition: id },
});

/** India is a team filter (men and women), the rest are competitions. */
export const SCOPE_CHIPS: ScopeChip[] = [
  { id: 'india', label: 'India', phrase: 'for India', filter: { team: 'ind,ind-w' } },
  comp('ipl', 'in the IPL'),
  comp('wpl', 'in the WPL'),
  comp('world-cup', 'at the World Cup'),
  comp('bbl', 'in the BBL'),
  comp('sa20', 'in the SA20'),
  comp('the-hundred', 'in The Hundred'),
];

/** One horizontal strip: Live · Upcoming · Recent | India · IPL · WPL · … Each group is single-select; tap again to clear. */
export function MatchFilters({
  state,
  scope,
  onState,
  onScope,
}: {
  state: StateFilter | null;
  scope: string | null;
  onState: (s: StateFilter | null) => void;
  onScope: (id: string | null) => void;
}) {
  const t = useTheme();
  return (
    <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.strip}>
      {STATE_CHIPS.map((c) => (
        <Chip
          key={c.id}
          label={c.label}
          live={c.id === 'live'}
          selected={state === c.id}
          onPress={() => onState(state === c.id ? null : c.id)}
        />
      ))}
      <View style={[styles.divider, { backgroundColor: t.ink }]} />
      {SCOPE_CHIPS.map((c) => (
        <Chip key={c.id} label={c.label} selected={scope === c.id} onPress={() => onScope(scope === c.id ? null : c.id)} />
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  strip: { paddingHorizontal: 20, paddingVertical: 10, alignItems: 'center' },
  divider: { width: 1.5, height: 22, marginLeft: 2, marginRight: 10, opacity: 0.5 },
});
