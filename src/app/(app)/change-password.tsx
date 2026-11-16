import { router } from 'expo-router';
import { useRef, useState } from 'react';
import { Alert, type TextInput } from 'react-native';

import { Button, Txt } from '@/components';
import { ApiError, changePassword } from '@/data';
import { AuthField, LengthMeter } from '@/features/auth/AuthField';
import { AuthForm } from '@/features/auth/AuthForm';
import { errorText } from '@/features/auth/errors';
import { lengthStatus } from '@/features/auth/validate';

/** Settings → Change password: the current one, then the new one. Other devices are signed out; this one stays in. */
export default function ChangePasswordScreen() {
  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [currentError, setCurrentError] = useState<string | null>(null);
  const [nextError, setNextError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const nextInput = useRef<TextInput>(null);

  const submit = async () => {
    if (busy) return;
    const noCurrent = !current;
    const short = !lengthStatus(next).ok;
    setCurrentError(noCurrent ? 'Enter your current password.' : null);
    setNextError(short ? 'Your new password needs at least 8 characters.' : null);
    if (noCurrent || short) return;
    setBusy(true);
    setError(null);
    try {
      await changePassword(current, next);
      Alert.alert('Password changed', "You're still signed in on this phone. Any other phones were signed out.", [{ text: 'OK', onPress: () => router.back() }]);
    } catch (e) {
      const code = e instanceof ApiError ? e.code : '';
      if (code === 'bad-password') setCurrentError(errorText(e, ''));
      else if (code === 'weak-password') setNextError(errorText(e, ''));
      else setError(errorText(e, "Couldn't change your password just now. Try again in a moment."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthForm title="Change password" lead="Any other phones signed in to your account will be signed out.">
      <AuthField
        label="Current password"
        value={current}
        onChangeText={(v) => {
          setCurrent(v);
          setCurrentError(null);
        }}
        error={currentError}
        secret="current"
        autoFocus
        returnKeyType="next"
        submitBehavior="submit"
        onSubmitEditing={() => nextInput.current?.focus()}
      />
      <AuthField
        label="New password"
        inputRef={nextInput}
        value={next}
        onChangeText={(v) => {
          setNext(v);
          setNextError(null);
        }}
        error={nextError}
        secret="new"
        placeholder="Choose a password"
        returnKeyType="go"
        onSubmitEditing={submit}
      />
      <LengthMeter password={next} />

      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 16 }}>
          {error}
        </Txt>
      ) : null}
      <Button label="Save new password" size="lg" onPress={submit} loading={busy} style={{ marginTop: 26 }} />
    </AuthForm>
  );
}
