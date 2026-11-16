import { Feather } from '@expo/vector-icons';
import { useState } from 'react';
import { Alert, StyleSheet, View } from 'react-native';

import { Button, Panel, Txt } from '@/components';
import { ApiError, deleteAccount } from '@/data';
import { AuthField } from '@/features/auth/AuthField';
import { AuthForm } from '@/features/auth/AuthForm';
import { errorText } from '@/features/auth/errors';
import { useTheme } from '@/theme';

const GOES = ['Every stub, rating and review', 'Your calls, chat messages and Bragging Board points', 'Replies, likes and follows', 'Your fan card and Top 5'];

/**
 * Settings → Delete account: what goes, then the password to prove it's you, then one last "are you sure". Deleting
 * signs every device out and the app goes back to Welcome.
 */
export default function DeleteAccountScreen() {
  const t = useTheme();
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const remove = async () => {
    setBusy(true);
    setError(null);
    try {
      await deleteAccount(password); // ends the session: the app goes back to Welcome
    } catch (e) {
      if (e instanceof ApiError && e.code === 'bad-password') setPasswordError(errorText(e, ''));
      else setError(errorText(e, "Couldn't delete your account. Try again."));
      setBusy(false);
    }
  };

  const confirm = () => {
    if (busy) return;
    if (!password) {
      setPasswordError('Enter your password to delete your account.');
      return;
    }
    Alert.alert('Delete your account for good?', "This can't be undone.", [
      { text: 'Keep my account', style: 'cancel' },
      { text: 'Delete for good', style: 'destructive', onPress: remove },
    ]);
  };

  return (
    <AuthForm title="Delete your account" lead="Everything you made on Innings is deleted for good. It can't be brought back.">
      <Panel style={{ marginTop: 18 }}>
        {GOES.map((line) => (
          <View key={line} style={styles.line}>
            <Feather name="x" size={14} color={t.red} />
            <Txt variant="bodySmall" style={{ marginLeft: 10, flex: 1 }}>
              {line}
            </Txt>
          </View>
        ))}
      </Panel>
      <AuthField
        label="Your password"
        value={password}
        onChangeText={(v) => {
          setPassword(v);
          setPasswordError(null);
        }}
        error={passwordError}
        secret="current"
        returnKeyType="done"
        onSubmitEditing={confirm}
      />
      {error ? (
        <Txt variant="bodySmall" tone="red" accessibilityLiveRegion="polite" style={{ marginTop: 16 }}>
          {error}
        </Txt>
      ) : null}
      <Button
        label={busy ? 'Deleting…' : 'Delete my account'}
        size="lg"
        icon={<Feather name="trash-2" size={16} color={t.onInk} />}
        onPress={confirm}
        loading={busy}
        style={{ marginTop: 26 }}
      />
    </AuthForm>
  );
}

const styles = StyleSheet.create({
  line: { flexDirection: 'row', alignItems: 'center', paddingVertical: 5 },
});
