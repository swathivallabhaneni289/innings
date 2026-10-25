import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { useMemo, useState } from 'react';
import { Pressable, SectionList, StyleSheet, TextInput, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Empty, Screen, Txt } from '@/components';
import { COUNTRIES, matchesCountry, type Country } from '@/features/auth/countries';
import { useSignInFlow } from '@/features/auth/flow';
import { fonts, useTheme } from '@/theme';

/** The country list for the phone screen: cricket nations first, then everyone else A–Z, with a search box. */
export default function CountryScreen() {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const current = useSignInFlow((s) => s.country);
  const setCountry = useSignInFlow((s) => s.setCountry);
  const [query, setQuery] = useState('');

  const sections = useMemo(() => {
    const cricket = COUNTRIES.cricket.filter((c) => matchesCountry(c, query));
    const others = COUNTRIES.others.filter((c) => matchesCountry(c, query));
    return [
      { title: 'Cricket nations', data: cricket },
      { title: query.trim() ? 'Everywhere else' : 'Everywhere else, A–Z', data: others },
    ].filter((s) => s.data.length > 0);
  }, [query]);

  const pick = (c: Country) => {
    setCountry(c.code);
    router.back();
  };

  const close = (
    <Pressable onPress={() => router.back()} hitSlop={12} accessibilityRole="button" accessibilityLabel="Close">
      <Feather name="x" size={24} color={t.ink} />
    </Pressable>
  );

  return (
    <Screen title="Your country" right={close} scroll={false}>
      <View style={[styles.search, { backgroundColor: t.card, borderColor: t.rule }]}>
        <Feather name="search" size={16} color={t.inkSoft} />
        <TextInput
          value={query}
          onChangeText={setQuery}
          placeholder="Search by name or code"
          placeholderTextColor={t.inkFaint}
          selectionColor={t.red}
          autoCorrect={false}
          autoCapitalize="none"
          returnKeyType="search"
          clearButtonMode="while-editing"
          accessibilityLabel="Search countries"
          style={[styles.searchInput, { color: t.ink }]}
        />
      </View>
      <SectionList
        sections={sections}
        keyExtractor={(c) => c.code}
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="on-drag"
        stickySectionHeadersEnabled={false}
        showsVerticalScrollIndicator={false}
        initialNumToRender={24}
        contentContainerStyle={{ paddingHorizontal: 20, paddingBottom: insets.bottom + 24 }}
        renderSectionHeader={({ section }) => (
          <Txt variant="label" tone="soft" style={styles.sectionHead}>
            {section.title}
          </Txt>
        )}
        renderItem={({ item, index, section }) => {
          const on = item.code === current;
          return (
            <Pressable
              onPress={() => pick(item)}
              accessibilityRole="button"
              accessibilityState={{ selected: on }}
              accessibilityLabel={`${item.name}, ${item.dial}`}
              style={({ pressed }) => [
                styles.row,
                { backgroundColor: t.card, opacity: pressed ? 0.7 : 1 },
                index > 0 ? { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule } : null,
                index === section.data.length - 1 ? { marginBottom: 4 } : null,
              ]}>
              <Txt variant="body" numberOfLines={1} style={{ flex: 1, fontFamily: on ? fonts.uiBold : fonts.serif }}>
                {item.name}
              </Txt>
              <Txt variant="monoSmall" tone="faint" style={{ marginLeft: 10 }}>
                {item.code}
              </Txt>
              <Txt variant="mono" style={styles.dial}>
                {item.dial}
              </Txt>
              <View style={styles.tick}>{on ? <Feather name="check" size={16} color={t.ink} /> : null}</View>
            </Pressable>
          );
        }}
        ListEmptyComponent={<Empty title="No country by that name" body="Try its name, its two-letter code or its dialling code." />}
      />
    </Screen>
  );
}

const styles = StyleSheet.create({
  search: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, marginHorizontal: 20, marginTop: 4, paddingHorizontal: 12, minHeight: 46 },
  searchInput: { flex: 1, fontFamily: fonts.serif, fontSize: 16, paddingVertical: 10, marginLeft: 8 },
  sectionHead: { marginTop: 22, marginBottom: 8 },
  row: { flexDirection: 'row', alignItems: 'center', minHeight: 50, paddingHorizontal: 14 },
  dial: { width: 58, textAlign: 'right' },
  tick: { width: 24, alignItems: 'flex-end' },
});
