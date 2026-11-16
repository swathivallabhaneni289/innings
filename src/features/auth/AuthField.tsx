import { useState, type Ref } from 'react';
import { Pressable, StyleSheet, TextInput, View, type TextInputProps, type ViewStyle } from 'react-native';

import { CricketBall, Txt } from '@/components';
import { fonts, useTheme } from '@/theme';
import { lengthStatus, PASSWORD_MAX, PASSWORD_MIN } from './validate';

type FieldProps = Omit<TextInputProps, 'style' | 'secureTextEntry'> & {
  label: string;
  /** Red line under the field; the box turns red too. */
  error?: string | null;
  /** Grey line under the field when there's no error. */
  hint?: string | null;
  style?: ViewStyle;
  inputRef?: Ref<TextInput>;
  /** A password, in dots until SHOW is tapped: the one the fan has ("current") or one they're choosing ("new"). */
  secret?: 'current' | 'new';
};

/**
 * Autofill hints for a password field. Android's autofill tells a new password from the current one. iOS gets
 * "password" for both: its "newPassword" type turns on Automatic Strong Passwords, which needs an associated web
 * domain this app doesn't have, and on iOS 26 it left the field with no keyboard at all.
 */
const passwordHints = (kind: 'current' | 'new') =>
  ({ autoComplete: kind === 'new' ? 'new-password' : 'current-password', textContentType: 'password' }) as const;

/**
 * One sign-in field on card stock: the label in ticket fine print above, a box that inks in while you type and
 * turns red with an error, and a line underneath for the hint or the error.
 */
export function AuthField({ label, error, hint, style, inputRef, secret, onFocus, onBlur, ...input }: FieldProps) {
  const t = useTheme();
  const [focused, setFocused] = useState(false);
  const [shown, setShown] = useState(false);
  const border = error ? t.red : focused ? t.ink : t.rule;
  return (
    <View style={[{ marginTop: 18 }, style]}>
      <Txt variant="label" tone="soft">
        {label}
      </Txt>
      <View style={[styles.box, { backgroundColor: t.card, borderColor: border }]}>
        <TextInput
          ref={inputRef}
          {...(secret ? { ...passwordHints(secret), autoCapitalize: 'none', autoCorrect: false, spellCheck: false } : null)}
          {...input}
          secureTextEntry={!!secret && !shown}
          onFocus={(e) => {
            setFocused(true);
            onFocus?.(e);
          }}
          onBlur={(e) => {
            setFocused(false);
            onBlur?.(e);
          }}
          placeholderTextColor={t.inkFaint}
          selectionColor={t.red}
          accessibilityLabel={label}
          style={[styles.input, { color: t.ink }]}
        />
        {secret ? (
          <Pressable
            onPress={() => setShown((s) => !s)}
            hitSlop={10}
            accessibilityRole="button"
            accessibilityLabel={shown ? 'Hide password' : 'Show password'}
            style={({ pressed }) => [styles.toggle, { opacity: pressed ? 0.5 : 1 }]}>
            <Txt variant="label" tone={shown ? 'red' : 'soft'}>
              {shown ? 'Hide' : 'Show'}
            </Txt>
          </Pressable>
        ) : null}
      </View>
      {error || hint ? (
        <Txt variant="bodySmall" tone={error ? 'red' : 'faint'} accessibilityLiveRegion="polite" style={{ marginTop: 6 }}>
          {error || hint}
        </Txt>
      ) : null}
    </View>
  );
}

/**
 * A new password's length as it's typed: eight balls bowling in one per character, turning green at "Good length".
 * The other rules (not your email, not a common password) come back from the server in words.
 */
export function LengthMeter({ password }: { password: string }) {
  const t = useTheme();
  const { ok, toGo, tooLong } = lengthStatus(password);
  const filled = Math.min(PASSWORD_MIN, [...password].length);
  const text = tooLong ? `Keep it under ${PASSWORD_MAX + 1} characters` : ok ? 'Good length' : password ? `${toGo} to go` : `At least ${PASSWORD_MIN} characters`;
  return (
    <View
      style={styles.meter}
      accessible
      accessibilityLabel={ok ? 'Password is long enough' : `Password needs ${toGo} more characters`}
      accessibilityLiveRegion="polite">
      <View style={styles.balls}>
        {Array.from({ length: PASSWORD_MIN }, (_, i) => (
          <CricketBall key={i} size={13} fill={i < filled ? 'full' : 'empty'} color={ok ? t.green : undefined} />
        ))}
      </View>
      <Txt variant="monoSmall" tone={tooLong ? 'red' : ok ? 'green' : 'soft'} style={{ marginLeft: 10 }}>
        {text}
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  box: { flexDirection: 'row', alignItems: 'center', borderWidth: 1.5, marginTop: 6, minHeight: 54 },
  input: { flex: 1, fontFamily: fonts.serif, fontSize: 17, paddingHorizontal: 14, paddingVertical: 14 },
  toggle: { paddingHorizontal: 14, minHeight: 44, justifyContent: 'center' },
  meter: { flexDirection: 'row', alignItems: 'center', marginTop: 10 },
  balls: { flexDirection: 'row', gap: 3 },
});
