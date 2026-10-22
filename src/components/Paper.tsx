import type { ReactNode } from 'react';
import { StyleSheet, View, type ViewStyle } from 'react-native';

import { useTheme } from '@/theme';

/** The concrete ground every screen sits on. Flat on purpose: the tickets carry the colour. */
export function Paper({ children, style }: { children?: ReactNode; style?: ViewStyle; lines?: boolean }) {
  const t = useTheme();
  return <View style={[styles.fill, { backgroundColor: t.paper }, style]}>{children}</View>;
}

const styles = StyleSheet.create({ fill: { flex: 1 } });
