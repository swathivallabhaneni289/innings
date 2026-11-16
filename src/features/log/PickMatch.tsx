import { Feather } from '@expo/vector-icons';
import { useState } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Empty, QueryState, SectionHeader } from '@/components';
import { useLogSuggestions, useNow, useSearch } from '@/data';
import type { Match } from '@/domain/types';
import { fonts, useTheme } from '@/theme';
import { MatchRow } from './MatchLine';

/** Step 1: live / just-ended first, then recent fixtures; search reaches older matches and Vault classics. */
export function PickMatch({ selectedId, onPick }: { selectedId: string | null; onPick: (m: Match) => void }) {
  const t = useTheme();
  const now = useNow(60_000);
  const [q, setQ] = useState('');
  const searching = q.trim().length >= 2;
  const suggestions = useLogSuggestions();
  const search = useSearch(q);

  const rows = (list: Match[]) =>
    list.map((m) => <MatchRow key={m.id} m={m} now={now} selected={m.id === selectedId} onPress={() => onPick(m)} />);

  return (
    <View>
      <View style={[styles.search, { backgroundColor: t.card }]}>
        <Feather name="search" size={17} color={t.inkSoft} />
        <TextInput
          value={q}
          onChangeText={setQ}
          placeholder="Teams and a year, like ind vs aus 2023"
          placeholderTextColor={t.inkFaint}
          autoCorrect={false}
          autoCapitalize="none"
          returnKeyType="search"
          style={[styles.input, { color: t.ink, marginRight: q ? 0 : 14 }]}
        />
        {q ? (
          <Pressable onPress={() => setQ('')} style={styles.clear} accessibilityRole="button" accessibilityLabel="Clear search">
            <Feather name="x" size={16} color={t.inkSoft} />
          </Pressable>
        ) : null}
      </View>

      {searching ? (
        <View>
          <SectionHeader title="Search results" style={{ marginTop: 24 }} />
          <QueryState q={search} loadingLabel="Searching the archive…">
            {(data) => {
              const matches = data.matches;
              if (matches.length === 0)
                return (
                  <Empty
                    title="No match found"
                    body={`Nothing for “${q.trim()}”. Innings has recent matches, the Vault's classics, and every World Cup, Champions Trophy, Asia Cup and IPL game since 1973. Try teams and a year, like “ind vs aus 2023”.`}
                  />
                );
              return <View>{rows(matches)}</View>;
            }}
          </QueryState>
        </View>
      ) : (
        <QueryState q={suggestions} loadingLabel="Checking the fixture list…">
          {(data) => (
            <View>
              <SectionHeader title="On now and just finished" style={{ marginTop: 24 }} />
              {data.suggested.length ? (
                rows(data.suggested)
              ) : (
                <Empty title="No play today" body="Nothing live or freshly finished. Pick a recent one below, or search the Vault." />
              )}
              <SectionHeader title="Recent" />
              {data.recent.length ? rows(data.recent) : <Empty title="Quiet week" body="No recent fixtures to log." />}
            </View>
          )}
        </QueryState>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  search: { flexDirection: 'row', alignItems: 'center', paddingLeft: 14, height: 48, marginTop: 4 },
  input: { flex: 1, alignSelf: 'stretch', fontFamily: fonts.serif, fontSize: 15, marginLeft: 10, paddingVertical: 0 },
  clear: { width: 44, height: 48, alignItems: 'center', justifyContent: 'center' },
});
