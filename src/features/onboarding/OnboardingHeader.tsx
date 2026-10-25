import { Feather } from '@expo/vector-icons';
import type { ReactNode } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { Txt } from '@/components';
import { fonts, useTheme } from '@/theme';

const STEPS = ['Fan card', 'Your people'];

/** Sign-up's title bar, in the Log flow's style: "Step 1 of 2", the step title, and a slim two-part progress bar. */
export function OnboardingHeader({ step, title, onBack, right }: { step: 1 | 2; title: string; onBack?: () => void; right?: ReactNode }) {
  const t = useTheme();
  return (
    <View style={styles.wrap}>
      <View style={styles.top}>
        {onBack ? (
          <Pressable onPress={onBack} hitSlop={12} accessibilityRole="button" accessibilityLabel="Back" style={styles.back}>
            <Feather name="arrow-left" size={22} color={t.ink} />
          </Pressable>
        ) : null}
        <View style={{ flex: 1 }}>
          <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
            Step {step} of {STEPS.length}
          </Txt>
          <Txt variant="h1" numberOfLines={1} accessibilityRole="header">
            {title}
          </Txt>
        </View>
        {right}
      </View>
      <View style={styles.progress}>
        {STEPS.map((label, i) => {
          const now = i === step - 1;
          const done = i < step - 1;
          return (
            <View key={label} style={styles.segment} accessibilityLabel={`Step ${i + 1}: ${label}${done ? ', done' : ''}`}>
              <View style={[styles.bar, { backgroundColor: done || now ? t.ink : t.rule }]} />
              <Txt
                numberOfLines={1}
                tone={now ? 'ink' : done ? 'soft' : 'faint'}
                style={{ fontFamily: now ? fonts.uiBold : fonts.ui, fontSize: 12, lineHeight: 16, marginTop: 6 }}>
                {label}
              </Txt>
            </View>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { paddingHorizontal: 20, paddingTop: 8, paddingBottom: 14 },
  top: { flexDirection: 'row', alignItems: 'center' },
  back: { marginRight: 12, paddingVertical: 8 },
  progress: { flexDirection: 'row', gap: 6, marginTop: 14 },
  segment: { flex: 1, paddingTop: 2 },
  bar: { height: 3 },
});
