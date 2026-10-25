import { Stack } from 'expo-router';

import { useTheme } from '@/theme';

// Deep links into the app (a match, a stub) open over the tabs, so Back always has somewhere to go.
export const unstable_settings = { initialRouteName: '(tabs)' };

/** The signed-in app: the tabs, every pushed screen, and the Log flow as a full-screen modal. */
export default function AppLayout() {
  const t = useTheme();
  return (
    <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: t.paper } }}>
      <Stack.Screen name="(tabs)" />
      <Stack.Screen name="log/index" options={{ presentation: 'fullScreenModal', animation: 'slide_from_bottom' }} />
    </Stack>
  );
}
