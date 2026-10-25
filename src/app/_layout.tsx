import { IBMPlexMono_400Regular, IBMPlexMono_500Medium, IBMPlexMono_600SemiBold } from '@expo-google-fonts/ibm-plex-mono';
import { BigShouldersDisplay_800ExtraBold, BigShouldersDisplay_900Black } from '@expo-google-fonts/big-shoulders-display';
import {
  SchibstedGrotesk_400Regular,
  SchibstedGrotesk_400Regular_Italic,
  SchibstedGrotesk_500Medium,
  SchibstedGrotesk_600SemiBold,
  SchibstedGrotesk_700Bold,
  SchibstedGrotesk_800ExtraBold,
} from '@expo-google-fonts/schibsted-grotesk';
import { QueryClientProvider } from '@tanstack/react-query';
import { useFonts } from 'expo-font';
import { Stack } from 'expo-router';
import * as SplashScreen from 'expo-splash-screen';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import { View } from 'react-native';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { Button, ErrorState, Loading, Paper } from '@/components';
import { loadSession, queryClient, signOut, useAccount, useBootstrap, useSession } from '@/data';
import { usePrefsHydrated } from '@/data/store';
import { useQaDriver } from '@/lib/qaDriver';
import { ThemeProvider, useTheme } from '@/theme';

SplashScreen.preventAutoHideAsync().catch(() => {});

export default function RootLayout() {
  const [fontsLoaded] = useFonts({
    BigShouldersDisplay_800ExtraBold,
    BigShouldersDisplay_900Black,
    SchibstedGrotesk_400Regular,
    SchibstedGrotesk_400Regular_Italic,
    SchibstedGrotesk_500Medium,
    SchibstedGrotesk_600SemiBold,
    SchibstedGrotesk_700Bold,
    SchibstedGrotesk_800ExtraBold,
    IBMPlexMono_400Regular,
    IBMPlexMono_500Medium,
    IBMPlexMono_600SemiBold,
  });
  const hydrated = usePrefsHydrated();
  const sessionKnown = useSession((s) => s.status !== 'loading');

  useEffect(() => {
    loadSession();
  }, []);

  // The splash stays up until the fonts, the prefs and the stored session have all been read.
  useEffect(() => {
    if (fontsLoaded && hydrated && sessionKnown) SplashScreen.hideAsync().catch(() => {});
  }, [fontsLoaded, hydrated, sessionKnown]);

  if (!fontsLoaded || !hydrated || !sessionKnown) return null;

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <SafeAreaProvider>
        <QueryClientProvider client={queryClient}>
          <ThemeProvider>
            <Gate />
          </ThemeProvider>
        </QueryClientProvider>
      </SafeAreaProvider>
    </GestureHandlerRootView>
  );
}

/**
 * Picks the side of the gate. Exactly one group is open at a time, so Expo Router lands the fan on it by itself:
 * (auth) when signed out, onboarding until the account is onboarded, then (app). A session found at launch is
 * bootstrapped (catalog, me, account) before any signed-in screen renders.
 */
function Gate() {
  const t = useTheme();
  const status = useSession((s) => s.status);
  const ready = useSession((s) => s.ready);
  const account = useAccount();
  const boot = useBootstrap();
  const { run } = boot;
  useQaDriver();
  useEffect(() => {
    if (status === 'signedIn' && !ready) run();
  }, [status, ready, run]);

  const signedIn = status === 'signedIn';
  const onboarded = !!account.data?.onboarded;
  const bar = <StatusBar style={t.mode === 'paper' ? 'dark' : 'light'} />;
  const error = boot.error ?? (account.error as Error | null);

  if (signedIn && (!ready || !account.data)) {
    return (
      <Paper>
        {bar}
        {error && !boot.running ? (
          <View style={{ flex: 1 }}>
            <ErrorState error={error} onRetry={run} />
            <Button label="Sign out" kind="ghost" size="sm" onPress={signOut} style={{ alignSelf: 'center', marginBottom: 48, minHeight: 44 }} />
          </View>
        ) : (
          <Loading label="Walking out to bat…" />
        )}
      </Paper>
    );
  }

  return (
    <>
      {bar}
      <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: t.paper } }}>
        <Stack.Protected guard={!signedIn}>
          <Stack.Screen name="(auth)" />
        </Stack.Protected>
        <Stack.Protected guard={signedIn && !onboarded}>
          <Stack.Screen name="onboarding" />
        </Stack.Protected>
        <Stack.Protected guard={signedIn && onboarded}>
          <Stack.Screen name="(app)" />
        </Stack.Protected>
      </Stack>
    </>
  );
}
