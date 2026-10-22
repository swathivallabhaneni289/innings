import { router } from 'expo-router';
import type { ReactNode } from 'react';
import { Pressable, RefreshControl, ScrollView, StyleSheet, View, type ViewStyle } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { Feather } from '@expo/vector-icons';

import { useTheme } from '@/theme';
import { Paper } from './Paper';
import { Txt } from './Txt';

/**
 * Standard screen: concrete ground, safe area, a title bar with optional back button and right actions,
 * and a scroll body with pull-to-refresh. Use `scroll={false}` for screens that manage their own lists.
 */
export function Screen({
  title,
  kicker,
  back,
  right,
  children,
  scroll = true,
  onRefresh,
  refreshing = false,
  contentStyle,
  header,
}: {
  title?: string;
  kicker?: string; // small caps line above the title
  back?: boolean;
  right?: ReactNode;
  children: ReactNode;
  scroll?: boolean;
  onRefresh?: () => void;
  refreshing?: boolean;
  contentStyle?: ViewStyle;
  header?: ReactNode; // replaces the default title bar
}) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const bar = header ?? (title || back || right ? (
    <View style={styles.bar}>
      {back ? (
        <Pressable onPress={() => (router.canGoBack() ? router.back() : router.replace('/'))} hitSlop={12} style={{ marginRight: 10 }}>
          <Feather name="arrow-left" size={22} color={t.ink} />
        </Pressable>
      ) : null}
      <View style={{ flex: 1 }}>
        {kicker ? (
          <Txt variant="label" tone="red" numberOfLines={1}>
            {kicker}
          </Txt>
        ) : null}
        {title ? (
          <Txt variant="h1" numberOfLines={2}>
            {title}
          </Txt>
        ) : null}
      </View>
      {right}
    </View>
  ) : null);
  return (
    <Paper>
      <View style={{ paddingTop: insets.top, flex: 1 }}>
        {bar}
        {scroll ? (
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + 140 }, contentStyle]}
            refreshControl={onRefresh ? <RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={t.red} /> : undefined}
            keyboardShouldPersistTaps="handled"
            showsVerticalScrollIndicator={false}>
            {children}
          </ScrollView>
        ) : (
          <View style={[{ flex: 1 }, contentStyle]}>{children}</View>
        )}
      </View>
    </Paper>
  );
}

const styles = StyleSheet.create({
  bar: { flexDirection: 'row', alignItems: 'center', paddingHorizontal: 20, paddingTop: 8, paddingBottom: 8 },
  content: { paddingHorizontal: 20, paddingTop: 4 },
});
