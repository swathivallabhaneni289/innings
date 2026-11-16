import { Redirect, router } from 'expo-router';
import { useState } from 'react';

import { Button, Txt } from '@/components';
import { ApiError, resetPassword } from '@/data';
import { AuthField, LengthMeter } from '@/features/auth/AuthField';
import { AuthForm } from '@/features/auth/AuthForm';
import { errorText } from '@/features/auth/errors';
import { useAuthFlow } from '@/features/auth/flow';
import { lengthStatus } from '@/features/auth/validate';

/**
 * "Choose a new password", after the right reset code. Saving signs this phone in and every other device out; the
 * root guard then opens the app (or the fan card, for an account that never finished signing up).
 */
export default function NewPasswordScreen() {
  const reset = useAuthFlow((s) => s.reset);
  const email = useAuthFlow((s) => s.email);
  const clear = useAuthFlow((s) => s.clear);
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [expired, setExpired] = useState(false);
  const [busy, setBusy] = useState(false);
  if (!reset) return <Redirect href="/forgot" />;

  const startAgain = () => router.replace('/forgot');

  const submit = async () => {
    if (busy) return;
    if (!lengthStatus(password).ok) {
      setPasswordError('Your password needs at least 8 characters.');
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const account = await resetPassword(reset.token, password);
      clear();
      if (account.onboarded) router.replace('/');
    } catch (e) {
      const code = e instanceof ApiError ? e.code : '';
      if (code === 'weak-password') setPasswordError(errorText(e, ''));
      else if (code === 'reset-expired') {
        setExpired(true);
        setError(errorText(e, 'That reset has run out. Ask for a new code.'));
      } else setError(errorText(e, "Couldn't save your new password just now. Try again in a moment."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthForm title="Choose a new password" lead={`For ${email}. You'll stay signed in on this phone and be signed out everywhere else.`}>
      <AuthField
        label="New password"
        value={password}
        onChangeText={(v) => {
          setPassword(v);
          setPasswordError(null);
        }}
        error={passwordError}
        secret="new"
        autoFocus
        placeholder="Choose a password"
        returnKeyType="go"
        onSubmitEditing={submit}
      />
      <LengthMeter password={password} />

      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 16 }}>
          {error}
        </Txt>
      ) : null}
      {expired ? (
        <Button label="Send a new code" kind="outline" size="lg" onPress={startAgain} style={{ marginTop: 22 }} />
      ) : (
        <Button label="Save and sign in" size="lg" onPress={submit} loading={busy} style={{ marginTop: 26 }} />
      )}
    </AuthForm>
  );
}
