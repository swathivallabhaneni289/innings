import { Alert, Pressable } from 'react-native';

import { QueryState, Screen, Txt } from '@/components';
import { signOut, useAccount, useMe } from '@/data';
import { OnboardingHeader } from '@/features/onboarding/OnboardingHeader';
import { ProfileStep } from '@/features/onboarding/ProfileStep';
import { fonts } from '@/theme';

/** Sign-up step 1: the fan card. "Sign out" is the way back for someone who meant to use another account. */
export default function OnboardingProfileScreen() {
  const me = useMe();
  const account = useAccount();

  const confirmSignOut = () =>
    Alert.alert('Sign out?', 'You can sign in again with this email or another one.', [
      { text: 'Stay', style: 'cancel' },
      { text: 'Sign out', style: 'destructive', onPress: signOut },
    ]);

  return (
    <Screen
      scroll={false}
      header={
        <OnboardingHeader
          step={1}
          title="Your fan card"
          right={
            <Pressable onPress={confirmSignOut} hitSlop={12} accessibilityRole="button" style={{ marginLeft: 12, paddingVertical: 8 }}>
              <Txt variant="bodySmall" tone="red" style={{ fontFamily: fonts.uiBold }}>
                Sign out
              </Txt>
            </Pressable>
          }
        />
      }>
      <QueryState q={me} loadingLabel="Printing your card…">
        {(user) => <ProfileStep me={user} account={account.data} />}
      </QueryState>
    </Screen>
  );
}
