import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { router } from 'expo-router';
import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { Button, Empty, Panel, Txt } from '@/components';
import { matchTitle, useActions } from '@/data';
import { competitionShort } from '@/domain/constants';
import type { Match } from '@/domain/types';
import { fonts, useTheme } from '@/theme';
import { competitionNoYear, matchYear } from './top5';

/** "World Cup · Final · 2011": what the match was, in the fine print under its teams. */
const aboutLine = (m: Match) =>
  [m.classic ? competitionNoYear(m.classic.competitionName) : competitionShort(m.competitionId), m.stage, matchYear(m)]
    .filter(Boolean)
    .join(' · ');

/**
 * The fan's Top 5, best first, to put right: move a game up or down, or take one out. Nothing changes until Save, so
 * Back leaves the Top 5 as it was.
 */
export function Top5Editor({ top5 }: { top5: Match[] }) {
  const t = useTheme();
  const { setTop5 } = useActions();
  const [order, setOrder] = useState(top5);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const changed = order.map((m) => m.id).join() !== top5.map((m) => m.id).join();

  const move = (i: number, by: -1 | 1) => {
    Haptics.selectionAsync();
    const next = [...order];
    [next[i], next[i + by]] = [next[i + by], next[i]];
    setOrder(next);
  };
  const remove = (i: number) => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
    setOrder(order.filter((_, j) => j !== i));
  };
  const save = async () => {
    setBusy(true);
    setError(null);
    try {
      await setTop5(order.map((m) => m.id));
      router.back();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not save your Top 5.');
    } finally {
      setBusy(false);
    }
  };

  if (top5.length === 0) {
    return (
      <Panel style={{ paddingVertical: 4 }}>
        <Empty title="No Top 5 yet" body="Open any match and tap “Pin to Top 5”. Five tickets, your call." />
      </Panel>
    );
  }

  return (
    <View>
      <Txt variant="body" tone="soft" style={{ marginBottom: 16 }}>
        Best first. Move a game up or down, or take one out, then save.
      </Txt>
      {order.map((m, i) => (
        <Panel key={m.id} style={styles.row}>
          <Txt variant="display" style={styles.num}>
            {i + 1}
          </Txt>
          <View style={{ flex: 1, marginRight: 4 }}>
            <Txt variant="h3" numberOfLines={1}>
              {matchTitle(m)}
            </Txt>
            <Txt variant="bodySmall" tone="soft" numberOfLines={2} style={{ fontFamily: fonts.ui }}>
              {aboutLine(m)}
            </Txt>
          </View>
          <Tool icon="arrow-up" label={`Move ${matchTitle(m)} up`} disabled={i === 0} onPress={() => move(i, -1)} />
          <Tool icon="arrow-down" label={`Move ${matchTitle(m)} down`} disabled={i === order.length - 1} onPress={() => move(i, 1)} />
          <Tool icon="x" label={`Take ${matchTitle(m)} out`} tint={t.red} onPress={() => remove(i)} />
        </Panel>
      ))}
      {order.length === 0 ? (
        <Panel style={{ paddingVertical: 4 }}>
          <Empty title="All five taken out" body="Save to clear your Top 5, or go back to keep it as it was." />
        </Panel>
      ) : null}
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 4, fontFamily: fonts.ui }}>
          {error}
        </Txt>
      ) : null}
      <Button label="Save" loading={busy} disabled={!changed} onPress={save} style={{ marginTop: 12, minHeight: 50 }} />
    </View>
  );
}

/** A 44-point square button holding one icon. */
function Tool({ icon, label, onPress, disabled, tint }: { icon: 'arrow-up' | 'arrow-down' | 'x'; label: string; onPress: () => void; disabled?: boolean; tint?: string }) {
  const t = useTheme();
  return (
    <Pressable onPress={onPress} disabled={disabled} accessibilityRole="button" accessibilityLabel={label} style={styles.tool}>
      <Feather name={icon} size={20} color={disabled ? t.rule : (tint ?? t.ink)} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 10, paddingRight: 4, marginBottom: 10 },
  num: { fontSize: 34, lineHeight: 38, width: 34 },
  tool: { width: 44, height: 44, alignItems: 'center', justifyContent: 'center' },
});
