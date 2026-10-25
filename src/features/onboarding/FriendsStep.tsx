import { useState } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Button, Empty, Panel, QueryState, Rule, SectionHeader, Txt } from '@/components';
import { finishOnboarding, useActions, useFriends } from '@/data';
import { errorText } from '@/features/auth/errors';
import { useTheme } from '@/theme';
import { ContactsFinder } from './ContactsFinder';
import { FanRow } from './FanRow';

/**
 * Step 2, "Find your people": fans from the address book (matched by hash), the simulated fans to follow, and
 * "Start watching", which finishes sign-up. The root guard then opens the app on Home.
 */
export function FriendsStep() {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const friends = useFriends();
  const { followAll } = useActions();
  const [matched, setMatched] = useState<Set<string>>(() => new Set());
  const [followingAll, setFollowingAll] = useState(false);
  const [followError, setFollowError] = useState<string | null>(null);
  const [finishing, setFinishing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // A contact who's also in the list below shows once, under contacts.
  const fans = (friends.data?.friends ?? []).filter((f) => !matched.has(f.user.id));
  const toFollow = fans.filter((f) => !f.following);

  const onFollowAll = async () => {
    if (followingAll || toFollow.length === 0) return;
    setFollowingAll(true);
    setFollowError(null);
    try {
      // The endpoint toggles, so only fans not followed yet are sent.
      await followAll(toFollow.map((f) => f.user.id));
    } catch (e) {
      setFollowError(errorText(e, "Couldn't follow everyone. Try again."));
    } finally {
      setFollowingAll(false);
    }
  };

  const finish = async () => {
    if (finishing) return;
    setFinishing(true);
    setError(null);
    try {
      await finishOnboarding();
    } catch (e) {
      setError(errorText(e, "Couldn't finish sign-up. Try again."));
      setFinishing(false);
    }
  };

  return (
    <View style={{ flex: 1 }}>
      <ScrollView showsVerticalScrollIndicator={false} contentContainerStyle={{ paddingHorizontal: 20, paddingBottom: 28 }}>
        <Txt variant="body" tone="soft">
          Cricket&apos;s better with rivals. Follow a few fans to fill your feed and race them on the Bragging Board.
        </Txt>

        <SectionHeader title="From your contacts" />
        <ContactsFinder Row={FanRow} onMatched={setMatched} />

        <SectionHeader
          title="Fans to follow"
          action={toFollow.length > 1 ? (followingAll ? 'Following…' : 'Follow all') : undefined}
          onAction={onFollowAll}
        />
        {followError ? (
          <Txt variant="bodySmall" tone="red" style={{ marginBottom: 8 }}>
            {followError}
          </Txt>
        ) : null}
        <QueryState q={friends} loadingLabel="Checking the dressing room…">
          {() =>
            fans.length === 0 ? (
              <Panel>
                <Empty title="Empty pavilion" body="No fans to suggest just now. Find more from the Friends screen later." />
              </Panel>
            ) : (
              <Panel style={{ paddingVertical: 2 }}>
                {fans.map((f, i) => (
                  <FanRow key={f.user.id} user={f.user} following={f.following} last={i === fans.length - 1} />
                ))}
              </Panel>
            )
          }
        </QueryState>
      </ScrollView>

      <View style={[styles.footer, { paddingBottom: Math.max(insets.bottom, 12), backgroundColor: t.paper }]}>
        <Rule />
        {error ? (
          <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
            {error}
          </Txt>
        ) : null}
        <Button label="Start watching" size="lg" onPress={finish} loading={finishing} style={{ marginTop: 10 }} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  footer: { paddingHorizontal: 20 },
});
