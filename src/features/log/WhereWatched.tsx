import { Feather } from '@expo/vector-icons';
import { Image } from 'expo-image';
import * as Haptics from 'expo-haptics';
import * as ImagePicker from 'expo-image-picker';
import { useState } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Avatar, Button, Chip, Empty, QueryState, SectionHeader, Txt } from '@/components';
import { useFriends } from '@/data';
import { WATCH_PLACES } from '@/domain/constants';
import type { User, WatchPlace } from '@/domain/types';
import { fonts, useTheme } from '@/theme';

export interface WhereValue {
  place: WatchPlace;
  placeNote: string;
  companions: number;
  watchedWith: string[];
  photoUri?: string;
  review: string;
}

function FriendToggle({ user, on, onPress }: { user: User; on: boolean; onPress: () => void }) {
  const t = useTheme();
  return (
    <Pressable
      onPress={onPress}
      accessibilityRole="checkbox"
      accessibilityState={{ checked: on }}
      style={[styles.friend, { borderColor: on ? t.ink : t.rule, backgroundColor: on ? t.ink : t.card }]}>
      <Avatar user={user} size={26} />
      <Txt numberOfLines={1} style={{ marginLeft: 8, fontFamily: fonts.label, fontSize: 13, color: on ? t.onInk : t.ink, flexShrink: 1 }}>
        {user.name.split(' ')[0]}
      </Txt>
      {on ? <Feather name="check" size={14} color={t.onInk} style={{ marginLeft: 6 }} /> : null}
    </Pressable>
  );
}

/** Last step: the place (7 spots), where exactly, how many people, tagged friends, a photo and a review. */
export function WhereWatched({
  value,
  defaulted,
  onChange,
  onReviewFocus,
  error,
}: {
  value: WhereValue;
  defaulted: 'last' | 'home' | null; // why a place is pre-selected (null = the fan picked it)
  onChange: (patch: Partial<WhereValue>) => void;
  onReviewFocus?: () => void; // lets the flow scroll the review above the keyboard
  error?: string | null;
}) {
  const t = useTheme();
  const friends = useFriends();
  const [photoError, setPhotoError] = useState<string | null>(null);

  const setCompanions = (n: number) => {
    Haptics.selectionAsync();
    onChange({ companions: Math.max(0, Math.min(999, n)) });
  };

  const toggleFriend = (id: string) => {
    Haptics.selectionAsync();
    const on = value.watchedWith.includes(id);
    const watchedWith = on ? value.watchedWith.filter((x) => x !== id) : [...value.watchedWith, id];
    // Tagged friends were with you, so the head count is at least that many.
    onChange({ watchedWith, companions: Math.max(value.companions, watchedWith.length) });
  };

  const pickPhoto = async () => {
    setPhotoError(null);
    try {
      const res = await ImagePicker.launchImageLibraryAsync({ mediaTypes: ['images'], allowsEditing: true, aspect: [4, 3], quality: 0.7 });
      if (!res.canceled && res.assets[0]) onChange({ photoUri: res.assets[0].uri });
    } catch (e) {
      setPhotoError(e instanceof Error ? e.message : "Couldn't open your photos.");
    }
  };

  return (
    <View>
      <SectionHeader title="Where were you?" style={{ marginTop: 20 }} />
      <View style={styles.wrap}>
        {WATCH_PLACES.map((p) => (
          <View key={p.id} style={{ marginBottom: 8 }}>
            <Chip
              label={`${p.emoji}  ${p.label}`}
              selected={value.place === p.id}
              onPress={() => {
                Haptics.selectionAsync();
                onChange({ place: p.id });
              }}
              style={styles.placeChip}
            />
          </View>
        ))}
      </View>
      {defaulted ? (
        <Txt variant="bodySmall" tone="faint">
          {defaulted === 'last' ? 'Same spot as your last log. Tap to change.' : 'Home by default. Tap to change.'}
        </Txt>
      ) : null}
      {error ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 4 }}>
          {error}
        </Txt>
      ) : null}

      <TextInput
        value={value.placeNote}
        onChangeText={(placeNote) => onChange({ placeNote })}
        placeholder="Where exactly? “Hostel room”, “Uncle's place”"
        placeholderTextColor={t.inkFaint}
        maxLength={200}
        returnKeyType="done"
        style={[styles.field, { color: t.ink, backgroundColor: t.card }]}
      />

      <View style={[styles.stepperRow, { backgroundColor: t.card }]}>
        <View style={{ flex: 1 }}>
          <Txt variant="h3">People with you</Txt>
          <Txt variant="bodySmall" tone="soft">
            {value.companions === 0 ? 'Just me' : `With ${value.companions} ${value.companions === 1 ? 'person' : 'people'}`}
          </Txt>
        </View>
        <View style={[styles.stepper, { borderColor: t.rule }]}>
          <Pressable
            onPress={() => setCompanions(value.companions - 1)}
            disabled={value.companions === 0}
            accessibilityRole="button"
            accessibilityLabel="One fewer person"
            style={styles.stepBtn}>
            <Feather name="minus" size={16} color={value.companions === 0 ? t.inkFaint : t.ink} />
          </Pressable>
          <View style={[styles.stepValue, { borderColor: t.rule }]}>
            <Txt variant="stat" style={{ fontSize: 22, lineHeight: 26 }}>
              {value.companions}
            </Txt>
          </View>
          <Pressable
            onPress={() => setCompanions(value.companions + 1)}
            accessibilityRole="button"
            accessibilityLabel="One more person"
            style={styles.stepBtn}>
            <Feather name="plus" size={16} color={t.ink} />
          </Pressable>
        </View>
      </View>

      <SectionHeader title="Who was with you?" />
      <QueryState q={friends} loadingLabel="Calling the dressing room…">
        {(data) => {
          const following = data.friends.filter((f) => f.following).map((f) => f.user);
          if (following.length === 0) return <Empty title="Batting alone" body="Follow some fans and you can tag them here." />;
          return (
            <View style={styles.wrap}>
              {following.map((u) => (
                <FriendToggle key={u.id} user={u} on={value.watchedWith.includes(u.id)} onPress={() => toggleFriend(u.id)} />
              ))}
            </View>
          );
        }}
      </QueryState>

      <SectionHeader title="Photo & review" />
      <View style={styles.photoRow}>
        {value.photoUri ? (
          <>
            <Image source={{ uri: value.photoUri }} style={styles.thumb} contentFit="cover" />
            <View style={{ marginLeft: 14, alignItems: 'flex-start' }}>
              <Button label="Change" kind="outline" onPress={pickPhoto} />
              <Pressable onPress={() => onChange({ photoUri: undefined })} accessibilityRole="button" style={styles.remove}>
                <Txt tone="red" style={{ fontFamily: fonts.uiBold, fontSize: 13 }}>
                  Remove
                </Txt>
              </Pressable>
            </View>
          </>
        ) : (
          <Button label="Add a photo" kind="outline" icon={<Feather name="camera" size={14} color={t.ink} />} onPress={pickPhoto} />
        )}
      </View>
      {photoError ? (
        <Txt variant="bodySmall" tone="red" style={{ marginTop: 6 }}>
          {photoError}
        </Txt>
      ) : null}

      <TextInput
        value={value.review}
        onChangeText={(review) => onChange({ review })}
        onFocus={onReviewFocus}
        placeholder="Your review (optional). What will you remember?"
        placeholderTextColor={t.inkFaint}
        multiline
        maxLength={2000}
        textAlignVertical="top"
        style={[styles.review, { color: t.ink, backgroundColor: t.card }]}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { flexDirection: 'row', flexWrap: 'wrap' },
  placeChip: { minHeight: 44, paddingVertical: 10 },
  field: { fontFamily: fonts.serif, fontSize: 15, height: 48, paddingHorizontal: 14, paddingVertical: 0, marginTop: 12 },
  stepperRow: { flexDirection: 'row', alignItems: 'center', padding: 14, marginTop: 12 },
  stepper: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, height: 44 },
  stepBtn: { width: 44, height: '100%', alignItems: 'center', justifyContent: 'center' },
  stepValue: { minWidth: 48, height: '100%', alignItems: 'center', justifyContent: 'center', borderLeftWidth: 1, borderRightWidth: 1, paddingHorizontal: 6 },
  friend: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, minHeight: 44, paddingLeft: 8, paddingRight: 12, marginRight: 8, marginBottom: 8, maxWidth: 170 },
  photoRow: { flexDirection: 'row', alignItems: 'center' },
  thumb: { width: 104, height: 78 },
  remove: { minHeight: 44, justifyContent: 'center' },
  review: { fontFamily: fonts.serif, fontSize: 16, lineHeight: 23, minHeight: 120, padding: 14, marginTop: 12 },
});
