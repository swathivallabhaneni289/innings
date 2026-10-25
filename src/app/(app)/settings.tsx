import Constants from 'expo-constants';
import { ScrollView } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Panel, QueryState, Screen, SectionHeader, Txt } from '@/components';
import { useMe } from '@/data';
import { API_URL } from '@/data/client';
import { AccountSection } from '@/features/auth/AccountSection';
import { AppearancePicker } from '@/features/profile/AppearancePicker';
import { SettingsForm } from '@/features/profile/SettingsForm';

export default function SettingsScreen() {
  const insets = useSafeAreaInsets();
  const me = useMe();
  return (
    // Own ScrollView so a tap on Save lands while the keyboard is still up.
    <Screen back title="Settings" scroll={false}>
      <ScrollView
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 4, paddingBottom: insets.bottom + 48 }}>
        <QueryState q={me} loadingLabel="Fetching your card…">
          {(user) => <SettingsForm me={user} />}
        </QueryState>

        <SectionHeader title="Appearance" style={{ marginTop: 36 }} />
        <Panel style={{ paddingVertical: 4 }}>
          <AppearancePicker />
        </Panel>

        <AccountSection />

        <SectionHeader title="About" />
        <Panel>
          <Txt variant="h3">Innings</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
            The fan&apos;s cricket diary. Every match you watched, on a ticket.
          </Txt>
          <Txt variant="monoSmall" tone="soft" style={{ marginTop: 12 }}>
            Version {Constants.expoConfig?.version ?? '1.0.0'}
          </Txt>
          <Txt variant="monoSmall" tone="faint" selectable style={{ marginTop: 2 }}>
            API {API_URL}
          </Txt>
        </Panel>
      </ScrollView>
    </Screen>
  );
}
