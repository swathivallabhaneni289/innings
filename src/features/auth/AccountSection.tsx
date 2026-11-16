import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { Alert, Pressable, StyleSheet, View } from 'react-native';

import { Panel, Rule, SectionHeader, Txt } from '@/components';
import { signOut, useAccount } from '@/data';
import { fonts, useTheme } from '@/theme';

/** Settings → Account: the email this account signs in with, Change password, Sign out, and Delete account. */
export function AccountSection() {
  const account = useAccount();

  const confirmSignOut = () =>
    Alert.alert('Sign out?', 'Your stubs, calls and follows stay on your account. Sign in with your email and password to pick up where you left off.', [
      { text: 'Stay', style: 'cancel' },
      { text: 'Sign out', style: 'destructive', onPress: signOut },
    ]);

  return (
    <View>
      <SectionHeader title="Account" />
      <Panel style={{ paddingVertical: 4 }}>
        <View style={styles.row}>
          <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label, marginRight: 12 }}>
            Email
          </Txt>
          <Txt variant="mono" numberOfLines={1} ellipsizeMode="middle" selectable style={{ flex: 1, textAlign: 'right' }}>
            {account.data?.email || '–'}
          </Txt>
        </View>
        <Txt variant="monoSmall" tone="faint" style={{ marginTop: -4, marginBottom: 12 }}>
          Only you see this. Sign in with it on any phone.
        </Txt>
        <Rule />
        <Row icon="lock" label="Change password" onPress={() => router.push('/change-password')} chevron />
        <Rule />
        <Row icon="log-out" label="Sign out" onPress={confirmSignOut} />
        <Rule />
        <Row icon="trash-2" label="Delete account" tone="red" onPress={() => router.push('/delete-account')} chevron />
      </Panel>
    </View>
  );
}

/** One line of the account panel: an icon, a label, and a chevron when it opens another screen. */
function Row({ icon, label, onPress, tone, chevron }: { icon: 'lock' | 'log-out' | 'trash-2'; label: string; onPress: () => void; tone?: 'red'; chevron?: boolean }) {
  const t = useTheme();
  return (
    <Pressable onPress={onPress} accessibilityRole="button" style={({ pressed }) => [styles.row, { opacity: pressed ? 0.6 : 1 }]}>
      <Feather name={icon} size={17} color={tone === 'red' ? t.red : t.ink} />
      <Txt variant="body" tone={tone} style={styles.label}>
        {label}
      </Txt>
      {chevron ? <Feather name="chevron-right" size={17} color={t.inkFaint} /> : null}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', minHeight: 50 },
  label: { marginLeft: 12, fontFamily: fonts.label, flex: 1 },
});
