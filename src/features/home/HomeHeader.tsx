import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { Pressable, StyleSheet, View } from 'react-native';

import { Txt } from '@/components';
import { useNow } from '@/data';
import { fonts, useTheme } from '@/theme';

const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

/** The masthead: lowercase "innings", today's date, search and the notifications bell. */
export function HomeHeader({ unread }: { unread: number }) {
  const t = useTheme();
  const now = new Date(useNow(60_000));
  return (
    <View style={styles.bar}>
      <View style={{ flex: 1 }}>
        <Txt accessibilityRole="header" style={[styles.wordmark, { color: t.ink }]}>
          innings
        </Txt>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui, marginTop: 1 }}>
          {DAYS[now.getDay()]} {now.getDate()} {MONTHS[now.getMonth()]}
        </Txt>
      </View>
      <Pressable
        onPress={() => router.push('/search')}
        style={({ pressed }) => [styles.icon, { opacity: pressed ? 0.6 : 1 }]}
        accessibilityRole="button"
        accessibilityLabel="Search">
        <Feather name="search" size={21} color={t.ink} />
      </Pressable>
      <Pressable
        onPress={() => router.push('/notifications')}
        style={({ pressed }) => [styles.icon, { opacity: pressed ? 0.6 : 1 }]}
        accessibilityRole="button"
        accessibilityLabel={unread > 0 ? `Notifications, ${unread} new` : 'Notifications'}>
        <Feather name="bell" size={21} color={t.ink} />
        {unread > 0 ? (
          <View style={[styles.badge, { backgroundColor: t.red }]}>
            <Txt tone="onRed" style={styles.badgeText}>
              {unread > 9 ? '9+' : unread}
            </Txt>
          </View>
        ) : null}
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  bar: { flexDirection: 'row', alignItems: 'center', paddingLeft: 20, paddingRight: 10, paddingTop: 10, paddingBottom: 6 },
  wordmark: { fontFamily: fonts.heading, fontSize: 26, lineHeight: 30, letterSpacing: -0.6 },
  icon: { width: 44, height: 44, alignItems: 'center', justifyContent: 'center' },
  badge: { position: 'absolute', top: 6, right: 5, minWidth: 16, height: 16, paddingHorizontal: 3, alignItems: 'center', justifyContent: 'center' },
  badgeText: { fontFamily: fonts.monoMedium, fontSize: 10, lineHeight: 12 },
});
