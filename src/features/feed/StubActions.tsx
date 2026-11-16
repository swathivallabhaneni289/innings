import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useRef, useState } from 'react';
import { Pressable, StyleSheet, View, type ViewStyle } from 'react-native';
import Svg, { Path } from 'react-native-svg';

import { Txt } from '@/components';
import { useActions } from '@/data';
import type { FeedItem } from '@/domain/api';
import { useTheme } from '@/theme';

const HEART = 'M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z';
// Feather's map pin with its dot cut out, so the filled pin keeps a hole on any background.
const PIN = 'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z M15 10a3 3 0 1 0-6 0a3 3 0 1 0 6 0z';

/** A Feather outline, inked in red when on. */
function Glyph({ d, filled, size = 18 }: { d: string; filled: boolean; size?: number }) {
  const t = useTheme();
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24">
      <Path d={d} fill={filled ? t.red : 'none'} fillRule="evenodd" stroke={filled ? t.red : t.inkSoft} strokeWidth={2} strokeLinejoin="round" />
    </Svg>
  );
}

const errText = (e: unknown) => (e instanceof Error ? e.message : 'That didn’t go through. Try again.');

/**
 * Like · reply · "I was there too" (a pin) for a logged stub, all three the same small size. Toggles are optimistic
 * and settle on the server's counts.
 */
export function StubActions({ item, onReply, style }: { item: FeedItem; onReply?: () => void; style?: ViewStyle }) {
  const t = useTheme();
  const { toggleLike, toggleThereToo } = useActions();
  const [like, setLike] = useState({ on: item.liked, n: item.likes });
  const [there, setThere] = useState({ on: item.iWasThere, n: item.thereToo });
  const [error, setError] = useState<string | null>(null);
  const busy = useRef({ like: false, there: false });

  // Settle on the server's counts whenever they change (a refetch, or another stub in this slot). Likes and
  // "there too" are followed apart, so a new like count never cuts across a there-too tap still in flight.
  const likeFrom = `${item.log.id}|${item.liked}|${item.likes}`;
  const [likeSeen, setLikeSeen] = useState(likeFrom);
  if (likeSeen !== likeFrom) {
    setLikeSeen(likeFrom);
    setLike({ on: item.liked, n: item.likes });
  }
  const thereFrom = `${item.log.id}|${item.iWasThere}|${item.thereToo}`;
  const [thereSeen, setThereSeen] = useState(thereFrom);
  if (thereSeen !== thereFrom) {
    setThereSeen(thereFrom);
    setThere({ on: item.iWasThere, n: item.thereToo });
  }

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
          <Glyph d={HEART} filled={like.on} />
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
        <Pressable
          onPress={onThere}
          hitSlop={8}
          style={styles.action}
          accessibilityRole="button"
          accessibilityLabel={`I was there too, ${there.n}`}
          accessibilityState={{ selected: there.on }}>
          <Glyph d={PIN} filled={there.on} />
          <Txt variant="mono" style={{ marginLeft: 6, color: there.on ? t.red : t.inkSoft }}>
            {there.n}
          </Txt>
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
});
