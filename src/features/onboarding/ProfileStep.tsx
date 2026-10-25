import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { router } from 'expo-router';
import { useRef, useState } from 'react';
import { ScrollView, StyleSheet, TextInput, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Avatar, Button, KeyboardAware, Panel, SectionHeader, Txt } from '@/components';
import { ApiError, useActions } from '@/data';
import type { Account, User } from '@/domain/types';
import { errorText } from '@/features/auth/errors';
import { NAME_MAX } from '@/features/profile/constants';
import { fonts, useTheme } from '@/theme';
import { FanCardTicket } from './FanCardTicket';
import { handleOk, HandleField, useHandleCheck } from './HandleField';
import { pickSquarePhoto } from './photo';
import { TeamChips } from './TeamChips';

const yearOf = (iso: string | undefined) => {
  const y = iso ? new Date(iso).getFullYear() : NaN;
  return Number.isFinite(y) ? y : new Date().getFullYear();
};

/**
 * Step 1, "Your fan card": name, @username, favourite team and an optional photo, with the card printing
 * itself above as the fan types. Continue saves it all with PUT /v1/me.
 */
export function ProfileStep({ me, account }: { me: User; account: Account | undefined }) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const { updateMe } = useActions();
  const [name, setName] = useState(me.name);
  const [handle, setHandle] = useState(me.handle);
  const [teamId, setTeamId] = useState(me.favouriteTeamId);
  const handleInput = useRef<TextInput>(null);
  const [photoUri, setPhotoUri] = useState(me.photoUri ?? '');
  const [saving, setSaving] = useState(false);
  const [handleError, setHandleError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const check = useHandleCheck(handle, me.handle || undefined);

  const nameOk = name.trim().length > 0;
  const ready = nameOk && handleOk(check) && !handleError && !!teamId;
  const missing = [!nameOk && 'your name', !handleOk(check) && 'a username', !teamId && 'a team'].filter(Boolean) as string[];

  const pickPhoto = async () => {
    setError(null);
    try {
      const uri = await pickSquarePhoto();
      if (uri) setPhotoUri(uri);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't open your photos.");
    }
  };

  const save = async () => {
    if (!ready || saving) return;
    setSaving(true);
    setError(null);
    try {
      const photoChanged = photoUri !== (me.photoUri ?? '');
      await updateMe({ name: name.trim(), handle, favouriteTeamId: teamId, ...(photoChanged ? { photoUri } : {}) });
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      router.push('/onboarding/friends');
    } catch (e) {
      if (e instanceof ApiError && (e.code === 'handle-taken' || e.code === 'bad-handle')) setHandleError(e.message);
      else setError(errorText(e, "Couldn't save your card. Try again."));
    } finally {
      setSaving(false);
    }
  };

  return (
    <KeyboardAware>
      <ScrollView
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="on-drag"
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingHorizontal: 20, paddingBottom: insets.bottom + 32 }}>
        <Txt variant="body" tone="soft">
          How other fans will see you on stubs, replies and the Bragging Board. You can change it any time.
        </Txt>

        <FanCardTicket
          userId={me.id}
          name={name}
          handle={handle}
          teamId={teamId}
          avatarColour={me.avatarColour}
          photoUri={photoUri}
          memberSince={yearOf(account?.createdAt)}
          style={{ marginTop: 16 }}
        />

        <Panel style={{ marginTop: 20 }}>
          <View style={styles.photoRow}>
            <Avatar user={{ name: name.trim() || '·', avatarColour: me.avatarColour, photoUri: photoUri || undefined }} size={56} />
            <View style={{ marginLeft: 14, flex: 1, alignItems: 'flex-start' }}>
              <Button
                label={photoUri ? 'Change photo' : 'Add a photo'}
                kind="outline"
                size="sm"
                icon={<Feather name="image" size={13} color={t.ink} />}
                onPress={pickPhoto}
                style={styles.tap}
              />
              {photoUri ? (
                <Button label="Remove photo" kind="ghost" size="sm" onPress={() => setPhotoUri('')} style={styles.tap} />
              ) : (
                <Txt variant="monoSmall" tone="faint" style={{ marginTop: 4 }}>
                  Optional. Initials do the job too.
                </Txt>
              )}
            </View>
          </View>

          <View style={{ marginTop: 20 }}>
            <View style={styles.fieldHead}>
              <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label }}>
                Name
              </Txt>
              <Txt variant="monoSmall" tone={name.length >= NAME_MAX ? 'red' : 'faint'}>
                {name.length}/{NAME_MAX}
              </Txt>
            </View>
            <TextInput
              value={name}
              onChangeText={setName}
              maxLength={NAME_MAX}
              placeholder="As it goes on the scorecard"
              placeholderTextColor={t.inkFaint}
              selectionColor={t.red}
              autoCapitalize="words"
              autoComplete="name"
              textContentType="name"
              returnKeyType="next"
              submitBehavior="submit"
              onSubmitEditing={() => handleInput.current?.focus()}
              accessibilityLabel="Name"
              style={[styles.input, { color: t.ink, borderBottomColor: t.inkFaint }]}
            />
          </View>

          <HandleField
            inputRef={handleInput}
            value={handle}
            onChangeText={(h) => {
              setHandle(h);
              setHandleError(null);
            }}
            state={check}
            serverError={handleError}
          />
        </Panel>

        <SectionHeader title="Favourite team" />
        <Panel>
          <TeamChips value={teamId} onChange={setTeamId} saved={me.favouriteTeamId} />
        </Panel>

        {error ? (
          <Txt variant="bodySmall" tone="red" style={{ marginTop: 16 }}>
            {error}
          </Txt>
        ) : null}
        <Button label="Continue" size="lg" onPress={save} loading={saving} disabled={!ready} style={{ marginTop: 26 }} />
        {!ready && missing.length > 0 ? (
          <Txt variant="bodySmall" tone="faint" align="center" style={{ marginTop: 8 }}>
            Add {missing.length > 1 ? `${missing.slice(0, -1).join(', ')} and ${missing[missing.length - 1]}` : missing[0]} to carry on.
          </Txt>
        ) : null}
      </ScrollView>
    </KeyboardAware>
  );
}

const styles = StyleSheet.create({
  photoRow: { flexDirection: 'row', alignItems: 'center' },
  tap: { minHeight: 44 },
  fieldHead: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  input: { fontFamily: fonts.serif, fontSize: 17, paddingVertical: 10, borderBottomWidth: StyleSheet.hairlineWidth },
});
