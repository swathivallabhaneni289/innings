import { useState } from 'react';
import { StyleSheet, TextInput, View } from 'react-native';

import { BallRating, Button, Panel, Txt } from '@/components';
import { useActions, useUserProfile } from '@/data';
import type { Match, MatchLog } from '@/domain/types';
import { lastPlaceOf, relog } from '@/features/log/draft';
import { fonts, useTheme } from '@/theme';

/**
 * AFTER: the fan's own review, typed right under Reviews. A review lives on the fan's log, so posting saves the log:
 * one they already have keeps everything else, and a first one needs a rating (the balls show here until there is one)
 * and starts where they logged from last, else Home. The saved review stays in the box, as the Popular list, sorted by
 * likes, may bury it; emptying the box and saving takes it off.
 */
export function ReviewBox({ match, log }: { match: Match; log: MatchLog | null }) {
  const t = useTheme();
  const me = useUserProfile('me');
  const { saveLog } = useActions();
  const saved = log?.review ?? '';
  const [text, setText] = useState(saved);
  const [shownFor, setShownFor] = useState(saved);
  const [rating, setRating] = useState(0);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<{ text: string; error: boolean } | null>(null);

  // A review saved elsewhere (the Log flow) replaces what the box shows. Done during render, like the Log flow's prefill.
  if (shownFor !== saved) {
    setShownFor(saved);
    setText(saved);
  }

  const typed = text.trim();
  const changed = typed !== saved.trim();
  const needsRating = !log && rating === 0;
  const canPost = !!me.data && !busy && changed && !needsRating && (!!log || typed.length > 0);

  const post = async () => {
    if (!me.data || !canPost) return;
    setBusy(true);
    setNote(null);
    try {
      const place = lastPlaceOf(me.data.logs.map((fi) => fi.log)) ?? 'home';
      await saveLog(relog(log, match.id, place, log ? { review: text } : { rating, review: text }));
      setNote({ text: typed ? 'Posted. Fans see it under Reviews.' : 'Review taken off.', error: false });
    } catch (e) {
      setNote({ text: e instanceof Error ? e.message : 'Could not post your review.', error: true });
    } finally {
      setBusy(false);
    }
  };

  const hint = note?.text ?? (needsRating && typed ? 'Rate it to post your review.' : null);
  return (
    <Panel style={{ marginBottom: 16 }}>
      <Txt variant="h3">{saved ? 'Your review' : 'Write a review'}</Txt>
      {!log ? (
        <View style={styles.rate}>
          <Txt variant="bodySmall" tone="soft" style={{ marginRight: 12, fontFamily: fonts.ui }}>
            Your rating
          </Txt>
          <BallRating value={rating} size={24} gap={6} onChange={setRating} />
        </View>
      ) : null}
      <TextInput
        value={text}
        onChangeText={(v) => {
          setText(v);
          setNote(null);
        }}
        placeholder="What will you remember about this one?"
        placeholderTextColor={t.inkFaint}
        multiline
        maxLength={2000}
        textAlignVertical="top"
        accessibilityLabel="Your review"
        style={[styles.input, { color: t.ink, borderColor: t.rule }]}
      />
      <View style={styles.foot}>
        <Txt variant="bodySmall" tone={note?.error || (!note && hint) ? 'red' : 'soft'} style={{ flex: 1, marginRight: 12, fontFamily: fonts.ui }}>
          {hint ?? ''}
        </Txt>
        <Button label={saved ? 'Save changes' : 'Post review'} size="sm" loading={busy} disabled={!canPost} onPress={post} style={{ minHeight: 44 }} />
      </View>
    </Panel>
  );
}

const styles = StyleSheet.create({
  rate: { flexDirection: 'row', alignItems: 'center', marginTop: 12 },
  input: {
    fontFamily: fonts.serif,
    fontSize: 16,
    lineHeight: 23,
    minHeight: 110,
    padding: 12,
    marginTop: 12,
    borderWidth: StyleSheet.hairlineWidth,
  },
  foot: { flexDirection: 'row', alignItems: 'center', marginTop: 12 },
});
