import { Stack } from 'expo-router';

import { useTheme } from '@/theme';

export const unstable_settings = { initialRouteName: 'profile' };

/** Signed in, not onboarded yet: the fan card, then finding friends. */
export default function OnboardingLayout() {
  const t = useTheme();
  return (
    <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: t.paper } }}>
      <Stack.Screen name="profile" />
      <Stack.Screen name="friends" />
    </Stack>
  );
}
