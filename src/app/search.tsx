import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { useEffect, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, TextInput, useWindowDimensions, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Avatar, Chip, Empty, LiveTag, QueryState, Screen, SectionHeader, TicketStub, Txt } from '@/components';
import { useSearch, type SearchResponse } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { Match, User } from '@/domain/types';
import { TeamCodes } from '@/features/home/TeamCodes';
import { ticketDate } from '@/lib/format';
import { fonts, stockFor, useTheme } from '@/theme';

const SUGGESTIONS = ['IND', 'CSK', 'Wankhede', 'World Cup', '2011'];
const STUB = 70;

export default function SearchScreen() {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const [text, setText] = useState('');
  const [query, setQuery] = useState('');
  useEffect(() => {
    const h = setTimeout(() => setQuery(text.trim()), 250);
    return () => clearTimeout(h);
  }, [text]);
  const res = useSearch(query);
  const typed = text.trim();

  return (
    <Screen back title="Search" scroll={false}>
      <View style={styles.inputWrap}>
        <View style={[styles.field, { backgroundColor: t.card }]}>
          <Feather name="search" size={18} color={t.inkSoft} />
          <TextInput
            autoFocus
            value={text}
            onChangeText={setText}
            placeholder="Teams, grounds, cities, years…"
            placeholderTextColor={t.inkFaint}
            returnKeyType="search"
            autoCapitalize="none"
            autoCorrect={false}
            clearButtonMode="while-editing"
            selectionColor={t.red}
            style={[styles.input, { color: t.ink }]}
          />
        </View>
      </View>
      <ScrollView
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="on-drag"
        contentContainerStyle={{ paddingHorizontal: 20, paddingBottom: insets.bottom + 40 }}
        showsVerticalScrollIndicator={false}>
        {typed.length < 2 ? (
          <View style={{ marginTop: 18 }}>
            <Txt variant="body" tone="soft">
              Two letters and we&apos;re off. Teams, grounds, cities, competitions and years all work.
            </Txt>
            <View style={styles.suggest}>
              {SUGGESTIONS.map((s) => (
                <Chip key={s} label={s} onPress={() => setText(s)} style={{ marginBottom: 8 }} />
              ))}
            </View>
          </View>
        ) : (
          <QueryState q={res} loadingLabel="Searching the stands…">
            {(d) => <Results data={d} query={query} />}
          </QueryState>
        )}
      </ScrollView>
    </Screen>
  );
}

function Results({ data, query }: { data: SearchResponse; query: string }) {
  if (data.matches.length === 0 && data.users.length === 0) {
    return <Empty title="No tickets for that" body={`No matches or fans for “${query}”. Try a team or a ground.`} />;
  }
  return (
    <View>
      {data.matches.length > 0 ? (
        <>
          <SectionHeader title="Matches" style={{ marginTop: 18 }} />
          {data.matches.map((m) => (
            <MatchRow key={m.id} match={m} />
          ))}
        </>
      ) : null}
      {data.users.length > 0 ? (
        <>
          <SectionHeader title="Fans" style={data.matches.length === 0 ? { marginTop: 18 } : undefined} />
          <FansCard users={data.users} />
        </>
      ) : null}
    </View>
  );
}

/** A search hit printed as a small ticket on its competition's stock, with the date on the stub. */
function MatchRow({ match: m }: { match: Match }) {
  const t = useTheme();
  const { width } = useWindowDimensions();
  const [day, mon, year] = ticketDate(m).split(' ');
  return (
    <Pressable
      onPress={() => router.push('/match/' + m.id)}
      accessibilityRole="button"
      style={({ pressed }) => ({ marginBottom: 10, opacity: pressed ? 0.88 : 1 })}>
      <TicketStub
        stock={stockFor(m, t)}
        height={108}
        sideSize={STUB}
        main={
          <View style={{ flex: 1, justifyContent: 'space-between' }}>
            <View style={styles.teams}>
              <Txt variant="label" numberOfLines={1} style={{ fontSize: 9.5, flex: 1, marginRight: 6 }}>
                {m.classic ? m.classic.competitionName : competitionShort(m.competitionId)} · {m.stage}
              </Txt>
              {m.state === 'live' ? <LiveTag /> : null}
            </View>
            <TeamCodes teamIds={m.teamIds} width={width - 40 - STUB - 28} max={30} min={20} />
            <Txt variant="bodySmall" numberOfLines={1}>
              {m.classic?.headline ?? m.result?.text ?? `${m.venue.name}, ${m.venue.city}`}
            </Txt>
          </View>
        }
        side={
          <View style={{ alignItems: 'center' }}>
            <Txt variant="label" tone="red" style={{ fontSize: 10, letterSpacing: 0.6 }}>
              {day} {mon}
            </Txt>
            <Txt variant="mono" tone="red" style={{ fontSize: 13, marginTop: 2 }}>
              {year}
            </Txt>
          </View>
        }
      />
    </Pressable>
  );
}

function FansCard({ users }: { users: User[] }) {
  const t = useTheme();
  return (
    <View style={{ backgroundColor: t.card, paddingHorizontal: 14 }}>
      {users.map((u, i) => (
        <FanRow key={u.id} user={u} first={i === 0} />
      ))}
    </View>
  );
}

function FanRow({ user, first }: { user: User; first: boolean }) {
  const t = useTheme();
  return (
    <Pressable
      onPress={() => router.push('/user/' + user.id)}
      accessibilityRole="button"
      style={({ pressed }) => [styles.row, !first ? { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule } : null, { opacity: pressed ? 0.7 : 1 }]}>
      <Avatar user={user} size={34} />
      <View style={{ flex: 1, marginLeft: 12 }}>
        <Txt variant="body" numberOfLines={1} style={{ fontFamily: fonts.uiBold }}>
          {user.name}
        </Txt>
        <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
          @{user.handle}
          {user.bio ? ` · ${user.bio}` : ''}
        </Txt>
      </View>
      <Feather name="chevron-right" size={16} color={t.inkFaint} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  inputWrap: { paddingHorizontal: 20, paddingBottom: 4 },
  field: { flexDirection: 'row', alignItems: 'center', minHeight: 50, paddingHorizontal: 14 },
  input: { flex: 1, marginLeft: 10, fontFamily: fonts.serif, fontSize: 17, paddingVertical: 12 },
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 12, minHeight: 58 },
  teams: { flexDirection: 'row', alignItems: 'center' },
  suggest: { flexDirection: 'row', flexWrap: 'wrap', marginTop: 14 },
});
