import { Redirect, router } from 'expo-router';
import { useRef, useState } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Screen, Txt } from '@/components';
import { ApiError, signIn, startSignIn, useNow } from '@/data';
import { CODE_LENGTH, CodeCells } from '@/features/auth/CodeCells';
import { errorText } from '@/features/auth/errors';
import { useSignInFlow } from '@/features/auth/flow';
import { countdown } from '@/lib/format';
import { formatPhone } from '@/lib/phone';
import { fonts, useTheme } from '@/theme';

/**
 * "Enter the code": six serial cells that verify themselves at the sixth digit. A new account then lands on
 * onboarding (the only group open to it); a returning, onboarded one goes straight to Home.
 */
export default function CodeScreen() {
  const t = useTheme();
  const sent = useSignInFlow((s) => s.sent);
  const setSent = useSignInFlow((s) => s.setSent);
  const now = useNow(1_000);
  const input = useRef<TextInput>(null);
  // One verify at a time: autofill can deliver the six digits twice. A ref, not state, so a second
  // onChangeText in the same tick sees it; the input stays editable so a retry keeps the keyboard up.
  const busy = useRef(false);
  const [code, setCode] = useState('');
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [expired, setExpired] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [resending, setResending] = useState(false);

  if (!sent) return <Redirect href="/phone" />;

  const verify = async (value: string) => {
    if (value.length !== CODE_LENGTH || busy.current) return;
    busy.current = true;
    setChecking(true);
    setError(null);
    setNotice(null);
    try {
      const account = await signIn(sent.phone, value);
      // The root guard has opened the right group by now; an onboarded fan is sent Home explicitly as well.
      if (account.onboarded) router.replace('/');
    } catch (e) {
      busy.current = false;
      setChecking(false);
      setCode('');
      if (e instanceof ApiError && e.code === 'code-expired') {
        setExpired(true);
        setError(e.message || 'That code has run out. Ask for a new one below.');
      } else {
        setError(errorText(e, "That code didn't work. Check the text and try again."));
      }
      input.current?.focus();
    }
  };

  const onChange = (digits: string) => {
    if (busy.current) return; // the code being checked stays put until the answer comes back
    setCode(digits);
    if (digits.length > 0) setError(null);
    if (digits.length === CODE_LENGTH) verify(digits);
  };

  const resendAt = sent.sentAt + sent.resendInSeconds * 1000;
  const wait = Math.max(0, Math.ceil((resendAt - now) / 1000));

  const resend = async () => {
    if (resending) return;
    setResending(true);
    setError(null);
    setNotice(null);
    try {
      setSent(await startSignIn(sent.phone));
      setExpired(false);
      setCode('');
      setNotice('New code on its way.');
      input.current?.focus();
    } catch (e) {
      setError(errorText(e, "Couldn't send a new code. Try again in a moment."));
    } finally {
      setResending(false);
    }
  };

  const fillDev = () => {
    if (!sent.devCode || busy.current) return;
    setCode(sent.devCode);
    verify(sent.devCode);
  };

  return (
    <Screen back>
      <Txt variant="h1" accessibilityRole="header" style={{ marginTop: 4 }}>
        Enter the code
      </Txt>
      <View style={styles.sentRow}>
        <Txt variant="body" tone="soft">
          Sent to{' '}
          <Txt variant="body" style={{ fontFamily: fonts.label }}>
            {formatPhone(sent.phone)}
          </Txt>
          {'  ·  '}
        </Txt>
        <Pressable onPress={() => router.back()} hitSlop={12} accessibilityRole="button" accessibilityLabel="Change number">
          <Txt variant="body" tone="red" style={{ fontFamily: fonts.uiBold }}>
            Change
          </Txt>
        </Pressable>
      </View>

      <CodeCells value={code} onChangeText={onChange} inputRef={input} error={!!error} />

      <View style={styles.status} accessibilityLiveRegion="polite">
        {checking ? (
          <Txt variant="bodySmall" tone="soft">
            Checking with the third umpire…
          </Txt>
        ) : error ? (
          <Txt variant="bodySmall" tone="red">
            {error}
          </Txt>
        ) : notice ? (
          <Txt variant="bodySmall" tone="soft">
            {notice}
          </Txt>
        ) : null}
      </View>

      {wait > 0 && !expired ? (
        <View style={styles.resendRow}>
          <Txt variant="bodySmall" tone="faint">
            Resend code in
          </Txt>
          <Txt variant="mono" tone="faint" style={{ marginLeft: 6 }}>
            {countdown(new Date(resendAt).toISOString(), now)}
          </Txt>
        </View>
      ) : (
        <Pressable onPress={resend} disabled={resending || checking} hitSlop={12} accessibilityRole="button" style={styles.resendRow}>
          <Txt variant="bodySmall" tone={resending ? 'faint' : 'red'} style={{ fontFamily: fonts.uiBold }}>
            {resending ? 'Sending…' : 'Resend code'}
          </Txt>
        </Pressable>
      )}

      {sent.devCode ? (
        <Pressable
          onPress={fillDev}
          accessibilityRole="button"
          accessibilityLabel={`Dev mode: your code is ${sent.devCode}. Tap to fill it in.`}
          style={({ pressed }) => [styles.dev, { borderColor: t.cardEdge, opacity: pressed ? 0.6 : 1 }]}>
          <Txt variant="mono">Dev mode: your code is {sent.devCode}</Txt>
          <Txt variant="monoSmall" tone="soft" style={{ marginTop: 2 }}>
            No text in dev builds. Tap to fill it in.
          </Txt>
        </Pressable>
      ) : null}
    </Screen>
  );
}

const styles = StyleSheet.create({
  sentRow: { flexDirection: 'row', alignItems: 'baseline', flexWrap: 'wrap', marginTop: 6 },
  status: { minHeight: 40, justifyContent: 'center', marginTop: 10 },
  resendRow: { flexDirection: 'row', alignItems: 'baseline', alignSelf: 'flex-start', minHeight: 44, paddingVertical: 10 },
  dev: { marginTop: 28, borderWidth: 1.5, borderStyle: 'dashed', paddingHorizontal: 14, paddingVertical: 12 },
});
