import { useState, type RefObject } from 'react';
import { StyleSheet, TextInput, View } from 'react-native';

import { Txt } from '@/components';
import { useTheme } from '@/theme';

export const CODE_LENGTH = 6;

/**
 * Six cells printed like a ticket's red serial digits, over one real TextInput. The input covers the cells
 * with its text drawn transparent (not hidden), so taps focus it and iOS and Android both offer the texted
 * code above the keyboard.
 */
export function CodeCells({
  value,
  onChangeText,
  inputRef,
  error = false,
}: {
  value: string;
  onChangeText: (digits: string) => void;
  inputRef: RefObject<TextInput | null>;
  error?: boolean;
}) {
  const t = useTheme();
  const [focused, setFocused] = useState(false);
  const active = Math.min(value.length, CODE_LENGTH - 1);
  return (
    <View style={styles.wrap}>
      <View style={styles.row} pointerEvents="none" importantForAccessibility="no-hide-descendants" accessibilityElementsHidden>
        {Array.from({ length: CODE_LENGTH }, (_, i) => {
          const digit = value[i] ?? '';
          const current = focused && i === active && value.length < CODE_LENGTH;
          return (
            <View
              key={i}
              style={[
                styles.cell,
                { backgroundColor: t.card, borderBottomColor: error ? t.red : current ? t.ink : t.rule },
                i === 2 ? { marginRight: 14 } : null, // read as two triplets, like a serial
              ]}>
              <Txt variant="display" tone="red" style={styles.digit}>
                {digit}
              </Txt>
            </View>
          );
        })}
      </View>
      <TextInput
        ref={inputRef}
        value={value}
        onChangeText={(s) => onChangeText(s.replace(/\D/g, '').slice(0, CODE_LENGTH))}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        autoFocus
        autoComplete="one-time-code"
        keyboardType="number-pad"
        maxLength={CODE_LENGTH}
        caretHidden
        contextMenuHidden
        selectionColor="transparent"
        accessibilityLabel={`${CODE_LENGTH}-digit code`}
        style={[StyleSheet.absoluteFill, styles.input]}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { marginTop: 28, alignSelf: 'flex-start' },
  row: { flexDirection: 'row' },
  cell: { width: 46, height: 62, marginRight: 6, alignItems: 'center', justifyContent: 'center', borderBottomWidth: 3 },
  digit: { fontSize: 40, lineHeight: 44 },
  // Present and focusable, drawn invisibly: the cells above are what the fan sees.
  input: { color: 'transparent', backgroundColor: 'transparent', fontSize: 1, opacity: 0.02 },
});
