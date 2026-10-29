import * as Haptics from 'expo-haptics';
import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { Panel, Txt } from '@/components';
import { ApiError, useActions } from '@/data';
import type { CallView } from '@/domain/api';
import { pct } from '@/lib/format';
import { fonts, useTheme } from '@/theme';

/** Predictions during the match: every call on it, newest first. */
export function LiveCalls({ calls }: { calls: CallView[] }) {
  return (
    <View style={{ gap: 12 }}>
      {calls.map((cv) => (
        <CallCard key={cv.call.id} view={cv} />
      ))}
    </View>
  );
}

/** One call: the question, its options with how fans split, and where it stands. Tap an option while it's open. */
export function CallCard({ view, compact = false }: { view: CallView; compact?: boolean }) {
  const t = useTheme();
  const { pickCall } = useActions();
  const [picked, setPicked] = useState<string | null>(null); // optimistic until the page refetches
  const [error, setError] = useState<string | null>(null);
  const { call, counts, points } = view;
  const mine = picked ?? view.mine;
  const open = call.state === 'open';
  const total = Object.values(counts).reduce((a, b) => a + b, 0) + (picked && picked !== view.mine ? 1 : 0);

  const pick = async (optionId: string) => {
    if (!open || optionId === mine) return;
    setError(null);
    setPicked(optionId);
    Haptics.selectionAsync().catch(() => {});
    try {
      await pickCall(call.matchId, call.id, optionId);
    } catch (e) {
      setPicked(null);
      setError(e instanceof ApiError && e.status === 409 ? 'Too late, that one has locked.' : 'That didn’t go through. Try again.');
    }
  };

  const status =
    call.state === 'open'
      ? `Locks ${call.locksAt}`
      : call.state === 'locked'
        ? 'Locked · waiting on the result'
        : call.state === 'settled'
          ? `${call.result}${mine ? (points > 0 ? ` · You called it, +${points}` : ' · Not this time') : ''}`
          : `No result · ${call.result}`;

  return (
    <Panel style={compact ? { paddingVertical: 12 } : undefined}>
      <View style={styles.head}>
        <Txt variant="label" tone={open ? 'red' : 'soft'}>
          {open ? 'Open' : call.state === 'settled' ? 'Settled' : call.state === 'void' ? 'Void' : 'Locked'}
        </Txt>
      </View>
      <Txt variant="h3" style={{ marginTop: 4 }}>
        {call.question}
      </Txt>
      <View style={[styles.options, { marginTop: compact ? 8 : 12 }]}>
        {call.options.map((o) => {
          const chosen = mine === o.id;
          const right = call.state === 'settled' && call.answerId === o.id;
          const n = (counts[o.id] ?? 0) + (picked === o.id && view.mine !== o.id ? 1 : 0) - (view.mine === o.id && picked && picked !== o.id ? 1 : 0);
          return (
            <Pressable
              key={o.id}
              accessibilityRole="button"
              accessibilityState={{ selected: chosen, disabled: !open }}
              accessibilityLabel={`${o.label}${total > 0 ? `, ${pct(n / total)} of fans` : ''}${right ? ', the right answer' : ''}`}
              disabled={!open}
              onPress={() => pick(o.id)}
              style={({ pressed }) => [
                styles.option,
                {
                  borderColor: right ? t.green : chosen ? t.ink : t.rule,
                  backgroundColor: chosen ? t.ink : right ? t.greenSoft : t.card,
                  opacity: pressed ? 0.8 : 1,
                },
              ]}>
              <Txt style={{ fontFamily: fonts.label, fontSize: 14, color: chosen ? t.onInk : t.ink }} numberOfLines={1}>
                {o.label}
              </Txt>
              {total > 0 ? (
                <Txt variant="monoSmall" style={{ color: chosen ? t.onInk : t.inkSoft, marginTop: 2 }}>
                  {pct(n / total)}
                </Txt>
              ) : null}
            </Pressable>
          );
        })}
      </View>
      <Txt variant="bodySmall" tone={call.state === 'settled' && points > 0 ? 'green' : 'soft'} style={{ marginTop: 10 }}>
        {status}
      </Txt>
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 4 }}>
          {error}
        </Txt>
      ) : null}
    </Panel>
  );
}

const styles = StyleSheet.create({
  head: { flexDirection: 'row', alignItems: 'center' },
  options: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  option: { borderWidth: 1, paddingHorizontal: 12, paddingVertical: 8, minHeight: 44, minWidth: 72, justifyContent: 'center' },
});
