import { Feather } from '@expo/vector-icons';
import { useState } from 'react';
import { ActivityIndicator, Alert, Pressable, StyleSheet, View } from 'react-native';

import { Panel, Rule, SectionHeader, Txt } from '@/components';
import { deleteAccount, signOut, useAccount } from '@/data';
import { maskPhone } from '@/lib/phone';
import { fonts, useTheme } from '@/theme';
import { errorText } from './errors';

/** Settings → Account: the number this account signs in with (masked), Sign out, and Delete account. */
export function AccountSection() {
  const t = useTheme();
  const account = useAccount();
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const confirmSignOut = () =>
    Alert.alert('Sign out?', 'Your stubs, calls and follows stay on your account. Sign in with the same number to pick up where you left off.', [
      { text: 'Stay', style: 'cancel' },
      { text: 'Sign out', style: 'destructive', onPress: signOut },
    ]);

  const remove = async () => {
    setDeleting(true);
    setError(null);
    try {
      await deleteAccount(); // ends the session: the app goes back to Welcome
    } catch (e) {
      setError(errorText(e, "Couldn't delete your account. Try again."));
      setDeleting(false);
    }
  };

  const confirmDelete = () =>
    Alert.alert(
      'Delete your account?',
      "This deletes your stubs, calls and follows for good. Your Top 5 and fan card go with them. It can't be undone.",
      [
        { text: 'Keep my account', style: 'cancel' },
        { text: 'Delete for good', style: 'destructive', onPress: remove },
      ],
    );

  return (
    <View>
      <SectionHeader title="Account" />
      <Panel style={{ paddingVertical: 4 }}>
        <View style={styles.row}>
          <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label, flex: 1 }}>
            Phone
          </Txt>
          <Txt variant="mono" accessibilityLabel={account.data ? `Phone ending ${account.data.phone.slice(-4)}` : undefined}>
            {account.data ? maskPhone(account.data.phone) : '–'}
          </Txt>
        </View>
        <Txt variant="monoSmall" tone="faint" style={{ marginTop: -4, marginBottom: 12 }}>
          Only you see this. Sign in with it on any phone.
        </Txt>
        <Rule />
        <Pressable onPress={confirmSignOut} accessibilityRole="button" style={({ pressed }) => [styles.row, { opacity: pressed ? 0.6 : 1 }]}>
          <Feather name="log-out" size={17} color={t.ink} />
          <Txt variant="body" style={styles.label}>
            Sign out
          </Txt>
        </Pressable>
        <Rule />
        <Pressable
          onPress={confirmDelete}
          disabled={deleting}
          accessibilityRole="button"
          style={({ pressed }) => [styles.row, { opacity: pressed ? 0.6 : 1 }]}>
          {deleting ? <ActivityIndicator size="small" color={t.red} /> : <Feather name="trash-2" size={17} color={t.red} />}
          <Txt variant="body" tone="red" style={styles.label}>
            {deleting ? 'Deleting…' : 'Delete account'}
          </Txt>
        </Pressable>
      </Panel>
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
          {error}
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', minHeight: 50 },
  label: { marginLeft: 12, fontFamily: fonts.label },
});
