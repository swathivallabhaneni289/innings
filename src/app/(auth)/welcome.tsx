import { router } from 'expo-router';
import { useEffect } from 'react';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Button, Paper, Rule, Txt } from '@/components';
import { WelcomeTicket } from '@/features/auth/WelcomeTicket';
import { useSignInFlow } from '@/features/auth/flow';
import { fonts, useTheme } from '@/theme';

const BEATS = [
  { tag: 'Toss', title: 'Call it before the first ball.', body: 'Winner, top run-scorer, top wicket-taker. Calls lock when play starts.' },
  { tag: 'Stumps', title: 'Rate it after the last.', body: 'Half-balls out of five for the match, marks out of ten for the players.' },
  { tag: 'Stub', title: 'Keep the stub.', body: 'Every match you watched, printed on a ticket and kept in your collection.' },
];

/** The front gate: the ticket, the pitch in one line, three beats, and the way in. */
export default function WelcomeScreen() {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const reset = useSignInFlow((s) => s.reset);
  // A fresh visit to the gate (first launch, or after signing out) starts sign-in from scratch.
  useEffect(() => reset(), [reset]);

  const start = () => router.push('/phone');

  return (
    <Paper>
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={[styles.content, { paddingTop: insets.top + 10, paddingBottom: Math.max(insets.bottom, 16) + 8 }]}>
        <Txt accessibilityRole="header" style={[styles.wordmark, { color: t.ink }]}>
          innings
        </Txt>

        <WelcomeTicket style={{ marginTop: 18 }} />

        <Txt variant="h1" style={{ marginTop: 30 }}>
          A matchday diary for cricket fans.
        </Txt>

        <View style={{ marginTop: 18 }}>
          {BEATS.map((b, i) => (
            <View key={b.tag}>
              {i > 0 ? <Rule /> : null}
              <View style={styles.beat}>
                <Txt variant="label" tone="soft" style={styles.tag}>
                  {b.tag}
                </Txt>
                <View style={{ flex: 1 }}>
                  <Txt variant="h3">{b.title}</Txt>
                  <Txt variant="bodySmall" tone="soft" style={{ marginTop: 2 }}>
                    {b.body}
                  </Txt>
                </View>
              </View>
            </View>
          ))}
        </View>

        <View style={{ flex: 1, minHeight: 28 }} />
        <Button label="Get started" size="lg" onPress={start} />
        <Pressable onPress={start} accessibilityRole="button" style={({ pressed }) => [styles.secondary, { opacity: pressed ? 0.6 : 1 }]}>
          <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label }}>
            I already have an account
          </Txt>
        </Pressable>
      </ScrollView>
    </Paper>
  );
}

const styles = StyleSheet.create({
  content: { flexGrow: 1, paddingHorizontal: 20 },
  wordmark: { fontFamily: fonts.heading, fontSize: 26, lineHeight: 30, letterSpacing: -0.6 },
  beat: { flexDirection: 'row', alignItems: 'flex-start', paddingVertical: 12 },
  tag: { width: 70, marginTop: 3 },
  secondary: { minHeight: 48, alignItems: 'center', justifyContent: 'center', marginTop: 4 },
});
