import { router } from 'expo-router';

import { Screen } from '@/components';
import { FriendsStep } from '@/features/onboarding/FriendsStep';
import { OnboardingHeader } from '@/features/onboarding/OnboardingHeader';

/** Sign-up step 2: find friends from contacts, follow a few fans, then start watching. */
export default function OnboardingFriendsScreen() {
  return (
    <Screen
      scroll={false}
      header={<OnboardingHeader step={2} title="Find your people" onBack={() => (router.canGoBack() ? router.back() : router.replace('/onboarding/profile'))} />}>
      <FriendsStep />
    </Screen>
  );
}
