import { Feather } from '@expo/vector-icons';
import { router } from 'expo-router';
import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import { BallRating, Button, MatchTicket, Panel, Txt } from '@/components';
import { useActions, useUserProfile } from '@/data';
import type { Match, MatchLog } from '@/domain/types';
import { lastPlaceOf, placeLine, relog, verdict } from '@/features/log/draft';
import { useTop5 } from '@/features/vault/top5';
import { fonts, useTheme } from '@/theme';

/**
 * AFTER: the fan's say on a finished match, right on its page, wherever they opened it from. Tapping the balls rates it
 * and saves at once: the first tap prints a stub, watched where they logged from last (else Home), and the line under
 * the balls says so. "Pin to Top 5" works for any match, not only the Vault's, and the Log flow is a tap away for the
 * players, place and friends. Once logged, the stub shows beneath.
 *
 * Both wait for the fan's profile: the Top 5 is saved whole, so a pin before it loads would wipe the other four. A
 * rating saves one tap at a time, so two quick taps can't land in the wrong order.
 */
export function YourTake({ match, log }: { match: Match; log: MatchLog | null }) {
  const t = useTheme();
  const me = useUserProfile('me');
  const { saveLog } = useActions();
  const top5 = useTop5((me.data?.top5 ?? []).map((m) => m.id));
  const [saving, setSaving] = useState<number | null>(null); // the rating on its way, shown until it lands
  const [error, setError] = useState<string | null>(null);
  const ready = !!me.data;
  const rating = saving ?? log?.rating ?? 0;
  const pinned = top5.isPinned(match.id);

  const rate = async (value: number) => {
    if (!me.data || saving !== null) return;
    setSaving(value);
    setError(null);
    try {
      const place = lastPlaceOf(me.data.logs.map((fi) => fi.log)) ?? 'home';
      await saveLog(relog(log, match.id, place, { rating: value }));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not save your rating.');
    } finally {
      setSaving(null);
    }
  };

  const line =
    error ??
    (saving !== null
      ? 'Saving…'
      : log
        ? `${placeLine(log)} · Edit your log to change it`
        : 'Tap a ball to rate it. It saves to your stubs straight away.');

  return (
    <View>
      <Panel>
        <Txt variant="h3">{log ? 'Your rating' : 'Rate it'}</Txt>
        <View style={styles.balls}>
          <BallRating value={rating} size={34} gap={10} onChange={ready && saving === null ? rate : undefined} />
          {rating > 0 ? (
            <Txt variant="mono" tone="soft" style={{ marginLeft: 12, flexShrink: 1 }} numberOfLines={1}>
              {verdict(rating)}
            </Txt>
          ) : null}
        </View>
        <Txt variant="bodySmall" tone={error ? 'red' : 'soft'} style={{ marginTop: 10, fontFamily: fonts.ui }}>
          {line}
        </Txt>
        <View style={styles.actions}>
          <Button
            label={log ? 'Edit your log' : 'Log this match'}
            size="sm"
            style={styles.action}
            icon={<Feather name={log ? 'edit-3' : 'plus'} size={14} color={t.onInk} />}
            onPress={() => router.push({ pathname: '/log', params: { matchId: match.id } })}
          />
          <Button
            label={pinned ? 'Pinned to Top 5' : 'Pin to Top 5'}
            kind={pinned ? 'ghost' : 'outline'}
            size="sm"
            loading={top5.busyId === match.id}
            disabled={!ready}
            style={styles.action}
            icon={<Feather name={pinned ? 'check' : 'bookmark'} size={13} color={pinned ? t.red : t.ink} />}
            onPress={() => top5.toggle(match.id)}
          />
        </View>
        {top5.message?.matchId === match.id ? (
          <Txt variant="bodySmall" tone="red" style={{ marginTop: 8, fontFamily: fonts.ui }}>
            {top5.message.text}
          </Txt>
        ) : null}
      </Panel>
      {log ? (
        <Pressable
          onPress={() => router.push(`/stub/${log.id}`)}
          accessibilityRole="button"
          accessibilityLabel="Open your ticket stub"
          style={{ marginTop: 12 }}>
          <MatchTicket match={match} rating={log.rating} stubNo={log.stubNo} instantClassic={log.instantClassic} />
        </Pressable>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  balls: { flexDirection: 'row', alignItems: 'center', marginTop: 12 },
  actions: { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', marginTop: 14, gap: 10 },
  action: { minHeight: 44 },
});
