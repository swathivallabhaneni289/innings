import * as Haptics from 'expo-haptics';
import { Pressable, StyleSheet, View } from 'react-native';

import { CricketBall, Txt } from '@/components';
import { usePrefs } from '@/data';
import type { ThemePreference } from '@/domain/types';
import { useTheme } from '@/theme';

const OPTIONS: { id: ThemePreference; label: string; blurb: string }[] = [
  { id: 'paper', label: 'Day', blurb: 'Bright stands and white tickets, for the afternoon session.' },
  { id: 'floodlights', label: 'Floodlights', blurb: 'Dark mode, for night matches under the lights.' },
  { id: 'auto', label: 'Match my phone', blurb: 'Day by day, floodlights when your phone goes dark.' },
];

/** Day / Floodlights / Match my phone. Applies straight away; stored on the device. */
export function AppearancePicker() {
  const t = useTheme();
  const theme = usePrefs((s) => s.theme);
  const setTheme = usePrefs((s) => s.setTheme);
  return (
    <View accessibilityRole="radiogroup">
      {OPTIONS.map((o, i) => {
        const on = theme === o.id;
        return (
          <Pressable
            key={o.id}
            onPress={() => {
              Haptics.selectionAsync();
              setTheme(o.id);
            }}
            accessibilityRole="radio"
            accessibilityState={{ checked: on }}
            style={({ pressed }) => [
              styles.row,
              { borderTopColor: t.rule, borderTopWidth: i === 0 ? 0 : StyleSheet.hairlineWidth, opacity: pressed ? 0.75 : 1 },
            ]}>
            <CricketBall size={20} fill={on ? 'full' : 'empty'} />
            <View style={{ flex: 1, marginLeft: 12 }}>
              <Txt variant="h3" tone={on ? 'ink' : 'soft'}>
                {o.label}
              </Txt>
              <Txt variant="bodySmall" tone="soft">
                {o.blurb}
              </Txt>
            </View>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 12, minHeight: 44 },
});
