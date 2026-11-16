import type { ReactNode } from 'react';
import { Pressable, ScrollView } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { KeyboardAware, Screen, Txt } from '@/components';
import { fonts } from '@/theme';

/**
 * The frame for a sign-in form: a back arrow, a heading, a line under it, then the fields. It scrolls the focused
 * field above the keyboard on both platforms (see KeyboardAware), and taps on buttons land while the keyboard is up.
 */
export function AuthForm({ title, lead, children }: { title: string; lead?: ReactNode; children: ReactNode }) {
  const insets = useSafeAreaInsets();
  return (
    <Screen back scroll={false}>
      <KeyboardAware scrollInsets>
        <ScrollView
          automaticallyAdjustKeyboardInsets
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
          contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 4, paddingBottom: insets.bottom + 32 }}>
          <Txt variant="h1" accessibilityRole="header" style={{ marginTop: 4 }}>
            {title}
          </Txt>
          {lead ? (
            <Txt variant="body" tone="soft" style={{ marginTop: 6 }}>
              {lead}
            </Txt>
          ) : null}
          {children}
        </ScrollView>
      </KeyboardAware>
    </Screen>
  );
}

/** "Already have an account? Sign in": a quiet line with one red action under a form's main button; all of it taps. */
export function SwitchLine({ text, action, onPress }: { text: string; action: string; onPress: () => void }) {
  return (
    <Pressable
      onPress={onPress}
      accessibilityRole="button"
      accessibilityLabel={`${text} ${action}`}
      style={({ pressed }) => [{ minHeight: 44, justifyContent: 'center', marginTop: 10, opacity: pressed ? 0.6 : 1 }]}>
      <Txt variant="bodySmall" tone="soft" align="center">
        {text}{' '}
        <Txt variant="bodySmall" tone="red" style={{ fontFamily: fonts.uiBold }}>
          {action}
        </Txt>
      </Txt>
    </Pressable>
  );
}
