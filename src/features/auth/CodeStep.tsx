import { useRef, useState } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Screen, Txt } from '@/components';
import { ApiError, useNow, type AuthCodeResponse } from '@/data';
import { countdown } from '@/lib/format';
import { fonts, useTheme } from '@/theme';
import { CODE_LENGTH, CodeCells } from './CodeCells';
import { errorText } from './errors';
import type { SentCode } from './flow';

/**
 * "Enter the code": six serial cells that check themselves at the sixth digit, a resend countdown, a way back to
 * fix the address, and (dev builds) the code itself to tap in. Used to confirm a new account and to reset a
 * password; `onCheck` does the rest and `onResend` asks for another code.
 */
export function CodeStep({
  title,
  lead,
  sent,
  onCheck,
  onResend,
  onChangeEmail,
}: {
  title: string;
  lead: string; // "We sent a 6-digit code to"
  sent: SentCode;
  onCheck: (code: string) => Promise<void>;
  onResend: () => Promise<AuthCodeResponse>;
  onChangeEmail: () => void;
}) {
  const t = useTheme();
  const now = useNow(1_000);
  const input = useRef<TextInput>(null);
  // One check at a time: autofill can deliver the six digits twice. A ref, not state, so a second onChangeText in
  // the same tick sees it; the input stays editable so a retry keeps the keyboard up.
  const busy = useRef(false);
  const [code, setCode] = useState('');
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [expired, setExpired] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [resending, setResending] = useState(false);

  const check = async (value: string) => {
    if (value.length !== CODE_LENGTH || busy.current) return;
    busy.current = true;
    setChecking(true);
    setError(null);
    setNotice(null);
    try {
      await onCheck(value);
      // Success moves on (to the next screen, or the root guard opens the app); keep the cells as they are.
    } catch (e) {
      busy.current = false;
      setChecking(false);
      setCode('');
      if (e instanceof ApiError && e.code === 'code-expired') {
        setExpired(true);
        setError(e.message || 'That code has run out. Ask for a new one below.');
      } else {
        setError(errorText(e, "That code didn't work. Check the email and try again."));
      }
      input.current?.focus();
    }
  };

  const onChange = (digits: string) => {
    if (busy.current) return; // the code being checked stays put until the answer comes back
    setCode(digits);
    if (digits.length > 0) setError(null);
    if (digits.length === CODE_LENGTH) check(digits);
  };

  const resendAt = sent.sentAt + sent.resendInSeconds * 1000;
  const wait = Math.max(0, Math.ceil((resendAt - now) / 1000));

  const resend = async () => {
    if (resending) return;
    setResending(true);
    setError(null);
    setNotice(null);
    try {
      const res = await onResend();
      setExpired(false);
      setCode('');
      setNotice(res.devCode || res.resendInSeconds >= 30 ? 'New code on its way.' : 'A code is already on its way.');
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
    check(sent.devCode);
  };

  return (
    <Screen back>
      <Txt variant="h1" accessibilityRole="header" style={{ marginTop: 4 }}>
        {title}
      </Txt>
      <Txt variant="body" tone="soft" style={{ marginTop: 6 }}>
        {lead}{' '}
        <Txt variant="body" style={{ fontFamily: fonts.label }}>
          {sent.email}
        </Txt>
      </Txt>
      <Pressable onPress={onChangeEmail} hitSlop={12} accessibilityRole="button" style={styles.change}>
        <Txt variant="bodySmall" tone="red" style={{ fontFamily: fonts.uiBold }}>
          Wrong email? Change it
        </Txt>
      </Pressable>

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
        ) : (
          <Txt variant="bodySmall" tone="faint">
            It can take a minute to arrive. Check spam if it doesn&apos;t.
          </Txt>
        )}
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
            Nothing is emailed until the server has a mail account. Tap to fill it in.
          </Txt>
        </Pressable>
      ) : null}
    </Screen>
  );
}

const styles = StyleSheet.create({
  change: { alignSelf: 'flex-start', minHeight: 36, justifyContent: 'center', marginTop: 2 },
  status: { minHeight: 40, justifyContent: 'center', marginTop: 10 },
  resendRow: { flexDirection: 'row', alignItems: 'baseline', alignSelf: 'flex-start', minHeight: 44, paddingVertical: 10 },
  dev: { marginTop: 28, borderWidth: 1.5, borderStyle: 'dashed', paddingHorizontal: 14, paddingVertical: 12 },
});
