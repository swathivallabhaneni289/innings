import { Stack } from 'expo-router';

import { useTheme } from '@/theme';

export const unstable_settings = { initialRouteName: 'welcome' };

/**
 * Signed out. Welcome → create an account (email and password) → the emailed code; or Welcome → sign in, with
 * "Forgot password?" → the emailed reset code → a new password.
 */
export default function AuthLayout() {
  const t = useTheme();
  return (
    <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: t.paper } }}>
      <Stack.Screen name="welcome" />
      <Stack.Screen name="sign-up" />
      <Stack.Screen name="sign-in" />
      <Stack.Screen name="confirm" />
      <Stack.Screen name="forgot" />
      <Stack.Screen name="reset-code" />
      <Stack.Screen name="new-password" />
    </Stack>
  );
}
