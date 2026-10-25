import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { useState } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Button, Screen, Txt } from '@/components';
import { startSignIn } from '@/data';
import { errorText } from '@/features/auth/errors';
import { useSignInFlow } from '@/features/auth/flow';
import { callingCode, formatAsYouType, splitInternational, toE164 } from '@/lib/phone';
import { fonts, useTheme } from '@/theme';

const MAX_DIGITS = 15; // E.164's ceiling, calling code included

/** "What's your number?": a country button, the number formatted as it's typed, and one button to text a code. */
export default function PhoneScreen() {
  const t = useTheme();
  const country = useSignInFlow((s) => s.country);
  const setCountry = useSignInFlow((s) => s.setCountry);
  const sent = useSignInFlow((s) => s.sent);
  const setSent = useSignInFlow((s) => s.setSent);
  const [digits, setDigits] = useState(''); // national digits only; the field shows them formatted
  // "+44 7…" typed a key at a time: until the calling code is complete enough to name a country, the field keeps
  // exactly what was typed. Once it parses, the country switches and the rest becomes national digits.
  const [intlDraft, setIntlDraft] = useState<string | null>(null);
  const [focused, setFocused] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const shown = intlDraft ?? formatAsYouType(digits, country);
  const e164 = intlDraft ? null : toE164(digits, country);

  const onChange = (text: string) => {
    setError(null);
    const intl = splitInternational(text, country);
    if (intl) {
      setIntlDraft(null);
      setCountry(intl.country);
      setDigits(intl.national.slice(0, MAX_DIGITS));
      return;
    }
    if (text.trim().startsWith('+')) {
      const typed = text.replace(/\D/g, '');
      setIntlDraft(typed ? `+${typed.slice(0, MAX_DIGITS)}` : '+');
      setDigits('');
      return;
    }
    setIntlDraft(null);
    let next = text.replace(/\D/g, '');
    // Backspacing over a space or bracket leaves the digits as they were: take the last digit off instead.
    if (next === digits && text.length < shown.length) next = next.slice(0, -1);
    setDigits(next.slice(0, MAX_DIGITS));
  };

  const send = async () => {
    if (!e164 || busy) return;
    // Back from the code screen with the same number and a code still live: no need to text another.
    if (sent && sent.phone === e164 && Date.now() < sent.sentAt + sent.expiresInSeconds * 1000) {
      router.push('/code');
      return;
    }
    setBusy(true);
    setError(null);
    try {
      setSent(await startSignIn(e164));
      router.push('/code');
    } catch (e) {
      setError(errorText(e, "Couldn't send a code just now. Try again in a moment."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Screen back>
      <Txt variant="h1" accessibilityRole="header" style={{ marginTop: 4 }}>
        What&apos;s your number?
      </Txt>
      <Txt variant="body" tone="soft" style={{ marginTop: 6 }}>
        New here or back for another season, it&apos;s the same way in: we text you a code.
      </Txt>

      <View style={[styles.field, { backgroundColor: t.card, borderColor: focused ? t.ink : t.rule }]}>
        <Pressable
          onPress={() => router.push('/country')}
          accessibilityRole="button"
          accessibilityLabel={`Country: ${country}, ${callingCode(country)}. Change`}
          style={({ pressed }) => [styles.country, { opacity: pressed ? 0.6 : 1 }]}>
          <Txt variant="score" style={styles.dial}>
            {callingCode(country)}
          </Txt>
          <Txt variant="label" tone="soft" style={{ marginLeft: 6 }}>
            {country}
          </Txt>
          <Feather name="chevron-down" size={14} color={t.inkSoft} style={{ marginLeft: 2 }} />
        </Pressable>
        <View style={[styles.divider, { backgroundColor: t.rule }]} />
        <TextInput
          value={shown}
          onChangeText={onChange}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          onSubmitEditing={send}
          autoFocus
          keyboardType="phone-pad"
          autoComplete="tel"
          textContentType="telephoneNumber"
          returnKeyType="done"
          placeholder="Mobile number"
          placeholderTextColor={t.inkFaint}
          selectionColor={t.red}
          accessibilityLabel="Mobile number"
          style={[styles.input, { color: t.ink }]}
        />
      </View>

      <Txt variant="monoSmall" tone="soft" style={{ marginTop: 10 }}>
        We&apos;ll text you a 6-digit code. Standard message rates may apply.
      </Txt>

      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 14 }}>
          {error}
        </Txt>
      ) : null}

      <Button label="Text me a code" size="lg" onPress={send} disabled={!e164} loading={busy} style={{ marginTop: 24 }} />
    </Screen>
  );
}

const styles = StyleSheet.create({
  field: { flexDirection: 'row', alignItems: 'center', borderWidth: 1.5, marginTop: 26, minHeight: 60 },
  country: { flexDirection: 'row', alignItems: 'baseline', paddingHorizontal: 14, paddingVertical: 14, minHeight: 56 },
  dial: { fontSize: 24, lineHeight: 28 },
  divider: { width: StyleSheet.hairlineWidth, alignSelf: 'stretch', marginVertical: 10 },
  input: { flex: 1, fontFamily: fonts.condensed, fontSize: 26, letterSpacing: 0.6, paddingHorizontal: 14, paddingVertical: 12 },
});
