import { router } from 'expo-router';
import { useRef, useState } from 'react';
import { Pressable, type TextInput } from 'react-native';

import { Button, Txt } from '@/components';
import { ApiError, resendConfirmation, signInWithPassword } from '@/data';
import { AuthField } from '@/features/auth/AuthField';
import { AuthForm, SwitchLine } from '@/features/auth/AuthForm';
import { errorText } from '@/features/auth/errors';
import { liveCode, useAuthFlow } from '@/features/auth/flow';
import { looksLikeEmail, normEmail } from '@/features/auth/validate';
import { fonts } from '@/theme';

/**
 * "Welcome back": email and password, with "Forgot password?" under the password. An account that never entered
 * its sign-up code is sent a new one and taken to enter it.
 */
export default function SignInScreen() {
  const savedEmail = useAuthFlow((s) => s.email);
  const setFlowEmail = useAuthFlow((s) => s.setEmail);
  const setSent = useAuthFlow((s) => s.setSent);
  const [email, setEmail] = useState(savedEmail);
  const [password, setPassword] = useState('');
  const [emailError, setEmailError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const passwordInput = useRef<TextInput>(null);

  const carry = (to: '/sign-up' | '/forgot') => {
    setFlowEmail(normEmail(email));
    if (to === '/sign-up') router.replace(to);
    else router.push(to);
  };

  /** Right password, unconfirmed address: a new code, then the code screen. Inside the resend gap the last one works. */
  const confirmFirst = async (address: string) => {
    try {
      setSent(await resendConfirmation(address), 'confirm');
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'too-soon' && liveCode(useAuthFlow.getState().sent, address, 'confirm'))) throw e;
    }
    router.push('/confirm');
  };

  const submit = async () => {
    if (busy) return;
    if (!looksLikeEmail(email)) {
      setEmailError('Enter your full email address, like name@example.com.');
      return;
    }
    if (!password) {
      setError('Enter your password.');
      return;
    }
    const address = normEmail(email);
    setBusy(true);
    setError(null);
    try {
      const account = await signInWithPassword(address, password);
      // The root guard has opened the right group by now; an onboarded fan is sent Home explicitly as well.
      if (account.onboarded) router.replace('/');
    } catch (e) {
      if (e instanceof ApiError && e.code === 'email-unconfirmed') {
        try {
          await confirmFirst(address);
        } catch (e2) {
          setError(errorText(e2, "Couldn't send a new code just now. Try again in a moment."));
        }
      } else if (e instanceof ApiError && e.code === 'bad-email') setEmailError(errorText(e, ''));
      else setError(errorText(e, "Couldn't sign you in just now. Try again in a moment."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthForm title="Welcome back" lead="Sign in with the email and password you signed up with.">
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
        autoFocus={!savedEmail}
        autoCapitalize="none"
        autoCorrect={false}
        spellCheck={false}
        keyboardType="email-address"
        autoComplete="email"
        textContentType="username"
        returnKeyType="next"
        submitBehavior="submit"
        onSubmitEditing={() => passwordInput.current?.focus()}
      />
      <AuthField
        label="Password"
        inputRef={passwordInput}
        value={password}
        onChangeText={(v) => {
          setPassword(v);
          setError(null);
        }}
        secret="current"
        autoFocus={!!savedEmail}
        placeholder="Your password"
        returnKeyType="go"
        onSubmitEditing={submit}
      />
      <Pressable
        onPress={() => carry('/forgot')}
        hitSlop={8}
        accessibilityRole="button"
        style={({ pressed }) => ({ alignSelf: 'flex-end', minHeight: 44, justifyContent: 'center', opacity: pressed ? 0.6 : 1 })}>
        <Txt variant="bodySmall" tone="red" style={{ fontFamily: fonts.uiBold }}>
          Forgot password?
        </Txt>
      </Pressable>

      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 6 }}>
          {error}
        </Txt>
      ) : null}
      <Button label="Sign in" size="lg" onPress={submit} loading={busy} style={{ marginTop: 20 }} />
      <SwitchLine text="New here?" action="Make an account" onPress={() => carry('/sign-up')} />
    </AuthForm>
  );
}
