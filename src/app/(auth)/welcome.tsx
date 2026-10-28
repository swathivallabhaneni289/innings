import { router } from 'expo-router';
import { useEffect, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Button, Paper, Txt } from '@/components';
import { WelcomeStory } from '@/features/auth/WelcomeStory';
import { useSignInFlow } from '@/features/auth/flow';
import { entranceLead } from '@/lib/splash';
import { fonts, useTheme } from '@/theme';

/**
 * The front gate, on one screen with nothing to scroll: the wordmark, the app told as a story of five tickets,
 * and the way in at the bottom, ready from the first frame. Only when the system text is so large that the screen
 * can't hold it all does it scroll, so nothing is ever cut off.
 */
export default function WelcomeScreen() {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const reset = useSignInFlow((s) => s.reset);
  // On a cold launch the story waits for the splash to fade; after a sign-out it starts straight away.
  const [lead] = useState(entranceLead);
  // A fresh visit to the gate (first launch, or after signing out) starts sign-in from scratch.
  useEffect(() => reset(), [reset]);

  const start = () => router.push('/phone');

  return (
    <Paper>
      <ScrollView
        alwaysBounceVertical={false}
        showsVerticalScrollIndicator={false}
        contentContainerStyle={[styles.page, { paddingTop: insets.top + 10, paddingBottom: Math.max(insets.bottom, 12) + 4 }]}>
        <View>
          <Txt accessibilityRole="header" allowFontScaling={false} style={[styles.wordmark, { color: t.ink }]}>
            innings
          </Txt>
          <Txt variant="bodySmall" tone="soft" maxFontSizeMultiplier={1.25}>
            A matchday diary for cricket fans.
          </Txt>
        </View>

        <View style={styles.story}>
          <WelcomeStory lead={lead} />
        </View>

        <Button label="Get started" size="lg" onPress={start} maxFontSizeMultiplier={1.25} />
        <Pressable onPress={start} accessibilityRole="button" style={({ pressed }) => [styles.secondary, { opacity: pressed ? 0.6 : 1 }]}>
          <Txt variant="bodySmall" tone="soft" maxFontSizeMultiplier={1.25} style={{ fontFamily: fonts.label }}>
            I already have an account
          </Txt>
        </Pressable>
      </ScrollView>
    </Paper>
  );
}

const styles = StyleSheet.create({
  page: { flexGrow: 1, paddingHorizontal: 20 },
  wordmark: { fontFamily: fonts.heading, fontSize: 26, lineHeight: 30, letterSpacing: -0.6 },
  story: { flexGrow: 1, justifyContent: 'center', paddingVertical: 8 },
  secondary: { minHeight: 48, alignItems: 'center', justifyContent: 'center', marginTop: 4 },
});
