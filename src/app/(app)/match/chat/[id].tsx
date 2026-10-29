import { router, useLocalSearchParams } from 'expo-router';
import { useEffect, useMemo, useRef, useState } from 'react';
import { FlatList, Keyboard, Platform, Pressable, StyleSheet, TextInput, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Avatar, Button, KeyboardAware, LiveTag, QueryState, Screen, Txt } from '@/components';
import { ApiError, getTeam, matchTitle, useActions, useMatchChat, useMatchPage, useMe, useNow } from '@/data';
import type { ChatResponse } from '@/domain/api';
import type { ChatMessage, Match, User } from '@/domain/types';
import { inningsLine, relative } from '@/lib/format';
import { fonts, useTheme } from '@/theme';

/** /match/chat/[id] — the match's chat room: fans talking through it while it's on. New messages arrive every 3 s. */
export default function MatchChatScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const page = useMatchPage(id);
  const q = useMatchChat(id);
  const match = page.data?.match;
  return (
    <KeyboardAware>
      <Screen back kicker="Match chat" title={match ? matchTitle(match) : 'Match chat'} scroll={false}>
        {match ? <ScoreLine match={match} /> : null}
        <QueryState q={q} loadingLabel="Finding a seat…">
          {(d) => <Room matchId={id} data={d} />}
        </QueryState>
      </Screen>
    </KeyboardAware>
  );
}

/** The score as it stands, so nobody has to leave the chat to check it. */
function ScoreLine({ match }: { match: Match }) {
  const t = useTheme();
  const innings = match.live?.innings ?? match.result?.innings ?? [];
  const now = innings[innings.length - 1];
  const status = match.live?.statusText ?? match.result?.text ?? (match.state === 'upcoming' ? `Starts ${relative(match.startTime)}` : '');
  return (
    <View style={[styles.score, { borderColor: t.rule }]}>
      <View style={styles.scoreTop}>
        {match.state === 'live' ? <LiveTag /> : null}
        {now ? (
          <Txt variant="mono" style={{ marginLeft: match.state === 'live' ? 10 : 0 }}>
            {getTeam(now.teamId).short} {inningsLine(now)}
          </Txt>
        ) : null}
      </View>
      {status ? (
        <Txt variant="bodySmall" tone="soft" numberOfLines={2} style={{ marginTop: 2 }}>
          {status}
        </Txt>
      ) : null}
    </View>
  );
}

function Room({ matchId, data }: { matchId: string; data: ChatResponse }) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const keyboard = useKeyboardShown();
  const me = useMe();
  const now = useNow(30_000);
  const { sendChat } = useActions();
  const inputRef = useRef<TextInput>(null);
  const [text, setText] = useState('');
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Newest first, for an inverted list that starts at the bottom and stays there as messages arrive.
  const rows = useMemo(() => [...data.messages].reverse(), [data.messages]);

  const send = async () => {
    const body = text.trim();
    if (!body || sending) return;
    setSending(true);
    setError(null);
    try {
      await sendChat(matchId, body);
      setText('');
    } catch (e) {
      if (e instanceof ApiError && e.status === 429) setError('Easy, one message a second.');
      else if (e instanceof ApiError && e.status === 409) setError('This chat is closed.');
      else setError('That didn’t go through. Try again.');
    } finally {
      setSending(false);
    }
  };

  const author = (m: ChatMessage): User | undefined => (m.userId === me.data?.id ? me.data : data.users[m.userId]);

  return (
    <View style={{ flex: 1 }}>
      <FlatList
        inverted
        data={rows}
        keyExtractor={(m) => m.id}
        contentContainerStyle={styles.list}
        keyboardShouldPersistTaps="handled"
        renderItem={({ item, index }) => (
          <Message
            message={item}
            user={author(item)}
            mine={item.userId === me.data?.id}
            // The list is inverted: the message above this one on screen is the next one in the array.
            sameAuthorAbove={rows[index + 1]?.userId === item.userId}
            now={now}
          />
        )}
        ListEmptyComponent={
          <View style={styles.empty}>
            <Txt variant="review" tone="soft" align="center">
              {data.open ? 'Nobody’s said anything yet. Start it off.' : data.opensAt ? `The chat opens ${relative(data.opensAt, now)}.` : 'This chat has closed.'}
            </Txt>
          </View>
        }
      />
      {/* The home-indicator gap only matters with the keyboard down; above the keyboard it would leave a gap. */}
      <View style={[styles.composer, { borderTopColor: t.rule, backgroundColor: t.paper, paddingBottom: keyboard ? 10 : Math.max(insets.bottom, 10) }]}>
        {error ? (
          <Txt variant="bodySmall" tone="red" style={{ marginBottom: 6 }}>
            {error}
          </Txt>
        ) : null}
        {data.open ? (
          <View style={styles.composerRow}>
            <TextInput
              ref={inputRef}
              value={text}
              onChangeText={setText}
              placeholder="Say something about the match…"
              placeholderTextColor={t.inkFaint}
              selectionColor={t.red}
              multiline
              maxLength={500}
              style={[styles.input, { color: t.ink, borderColor: t.rule, backgroundColor: t.card }]}
            />
            <Button
              label="Send"
              size="sm"
              kind={text.trim() || sending ? 'primary' : 'outline'}
              onPress={send}
              loading={sending}
              disabled={!text.trim()}
              style={{ marginLeft: 10 }}
            />
          </View>
        ) : (
          <Txt variant="bodySmall" tone="soft">
            {data.opensAt ? `You can post once the chat opens, a day before the first ball.` : 'This match’s chat has closed.'}
          </Txt>
        )}
      </View>
    </View>
  );
}

function useKeyboardShown() {
  const [shown, setShown] = useState(false);
  useEffect(() => {
    const ios = Platform.OS === 'ios';
    const show = Keyboard.addListener(ios ? 'keyboardWillShow' : 'keyboardDidShow', () => setShown(true));
    const hide = Keyboard.addListener(ios ? 'keyboardWillHide' : 'keyboardDidHide', () => setShown(false));
    return () => {
      show.remove();
      hide.remove();
    };
  }, []);
  return shown;
}

function Message({ message, user, mine, sameAuthorAbove, now }: { message: ChatMessage; user?: User; mine: boolean; sameAuthorAbove: boolean; now: number }) {
  const t = useTheme();
  return (
    <View style={[styles.message, { marginTop: sameAuthorAbove ? 2 : 12 }]}>
      <View style={{ width: 30 }}>
        {sameAuthorAbove ? null : (
          <Pressable onPress={() => router.push('/user/' + message.userId)} hitSlop={6} accessibilityLabel={user?.name ?? 'A fan'}>
            {user ? <Avatar user={user} size={30} /> : <View style={{ width: 30, height: 30, borderRadius: 15, backgroundColor: t.rule }} />}
          </Pressable>
        )}
      </View>
      <View style={{ flex: 1, marginLeft: 10 }}>
        {sameAuthorAbove ? null : (
          <View style={styles.head}>
            <Txt variant="bodySmall" numberOfLines={1} style={{ fontFamily: fonts.uiBold, flexShrink: 1 }} tone={mine ? 'red' : 'ink'}>
              {mine ? 'You' : (user?.name ?? 'A fan')}
            </Txt>
            <Txt variant="bodySmall" tone="faint" style={{ fontSize: 12, marginLeft: 8 }}>
              {relative(message.createdAt, now)}
            </Txt>
          </View>
        )}
        <Txt variant="body" style={{ marginTop: sameAuthorAbove ? 0 : 1 }}>
          {message.text}
        </Txt>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  score: { marginHorizontal: 20, paddingVertical: 10, borderTopWidth: StyleSheet.hairlineWidth, borderBottomWidth: StyleSheet.hairlineWidth },
  scoreTop: { flexDirection: 'row', alignItems: 'center' },
  list: { paddingHorizontal: 20, paddingVertical: 12, flexGrow: 1 },
  empty: { flex: 1, justifyContent: 'center', paddingVertical: 60, transform: [{ scaleY: -1 }] },
  message: { flexDirection: 'row', alignItems: 'flex-start' },
  head: { flexDirection: 'row', alignItems: 'baseline' },
  composer: { borderTopWidth: StyleSheet.hairlineWidth, paddingHorizontal: 20, paddingTop: 10 },
  composerRow: { flexDirection: 'row', alignItems: 'flex-end' },
  input: { flex: 1, minHeight: 40, maxHeight: 110, borderWidth: StyleSheet.hairlineWidth, paddingHorizontal: 12, paddingTop: 9, paddingBottom: 9, fontFamily: fonts.serif, fontSize: 15 },
});
