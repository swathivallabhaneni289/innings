import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useEffect, useRef, useState } from 'react';
import { Pressable, StyleSheet, View, type ViewStyle } from 'react-native';
import Svg, { Path } from 'react-native-svg';

import { Txt } from '@/components';
import { useActions } from '@/data';
import type { FeedItem } from '@/domain/api';
import { fonts, useTheme } from '@/theme';

const HEART = 'M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z';

/** Feather's heart outline, inked in red when liked. */
function Heart({ filled, size = 18 }: { filled: boolean; size?: number }) {
  const t = useTheme();
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24">
      <Path d={HEART} fill={filled ? t.red : 'none'} stroke={filled ? t.red : t.inkSoft} strokeWidth={2} strokeLinejoin="round" />
    </Svg>
  );
}

const errText = (e: unknown) => (e instanceof Error ? e.message : 'That didn’t go through. Try again.');

/** Like · reply · "I was there too" for a logged stub. Toggles are optimistic and settle on the server's counts. */
export function StubActions({ item, onReply, style }: { item: FeedItem; onReply?: () => void; style?: ViewStyle }) {
  const t = useTheme();
  const { toggleLike, toggleThereToo } = useActions();
  const [like, setLike] = useState({ on: item.liked, n: item.likes });
  const [there, setThere] = useState({ on: item.iWasThere, n: item.thereToo });
  const [error, setError] = useState<string | null>(null);
  const busy = useRef({ like: false, there: false });

  useEffect(() => setLike({ on: item.liked, n: item.likes }), [item.log.id, item.liked, item.likes]);
  useEffect(() => setThere({ on: item.iWasThere, n: item.thereToo }), [item.log.id, item.iWasThere, item.thereToo]);

  const onLike = async () => {
    if (busy.current.like) return;
    busy.current.like = true;
    Haptics.selectionAsync().catch(() => {});
    const prev = like;
    setLike({ on: !prev.on, n: Math.max(0, prev.n + (prev.on ? -1 : 1)) });
    setError(null);
    try {
      const r = await toggleLike(item.log.id);
      setLike({ on: r.liked, n: r.likes });
    } catch (e) {
      setLike(prev);
      setError(errText(e));
    } finally {
      busy.current.like = false;
    }
  };

  const onThere = async () => {
    if (busy.current.there) return;
    busy.current.there = true;
    Haptics.selectionAsync().catch(() => {});
    const prev = there;
    setThere({ on: !prev.on, n: Math.max(0, prev.n + (prev.on ? -1 : 1)) });
    setError(null);
    try {
      const r = await toggleThereToo(item.log.id);
      setThere({ on: r.iWasThere, n: r.thereToo });
    } catch (e) {
      setThere(prev);
      setError(errText(e));
    } finally {
      busy.current.there = false;
    }
  };

  return (
    <View style={style}>
      <View style={styles.row}>
        <Pressable
          onPress={onLike}
          hitSlop={8}
          style={styles.action}
          accessibilityRole="button"
          accessibilityLabel={like.on ? 'Unlike' : 'Like'}
          accessibilityState={{ selected: like.on }}>
          <Heart filled={like.on} />
          <Txt variant="mono" style={{ marginLeft: 6, color: like.on ? t.red : t.inkSoft }}>
            {like.n}
          </Txt>
        </Pressable>
        <Pressable onPress={onReply} hitSlop={8} style={styles.action} accessibilityRole="button" accessibilityLabel="Replies">
          <Feather name="message-square" size={17} color={t.inkSoft} />
          <Txt variant="mono" tone="soft" style={{ marginLeft: 6 }}>
            {item.replies.length}
          </Txt>
        </Pressable>
        <View style={{ flex: 1 }} />
        <Pressable
          onPress={onThere}
          hitSlop={6}
          accessibilityRole="button"
          accessibilityState={{ selected: there.on }}
          accessibilityLabel={`I was there too, ${there.n}`}
          style={({ pressed }) => [
            styles.there,
            { borderColor: there.on ? t.ink : t.cardEdge, backgroundColor: there.on ? t.ink : 'transparent', opacity: pressed ? 0.8 : 1 },
          ]}>
          <Feather name="map-pin" size={13} color={there.on ? t.onInk : t.ink} />
          <Txt style={{ fontFamily: fonts.label, fontSize: 13, lineHeight: 17, marginLeft: 6, color: there.on ? t.onInk : t.ink }}>I was there too</Txt>
          {there.n > 0 ? (
            <Txt variant="monoSmall" style={{ marginLeft: 7, color: there.on ? t.onInk : t.inkSoft }}>
              {there.n}
            </Txt>
          ) : null}
        </Pressable>
      </View>
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 6 }}>
          {error}
        </Txt>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center' },
  action: { flexDirection: 'row', alignItems: 'center', marginRight: 20, minHeight: 44 },
  there: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, minHeight: 36, paddingHorizontal: 12 },
});
