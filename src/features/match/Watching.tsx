import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';

import { Button, Panel, Txt } from '@/components';
import { useActions } from '@/data';
import { WATCH_PLACES } from '@/domain/constants';
import type { WatchPlace } from '@/domain/types';
import { fonts, useTheme } from '@/theme';

/** "I'm watching": opens an inline chooser of the seven watch places, shows the chosen one, and lets you clear it. */
export function WatchingButton({ matchId, watching }: { matchId: string; watching: WatchPlace | null }) {
  const t = useTheme();
  const actions = useActions();
  const [open, setOpen] = useState(false);
  const [picked, setPicked] = useState<WatchPlace | null | undefined>(undefined); // optimistic until the page refetches
  const [busy, setBusy] = useState<WatchPlace | 'clear' | null>(null);
  const [error, setError] = useState<string | null>(null);

  // A new answer from the server (the page refetched) replaces the optimistic pick.
  const [seen, setSeen] = useState(watching);
  if (seen !== watching) {
    setSeen(watching);
    setPicked(undefined);
  }
  const current = picked === undefined ? watching : picked;
  const place = WATCH_PLACES.find((w) => w.id === current);

  const save = async (next: WatchPlace | null) => {
    setBusy(next ?? 'clear');
    setError(null);
    try {
      await actions.setWatching(matchId, next);
      Haptics.selectionAsync().catch(() => {});
      setPicked(next);
      setOpen(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't save that. Try again.");
    } finally {
      setBusy(null);
    }
  };

  if (!open && !place) {
    return (
      <View>
        <Button
          label="I'm watching"
          icon={<Feather name="tv" size={17} color={t.onInk} />}
          onPress={() => setOpen(true)}
          style={{ minHeight: 50 }}
        />
        {error ? <ErrorLine text={error} /> : null}
      </View>
    );
  }

  if (!open && place) {
    return (
      <Panel>
        <View style={styles.chosen}>
          <Txt style={{ fontSize: 26, marginRight: 12 }}>{place.emoji}</Txt>
          <View style={{ flex: 1 }}>
            <Txt variant="bodySmall" tone="soft">
              You&apos;re watching from
            </Txt>
            <Txt variant="h3">{place.label}</Txt>
          </View>
          <Button label="Change" kind="ghost" size="sm" onPress={() => setOpen(true)} style={styles.ghost} />
          <Button label="Clear" kind="ghost" size="sm" loading={busy === 'clear'} onPress={() => save(null)} style={styles.ghost} />
        </View>
        {error ? <ErrorLine text={error} /> : null}
      </Panel>
    );
  }

  return (
    <Panel>
      <Txt variant="h3" style={{ marginBottom: 12 }}>
        Where are you watching from?
      </Txt>
      <View style={styles.grid}>
        {WATCH_PLACES.map((w) => {
          const selected = current === w.id;
          return (
            <Pressable
              key={w.id}
              accessibilityRole="button"
              accessibilityState={{ selected }}
              disabled={busy !== null}
              onPress={() => save(w.id)}
              style={({ pressed }) => [
                styles.option,
                { borderColor: selected ? t.ink : t.rule, backgroundColor: selected ? t.ink : 'transparent', opacity: pressed ? 0.7 : 1 },
              ]}>
              <Txt style={{ fontSize: 16, marginRight: 8 }}>{w.emoji}</Txt>
              <Txt variant="bodySmall" numberOfLines={1} style={{ flex: 1, fontFamily: fonts.ui, color: selected ? t.onInk : t.ink }}>
                {w.label}
              </Txt>
              {busy === w.id ? <ActivityIndicator size="small" color={selected ? t.onInk : t.ink} /> : null}
            </Pressable>
          );
        })}
      </View>
      <View style={styles.footer}>
        {current ? (
          <Button label="Not watching" kind="ghost" size="sm" loading={busy === 'clear'} onPress={() => save(null)} style={styles.ghost} />
        ) : (
          <View />
        )}
        <Button label="Cancel" kind="ghost" size="sm" onPress={() => setOpen(false)} style={styles.ghost} />
      </View>
      {error ? <ErrorLine text={error} /> : null}
    </Panel>
  );
}

function ErrorLine({ text }: { text: string }) {
  return (
    <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
      {text}
    </Txt>
  );
}

const styles = StyleSheet.create({
  chosen: { flexDirection: 'row', alignItems: 'center' },
  grid: { flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between' },
  option: { width: '48.5%', minHeight: 46, flexDirection: 'row', alignItems: 'center', borderWidth: 1, paddingHorizontal: 10, paddingVertical: 10, marginBottom: 8 },
  ghost: { minHeight: 44, paddingHorizontal: 8 },
  footer: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginTop: 2 },
});
