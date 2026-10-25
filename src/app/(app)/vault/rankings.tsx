import { useState, type ReactNode } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';

import { Chip, Empty, Panel, QueryState, Screen, Txt } from '@/components';
import { useRankings, useVault } from '@/data';
import type { Format } from '@/domain/types';
import { RankRow } from '@/features/vault/RankRow';
import { fonts } from '@/theme';

const FORMATS: { id: Format | 'all'; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'T20', label: 'T20' },
  { id: 'ODI', label: 'ODI' },
  { id: 'Test', label: 'Test' },
];

export default function RankingsScreen() {
  const [format, setFormat] = useState<Format | 'all'>('all');
  const [era, setEra] = useState<string>('all');
  const vault = useVault();
  const rankings = useRankings({ format, era });
  const eras = vault.data?.eras ?? [];
  // The spinner shows only while the fan pulls; filter changes and background refetches stay quiet.
  const [pulling, setPulling] = useState(false);

  const refresh = async () => {
    setPulling(true);
    try {
      await rankings.refetch();
    } finally {
      setPulling(false);
    }
  };

  return (
    <Screen back title="Fan rankings" onRefresh={refresh} refreshing={pulling}>
      <Txt variant="body" tone="soft" style={{ marginBottom: 18 }}>
        The highest-rated matches ever, by the fans who watched them.
      </Txt>

      <Txt variant="bodySmall" tone="soft" style={styles.filterLabel}>
        Format
      </Txt>
      <FilterRow>
        {FORMATS.map((f) => (
          <Chip key={f.id} label={f.label} selected={format === f.id} onPress={() => setFormat(f.id)} />
        ))}
      </FilterRow>

      <Txt variant="bodySmall" tone="soft" style={[styles.filterLabel, { marginTop: 14 }]}>
        Era
      </Txt>
      <FilterRow>
        <Chip label="All" selected={era === 'all'} onPress={() => setEra('all')} />
        {eras.map((e) => (
          <Chip key={e} label={e} selected={era === e} onPress={() => setEra(e)} />
        ))}
      </FilterRow>

      <View style={{ height: 24 }} />

      <QueryState q={rankings} loadingLabel="Counting the votes…">
        {(r) =>
          r.matches.length === 0 ? (
            <Panel>
              <Empty title={era === 'all' ? 'Nothing rated here yet' : 'Nothing rated in this era yet'} body={emptyBody(format, era)} />
            </Panel>
          ) : (
            <Panel style={{ paddingVertical: 2 }}>
              {r.matches.map((item, i) => (
                <RankRow key={item.match.id} rank={i + 1} item={item} last={i === r.matches.length - 1} />
              ))}
            </Panel>
          )
        }
      </QueryState>
    </Screen>
  );
}

/** "No fan-rated Test matches from the 1970s yet. Try All." — reads cleanly for every filter combination. */
function emptyBody(format: Format | 'all', era: string) {
  const kind = format === 'all' ? 'matches' : `${format} matches`;
  const when = era === 'all' ? '' : ` from the ${era}`;
  const nudge = era !== 'all' ? ' Try another era, or All.' : format !== 'all' ? ' Try All formats.' : '';
  return `No fan-rated ${kind}${when} yet.${nudge}`;
}

function FilterRow({ children }: { children: ReactNode }) {
  return (
    <ScrollView horizontal showsHorizontalScrollIndicator={false} style={{ marginHorizontal: -20 }} contentContainerStyle={{ paddingHorizontal: 20 }}>
      {children}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  filterLabel: { fontFamily: fonts.label, marginBottom: 8 },
});
