import { Feather } from '@expo/vector-icons';
import { Pressable, StyleSheet, View } from 'react-native';

import { Txt } from '@/components';
import { fonts, useTheme } from '@/theme';
import { STEP_SHORT, STEP_TITLES, type StepId } from './draft';

/** The whole Log flow, counted once: every log is "of 4", whichever way the fan came in. */
const FLOW: StepId[] = ['match', 'rating', 'players', 'place'];

/**
 * The flow's title bar: a quiet "Step 2 of 4", the step title, a plain close button, and a slim segmented
 * progress bar with one segment per step. Numbering runs against the full flow, so "Rate it" is always step 2,
 * whether the match was picked here or came preset from a Match Page, and the count never resets. A step that
 * doesn't apply to this match (no XI to rate) shows struck through. Finished segments can be tapped to go back.
 */
export function LogHeader({
  steps,
  current,
  onClose,
  onJump,
  title,
}: {
  steps: StepId[] | null; // the steps this log actually visits; null while the match is still loading
  current: StepId | 'printed';
  onClose: () => void;
  onJump: (s: StepId) => void;
  title?: string;
}) {
  const t = useTheme();
  const at = current === 'printed' ? -1 : FLOW.indexOf(current);
  const kicker = current === 'printed' ? 'Close of play' : at >= 0 ? `Step ${at + 1} of ${FLOW.length}` : 'Log a match';
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
      {current !== 'printed' ? (
        <View style={styles.progress}>
          {FLOW.map((s, i) => {
            // The match step is always settled once you're past it (picked here or preset); only players can be skipped.
            const skipped = s !== 'match' && !!steps && !steps.includes(s);
            const done = i < at && !skipped;
            const now = i === at;
            const canJump = done && !!steps && steps.includes(s);
            return (
              <Pressable
                key={s}
                disabled={!canJump}
                onPress={() => onJump(s)}
                hitSlop={{ top: 10, bottom: 10 }}
                accessibilityRole="button"
                accessibilityState={{ selected: now, disabled: !canJump }}
                accessibilityLabel={`Step ${i + 1}: ${STEP_TITLES[s]}${skipped ? ', not needed for this match' : done ? ', done' : ''}`}
                style={styles.segment}>
                <View style={[styles.bar, { backgroundColor: done || now ? t.ink : t.rule }]} />
                <Txt
                  numberOfLines={1}
                  tone={now ? 'ink' : done ? 'soft' : 'faint'}
                  style={{
                    fontFamily: now ? fonts.uiBold : fonts.ui,
                    fontSize: 12,
                    lineHeight: 16,
                    marginTop: 6,
                    textDecorationLine: skipped ? 'line-through' : 'none',
                  }}>
                  {STEP_SHORT[s]}
                </Txt>
              </Pressable>
            );
          })}
        </View>
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
});
