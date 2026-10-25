import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { router, Tabs } from 'expo-router';
import type { ComponentProps } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { CricketBall, Txt } from '@/components';
import { fonts, useTheme } from '@/theme';

type BottomTabBarProps = Parameters<NonNullable<ComponentProps<typeof Tabs>['tabBar']>>[0];

const ICONS: Record<string, keyof typeof Feather.glyphMap> = {
  index: 'home',
  matches: 'calendar',
  vault: 'archive',
  profile: 'user',
};
const LABELS: Record<string, string> = { index: 'Home', matches: 'Matches', vault: 'Vault', profile: 'Profile' };

/** Bottom bar: Home · Matches · [big red ball: Log] · Vault · Profile, printed on paper over a double rule. */
function TabBar({ state, navigation }: BottomTabBarProps) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  return (
    <View style={[styles.bar, { paddingBottom: Math.max(insets.bottom, 10), backgroundColor: t.card, borderTopColor: t.rule }]}>
      {state.routes.map((route, i) => {
        const focused = state.index === i;
        if (route.name === 'log') {
          return (
            <View key={route.key} style={styles.item}>
              <Pressable
                accessibilityRole="button"
                accessibilityLabel="Log a match"
                onPress={() => {
                  Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
                  router.push('/log');
                }}
                style={({ pressed }) => [styles.logButton, { backgroundColor: t.card, transform: [{ scale: pressed ? 0.94 : 1 }] }]}>
                <CricketBall size={58} />
                <View style={[styles.plus, { backgroundColor: t.ink, borderColor: t.card }]}>
                  <Feather name="plus" size={13} color={t.onInk} />
                </View>
              </Pressable>
              <Txt style={[styles.label, { fontFamily: fonts.uiBold, color: t.red }]}>Log</Txt>
            </View>
          );
        }
        return (
          <Pressable
            key={route.key}
            accessibilityRole="button"
            accessibilityState={{ selected: focused }}
            onPress={() => {
              const event = navigation.emit({ type: 'tabPress', target: route.key, canPreventDefault: true });
              if (!focused && !event.defaultPrevented) navigation.navigate(route.name);
            }}
            style={styles.item}>
            <Feather name={ICONS[route.name] ?? 'circle'} size={22} color={focused ? t.ink : t.inkFaint} />
            <Txt style={[styles.label, { fontFamily: focused ? fonts.uiBold : fonts.ui, color: focused ? t.ink : t.inkFaint }]}>
              {LABELS[route.name] ?? route.name}
            </Txt>
          </Pressable>
        );
      })}
    </View>
  );
}

export default function TabsLayout() {
  return (
    <Tabs screenOptions={{ headerShown: false }} tabBar={(props) => <TabBar {...props} />}>
      <Tabs.Screen name="index" />
      <Tabs.Screen name="matches" />
      <Tabs.Screen name="log" />
      <Tabs.Screen name="vault" />
      <Tabs.Screen name="profile" />
    </Tabs>
  );
}

const styles = StyleSheet.create({
  bar: { flexDirection: 'row', borderTopWidth: StyleSheet.hairlineWidth, paddingTop: 8 },
  item: { flex: 1, alignItems: 'center', justifyContent: 'flex-end' },
  label: { fontSize: 11, marginTop: 3 },
  logButton: { marginTop: -30, width: 68, height: 68, borderRadius: 34, alignItems: 'center', justifyContent: 'center' },
  plus: { position: 'absolute', right: 3, bottom: 3, width: 20, height: 20, borderRadius: 10, borderWidth: 2, alignItems: 'center', justifyContent: 'center' },
});
