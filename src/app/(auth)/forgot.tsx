import { router } from 'expo-router';
import { useState } from 'react';

import { Button, Txt } from '@/components';
import { ApiError, forgotPassword } from '@/data';
import { AuthField } from '@/features/auth/AuthField';
import { AuthForm } from '@/features/auth/AuthForm';
import { errorText } from '@/features/auth/errors';
import { useAuthFlow } from '@/features/auth/flow';
import { looksLikeEmail, normEmail } from '@/features/auth/validate';

/** "Forgot your password?": the email to send a reset code to. */
export default function ForgotScreen() {
  const savedEmail = useAuthFlow((s) => s.email);
  const setSent = useAuthFlow((s) => s.setSent);
  const [email, setEmail] = useState(savedEmail);
  const [emailError, setEmailError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    if (busy) return;
    if (!looksLikeEmail(email)) {
      setEmailError('Enter your full email address, like name@example.com.');
      return;
    }
    setBusy(true);
    setError(null);
    try {
      setSent(await forgotPassword(normEmail(email)), 'reset');
      router.push('/reset-code');
    } catch (e) {
      if (e instanceof ApiError && e.code === 'bad-email') setEmailError(errorText(e, ''));
      else setError(errorText(e, "Couldn't send a code just now. Try again in a moment."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthForm title="Forgot your password?" lead="Enter the email you signed up with, and we'll send you a code to choose a new password.">
      <AuthField
        label="Email"
        value={email}
        onChangeText={(v) => {
          setEmail(v);
          setEmailError(null);
          setError(null);
        }}
        error={emailError}
        placeholder="name@example.com"
        autoFocus
        autoCapitalize="none"
        autoCorrect={false}
        spellCheck={false}
        keyboardType="email-address"
        autoComplete="email"
        textContentType="username"
        returnKeyType="send"
        onSubmitEditing={submit}
      />
      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 16 }}>
          {error}
        </Txt>
      ) : null}
      <Button label="Send me a code" size="lg" onPress={submit} loading={busy} style={{ marginTop: 26 }} />
      <Txt variant="monoSmall" tone="faint" align="center" style={{ marginTop: 10 }}>
        Your password stays the same until you choose a new one.
      </Txt>
    </AuthForm>
  );
}
