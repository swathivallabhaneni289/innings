import { router } from 'expo-router';
import { useRef, useState } from 'react';
import { Pressable, type TextInput } from 'react-native';

import { Button, Txt } from '@/components';
import { ApiError, signUp } from '@/data';
import { AuthField, LengthMeter } from '@/features/auth/AuthField';
import { AuthForm, SwitchLine } from '@/features/auth/AuthForm';
import { errorText } from '@/features/auth/errors';
import { useAuthFlow } from '@/features/auth/flow';
import { lengthStatus, looksLikeEmail, normEmail } from '@/features/auth/validate';
import { fonts } from '@/theme';

/** "Make your account": an email and a password. A 6-digit code then goes to the email to check it's really yours. */
export default function SignUpScreen() {
  const savedEmail = useAuthFlow((s) => s.email);
  const setFlowEmail = useAuthFlow((s) => s.setEmail);
  const setSent = useAuthFlow((s) => s.setSent);
  const [email, setEmail] = useState(savedEmail);
  const [password, setPassword] = useState('');
  const [emailError, setEmailError] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [taken, setTaken] = useState(false);
  const [busy, setBusy] = useState(false);
  const passwordInput = useRef<TextInput>(null);

  const toSignIn = () => {
    setFlowEmail(normEmail(email));
    router.replace('/sign-in');
  };

  const submit = async () => {
    if (busy) return;
    const badEmail = !looksLikeEmail(email);
    const badPassword = !lengthStatus(password).ok;
    setEmailError(badEmail ? 'Enter your full email address, like name@example.com.' : null);
    setPasswordError(badPassword ? 'Your password needs at least 8 characters.' : null);
    if (badEmail || badPassword) return;
    setBusy(true);
    setError(null);
    setTaken(false);
    try {
      setSent(await signUp(normEmail(email), password), 'confirm');
      router.push('/confirm');
    } catch (e) {
      const code = e instanceof ApiError ? e.code : '';
      if (code === 'email-taken') {
        setTaken(true);
        setEmailError(errorText(e, 'There is already an account with this email.'));
      } else if (code === 'bad-email') setEmailError(errorText(e, ''));
      else if (code === 'weak-password') setPasswordError(errorText(e, ''));
      else setError(errorText(e, "Couldn't make your account just now. Try again in a moment."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthForm title="Make your account" lead="Your email and a password. We'll send a code to the email to check it's yours.">
      <AuthField
        label="Email"
        value={email}
        onChangeText={(v) => {
          setEmail(v);
          setEmailError(null);
          setTaken(false);
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
      {taken ? (
        <Pressable onPress={toSignIn} hitSlop={8} accessibilityRole="button" style={{ alignSelf: 'flex-start', minHeight: 40, justifyContent: 'center' }}>
          <Txt variant="bodySmall" tone="red" style={{ fontFamily: fonts.uiBold }}>
            Sign in with this email instead
          </Txt>
        </Pressable>
      ) : null}
      <AuthField
        label="Password"
        inputRef={passwordInput}
        value={password}
        onChangeText={(v) => {
          setPassword(v);
          setPasswordError(null);
        }}
        error={passwordError}
        secret="new"
        autoFocus={!!savedEmail}
        placeholder="Choose a password"
        returnKeyType="go"
        onSubmitEditing={submit}
      />
      <LengthMeter password={password} />
      <Txt variant="monoSmall" tone="faint" style={{ marginTop: 8 }}>
        Spaces are fine, so a short phrase works well.
      </Txt>

      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 16 }}>
          {error}
        </Txt>
      ) : null}
      <Button label="Make my account" size="lg" onPress={submit} loading={busy} style={{ marginTop: 26 }} />
      <Txt variant="monoSmall" tone="faint" align="center" style={{ marginTop: 10 }}>
        Your email is only for signing in. Other fans never see it.
      </Txt>
      <SwitchLine text="Already have an account?" action="Sign in" onPress={toSignIn} />
    </AuthForm>
  );
}
