import { Feather } from '@expo/vector-icons';
import { Pressable, StyleSheet, View } from 'react-native';

import { Txt } from '@/components';
import { fonts, useTheme } from '@/theme';
import { STEP_SHORT, STEP_TITLES, type StepId } from './draft';

/**
 * The flow's title bar: a quiet "Step 2 of 3", the step title, a plain close button, and a slim segmented
 * progress bar with one segment per step. Finished segments can be tapped to go back.
 */
export function LogHeader({
  steps,
  current,
  onClose,
  onJump,
  title,
}: {
  steps: StepId[] | null; // null while the match is still loading (step count unknown)
  current: StepId | 'printed';
  onClose: () => void;
  onJump: (s: StepId) => void;
  title?: string;
}) {
  const t = useTheme();
  const idx = steps && current !== 'printed' ? steps.indexOf(current) : -1;
  const kicker =
    current === 'printed' ? 'Close of play' : steps && idx >= 0 ? `Step ${idx + 1} of ${steps.length}` : 'Log a match';
  return (
    <View style={styles.wrap}>
      <View style={styles.top}>
        <View style={{ flex: 1 }}>
          <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
            {kicker}
          </Txt>
          <Txt variant="h1" numberOfLines={1} accessibilityRole="header">
            {title ?? (current === 'printed' ? 'Ticket printed' : STEP_TITLES[current])}
          </Txt>
        </View>
        <Pressable
          onPress={onClose}
          accessibilityRole="button"
          accessibilityLabel="Close"
          style={({ pressed }) => [styles.close, { opacity: pressed ? 0.5 : 1 }]}>
          <Feather name="x" size={24} color={t.ink} />
        </Pressable>
      </View>
      {steps && current !== 'printed' ? (
        <View style={styles.progress}>
          {steps.map((s, i) => {
            const done = i < idx;
            const now = i === idx;
            return (
              <Pressable
                key={s}
                disabled={!done}
                onPress={() => onJump(s)}
                hitSlop={{ top: 10, bottom: 10 }}
                accessibilityRole="button"
                accessibilityState={{ selected: now, disabled: !done }}
                accessibilityLabel={`Step ${i + 1}: ${STEP_TITLES[s]}${done ? ', done' : ''}`}
                style={styles.segment}>
                <View style={[styles.bar, { backgroundColor: done || now ? t.ink : t.rule }]} />
                <Txt
                  numberOfLines={1}
                  tone={now ? 'ink' : done ? 'soft' : 'faint'}
                  style={{ fontFamily: now ? fonts.uiBold : fonts.ui, fontSize: 12, lineHeight: 16, marginTop: 6 }}>
                  {STEP_SHORT[s]}
                </Txt>
              </Pressable>
            );
          })}
        </View>
      ) : current !== 'printed' ? (
        <View style={styles.progressSpace} /> // keeps the title still while the step count loads
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { paddingHorizontal: 20, paddingTop: 8, paddingBottom: 14 },
  top: { flexDirection: 'row', alignItems: 'center' },
  close: { width: 44, height: 44, alignItems: 'center', justifyContent: 'center', marginLeft: 8, marginRight: -10 },
  progress: { flexDirection: 'row', gap: 6, marginTop: 14 },
  segment: { flex: 1, paddingTop: 2 },
  bar: { height: 3 },
  progressSpace: { marginTop: 14, height: 27 },
});
