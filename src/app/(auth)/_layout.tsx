import { Stack } from 'expo-router';

import { useTheme } from '@/theme';

export const unstable_settings = { initialRouteName: 'welcome' };

/** Signed out: Welcome → phone number → code. The country list slides up over the phone screen. */
export default function AuthLayout() {
  const t = useTheme();
  return (
    <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: t.paper } }}>
      <Stack.Screen name="welcome" />
      <Stack.Screen name="phone" />
      <Stack.Screen name="code" />
      <Stack.Screen name="country" options={{ presentation: 'fullScreenModal', animation: 'slide_from_bottom' }} />
    </Stack>
  );
}
