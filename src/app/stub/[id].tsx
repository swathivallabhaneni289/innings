import { router, useLocalSearchParams } from 'expo-router';
import { useRef, useState, type ReactNode } from 'react';
import { KeyboardAvoidingView, Platform, Pressable, ScrollView, StyleSheet, TextInput, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Avatar, Button, CricketBall, MatchTicket, QueryState, Screen, SectionHeader, TeamBadge, Txt } from '@/components';
import { getPlayer, matchTitle, playerName, useActions, useLogItem, useNow } from '@/data';
import type { FeedItem } from '@/domain/api';
import type { Reply } from '@/domain/types';
import { placeLine } from '@/features/feed/FeedCard';
import { useKnownUsers, useUser, UserChip, type KnownUsers } from '@/features/feed/people';
import { StubActions } from '@/features/feed/StubActions';
import { relative, stubNo } from '@/lib/format';
import { fonts, useTheme } from '@/theme';

export default function StubScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const q = useLogItem(id);
  const item = q.data;
  return (
    <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={{ flex: 1 }}>
      <Screen
        back
        title={item ? matchTitle(item.match) : 'Stub'}
        kicker={item ? `${stubNo(item.log.stubNo)} · ${item.user.name.split(' ')[0]}'s stub` : undefined}
        scroll={false}>
        <QueryState q={q}>{(d) => <StubBody item={d} />}</QueryState>
      </Screen>
    </KeyboardAvoidingView>
  );
}

function StubBody({ item }: { item: FeedItem }) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const now = useNow(60_000);
  const known = useKnownUsers([item.user]);
  const { addReply } = useActions();
  const inputRef = useRef<TextInput>(null);
  const scrollRef = useRef<ScrollView>(null);
  const [text, setText] = useState('');
  const [posting, setPosting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { log, user, match } = item;

  const ratings = Object.entries(log.playerRatings).sort((a, b) => b[1] - a[1]);

  const post = async () => {
    const body = text.trim();
    if (!body || posting) return;
    setPosting(true);
    setError(null);
    try {
      await addReply(log.id, body);
      setText('');
      inputRef.current?.blur();
      setTimeout(() => scrollRef.current?.scrollToEnd({ animated: true }), 250);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Your reply didn’t go through. Try again.');
    } finally {
      setPosting(false);
    }
  };

  return (
    <View style={{ flex: 1 }}>
      <ScrollView
        ref={scrollRef}
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="interactive"
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.content}>
        <Pressable onPress={() => router.push('/user/' + user.id)} style={styles.author} accessibilityRole="button">
          <Avatar user={user} size={42} />
          <View style={{ marginLeft: 12, flex: 1 }}>
            <Txt variant="h3" numberOfLines={1}>
              {user.name}
            </Txt>
            <Txt variant="bodySmall" tone="soft" numberOfLines={1}>
              @{user.handle} · logged {relative(log.createdAt, now)}
            </Txt>
          </View>
        </Pressable>

        <Pressable onPress={() => router.push('/match/' + match.id)} accessibilityRole="button" accessibilityLabel={`Open ${matchTitle(match)}`}>
          <MatchTicket match={match} size="lg" rating={log.rating} stubNo={log.stubNo} instantClassic={log.instantClassic} style={{ marginTop: 16 }} />
        </Pressable>
        {match.result?.text ? (
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 8 }}>
            {match.result.text}
          </Txt>
        ) : null}

        {log.review ? (
          <Txt variant="review" style={styles.review}>
            “{log.review}”
          </Txt>
        ) : (
          <Txt variant="bodySmall" tone="faint" style={{ marginTop: 18 }}>
            No review. The balls say it all.
          </Txt>
        )}

        <View style={[styles.facts, { borderColor: t.rule }]}>
          <Fact label="Fan POTM">
            <Txt variant="bodySmall">{log.fanPotmId ? playerName(log.fanPotmId) : 'Not picked'}</Txt>
          </Fact>
          <Fact label="Watched">
            <Txt variant="bodySmall">📍 {placeLine(log)}</Txt>
          </Fact>
          {log.watchedWith.length > 0 ? (
            <Fact label="With">
              <View style={styles.with}>
                {log.watchedWith.map((uid) => (
                  <UserChip key={uid} id={uid} known={known} style={{ marginRight: 12, marginBottom: 4 }} />
                ))}
              </View>
            </Fact>
          ) : null}
        </View>

        <StubActions item={item} onReply={() => inputRef.current?.focus()} style={{ marginTop: 16 }} />

        <SectionHeader title="Player ratings" />
        {ratings.length === 0 ? (
          <Txt variant="bodySmall" tone="faint">
            No player ratings on this stub.
          </Txt>
        ) : (
          ratings.map(([pid, score], i) => {
            const p = getPlayer(pid);
            return (
              <View key={pid} style={[styles.ratingRow, i > 0 ? { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule } : null]}>
                <View style={{ width: 48 }}>{p ? <TeamBadge teamId={p.teamId} size="sm" /> : null}</View>
                <Txt variant="bodySmall" numberOfLines={1} style={{ flex: 1, marginLeft: 8 }}>
                  {playerName(pid)}
                </Txt>
                {pid === log.fanPotmId ? (
                  <View style={{ marginRight: 8 }}>
                    <CricketBall size={11} />
                  </View>
                ) : null}
                <Txt variant="mono">
                  {score}
                  <Txt variant="monoSmall" tone="faint">
                    /10
                  </Txt>
                </Txt>
              </View>
            );
          })
        )}

        <SectionHeader title={`Replies · ${item.replies.length}`} />
        {item.replies.length === 0 ? (
          <Txt variant="review" tone="soft">
            No replies yet. Start the chatter.
          </Txt>
        ) : (
          item.replies.map((r, i) => <ReplyRow key={r.id} reply={r} known={known} now={now} first={i === 0} />)
        )}
      </ScrollView>

      <View style={[styles.composer, { borderTopColor: t.rule, backgroundColor: t.paper, paddingBottom: Math.max(insets.bottom, 10) }]}>
        {error ? (
          <Txt variant="bodySmall" tone="red" style={{ marginBottom: 6 }}>
            {error}
          </Txt>
        ) : null}
        <View style={styles.composerRow}>
          <TextInput
            ref={inputRef}
            value={text}
            onChangeText={setText}
            placeholder={`Reply to ${user.name.split(' ')[0]}…`}
            placeholderTextColor={t.inkFaint}
            selectionColor={t.red}
            multiline
            maxLength={500}
            style={[styles.input, { color: t.ink, borderColor: t.rule, backgroundColor: t.card }]}
          />
          <Button label="Post" size="sm" onPress={post} loading={posting} disabled={!text.trim()} style={{ marginLeft: 10 }} />
        </View>
      </View>
    </View>
  );
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <View style={styles.fact}>
      <Txt variant="label" tone="soft" style={{ fontSize: 10, width: 74, marginTop: 2 }}>
        {label}
      </Txt>
      <View style={{ flex: 1 }}>{children}</View>
    </View>
  );
}

function ReplyRow({ reply, known, now, first }: { reply: Reply; known: KnownUsers; now: number; first: boolean }) {
  const t = useTheme();
  const user = useUser(reply.userId, known);
  return (
    <View style={[styles.reply, !first ? { borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: t.rule } : null]}>
      <Pressable onPress={() => router.push('/user/' + reply.userId)} hitSlop={6}>
        {user ? <Avatar user={user} size={28} /> : <View style={{ width: 28, height: 28, borderRadius: 14, backgroundColor: t.rule }} />}
      </Pressable>
      <View style={{ flex: 1, marginLeft: 10 }}>
        <View style={styles.replyHead}>
          <Txt variant="bodySmall" numberOfLines={1} style={{ fontFamily: fonts.uiBold, flexShrink: 1 }}>
            {user?.name ?? 'A fan'}
          </Txt>
          <Txt variant="bodySmall" tone="faint" style={{ fontSize: 12, marginLeft: 8 }}>
            {relative(reply.createdAt, now)}
          </Txt>
        </View>
        <Txt variant="bodySmall" style={{ marginTop: 2 }}>
          {reply.text}
        </Txt>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  content: { paddingHorizontal: 20, paddingTop: 4, paddingBottom: 28 },
  author: { flexDirection: 'row', alignItems: 'center' },
  review: { marginTop: 18, fontSize: 18, lineHeight: 28 },
  facts: { marginTop: 18, borderTopWidth: StyleSheet.hairlineWidth, borderBottomWidth: StyleSheet.hairlineWidth, paddingVertical: 6 },
  fact: { flexDirection: 'row', alignItems: 'flex-start', paddingVertical: 5 },
  with: { flexDirection: 'row', flexWrap: 'wrap' },
  ratingRow: { flexDirection: 'row', alignItems: 'center', paddingVertical: 8 },
  reply: { flexDirection: 'row', alignItems: 'flex-start', paddingVertical: 10 },
  replyHead: { flexDirection: 'row', alignItems: 'baseline' },
  composer: { borderTopWidth: StyleSheet.hairlineWidth, paddingHorizontal: 20, paddingTop: 10 },
  composerRow: { flexDirection: 'row', alignItems: 'flex-end' },
  input: { flex: 1, minHeight: 40, maxHeight: 110, borderWidth: StyleSheet.hairlineWidth, paddingHorizontal: 12, paddingTop: 9, paddingBottom: 9, fontFamily: fonts.serif, fontSize: 15 },
});
