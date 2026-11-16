import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { useMemo, useState, type ReactNode } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Avatar, Button, Panel, SectionHeader, Txt } from '@/components';
import { ApiError, useActions } from '@/data';
import type { User } from '@/domain/types';
import { handleOk, HandleField, useHandleCheck } from '@/features/onboarding/HandleField';
import { pickSquarePhoto } from '@/features/onboarding/photo';
import { TeamChips } from '@/features/onboarding/TeamChips';
import { sameTeams } from '@/features/onboarding/teams';
import { fonts, useTheme } from '@/theme';
import { AVATAR_SWATCHES, BIO_MAX, NAME_MAX } from './constants';

type Patch = Partial<Omit<User, 'id'>>;

/** Edit the fan card: photo, name, @username, one-line bio, favourite team and avatar colour. Saved with updateMe. */
export function SettingsForm({ me }: { me: User }) {
  const t = useTheme();
  const { updateMe } = useActions();
  const [name, setName] = useState(me.name);
  const [handle, setHandle] = useState(me.handle);
  const [handleError, setHandleError] = useState<string | null>(null);
  const handleCheck = useHandleCheck(handle, me.handle || undefined);
  const [bio, setBio] = useState(me.bio ?? '');
  const [teamIds, setTeamIds] = useState(me.teamIds);
  const [colour, setColour] = useState(me.avatarColour);
  const [photoUri, setPhotoUri] = useState(me.photoUri ?? '');
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState<{ ok: boolean; text: string } | null>(null);

  const patch = useMemo(() => {
    const p: Patch = {};
    if (name.trim() !== me.name) p.name = name.trim();
    if (handle !== me.handle) p.handle = handle;
    if (bio.trim() !== (me.bio ?? '')) p.bio = bio.trim();
    if (teamIds.length > 0 && !sameTeams(teamIds, me.teamIds)) p.teamIds = teamIds;
    if (colour.toLowerCase() !== me.avatarColour.toLowerCase()) p.avatarColour = colour;
    if (photoUri !== (me.photoUri ?? '')) p.photoUri = photoUri;
    return p;
  }, [name, handle, bio, teamIds, colour, photoUri, me]);
  const dirty = Object.keys(patch).length > 0;
  const nameError = name.trim().length === 0 ? 'Every fan needs a name on the scorecard.' : null;
  const handleBlocked = !handleOk(handleCheck) || !!handleError;

  const pickPhoto = async () => {
    setStatus(null);
    try {
      const uri = await pickSquarePhoto();
      if (uri) setPhotoUri(uri);
    } catch (e) {
      setStatus({ ok: false, text: e instanceof Error ? e.message : "Couldn't open your photos." });
    }
  };

  const save = async () => {
    if (!dirty || nameError || handleBlocked) return;
    setSaving(true);
    setStatus(null);
    try {
      await updateMe(patch);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      setStatus({ ok: true, text: 'Saved. Your fan card is up to date.' });
    } catch (e) {
      if (e instanceof ApiError && (e.code === 'handle-taken' || e.code === 'bad-handle')) setHandleError(e.message);
      else setStatus({ ok: false, text: e instanceof Error ? e.message : "Couldn't save. Try again." });
    } finally {
      setSaving(false);
    }
  };

  const input = [styles.input, { color: t.ink, borderBottomColor: t.inkFaint }];

  return (
    <View>
      <SectionHeader title="Your fan card" style={{ marginTop: 8 }} />
      <Panel>
        <View style={styles.photoRow}>
          <Avatar user={{ name: name.trim() || me.name, avatarColour: colour, photoUri: photoUri || undefined }} size={72} />
          <View style={{ marginLeft: 16, alignItems: 'flex-start' }}>
            <Button
              label={photoUri ? 'Change photo' : 'Choose photo'}
              kind="outline"
              size="sm"
              icon={<Feather name="image" size={13} color={t.ink} />}
              onPress={pickPhoto}
              style={styles.tap}
            />
            {photoUri ? <Button label="Remove photo" kind="ghost" size="sm" onPress={() => setPhotoUri('')} style={styles.tap} /> : null}
          </View>
        </View>

        <Field label="Name" right={<Counter n={name.length} max={NAME_MAX} />}>
          <TextInput
            value={name}
            onChangeText={setName}
            maxLength={NAME_MAX}
            placeholder="Your name"
            placeholderTextColor={t.inkFaint}
            selectionColor={t.red}
            autoCapitalize="words"
            returnKeyType="done"
            style={input}
          />
          {nameError ? (
            <Txt variant="bodySmall" tone="red" style={{ marginTop: 4 }}>
              {nameError}
            </Txt>
          ) : null}
        </Field>

        <HandleField
          value={handle}
          onChangeText={(h) => {
            setHandle(h);
            setHandleError(null);
            setStatus(null);
          }}
          state={handleCheck}
          serverError={handleError}
        />

        <Field label="Fan bio, one line" right={<Counter n={bio.length} max={BIO_MAX} />}>
          <TextInput
            value={bio}
            onChangeText={(s) => setBio(s.replace(/\n/g, ' '))}
            maxLength={BIO_MAX}
            placeholder="Dhoni finishes, I cry"
            placeholderTextColor={t.inkFaint}
            selectionColor={t.red}
            returnKeyType="done"
            style={input}
          />
        </Field>
      </Panel>

      <SectionHeader title="Teams you back" />
      <Panel>
        <TeamChips value={teamIds} onChange={setTeamIds} saved={me.teamIds} />
        {teamIds.length === 0 ? (
          <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
            Back at least one team.
          </Txt>
        ) : null}
      </Panel>

      <SectionHeader title="Avatar colour" />
      <Panel>
        <View style={styles.swatches}>
          {AVATAR_SWATCHES.map((s) => {
            const on = s.colour.toLowerCase() === colour.toLowerCase();
            return (
              <Pressable
                key={s.colour}
                onPress={() => {
                  Haptics.selectionAsync();
                  setColour(s.colour);
                }}
                accessibilityRole="radio"
                accessibilityState={{ checked: on }}
                accessibilityLabel={s.name}
                style={[styles.swatchRing, { borderColor: on ? t.ink : 'transparent' }]}>
                <View style={[styles.swatch, { backgroundColor: s.colour }]} />
              </Pressable>
            );
          })}
        </View>
        <Txt variant="bodySmall" tone="soft" style={{ marginTop: 8 }}>
          {AVATAR_SWATCHES.find((s) => s.colour.toLowerCase() === colour.toLowerCase())?.name ?? 'Custom'}
          {photoUri ? ' · shows when you have no photo' : ''}
        </Txt>
      </Panel>

      <Button label="Save changes" size="lg" onPress={save} loading={saving} disabled={!dirty || !!nameError || handleBlocked} style={{ marginTop: 26 }} />
      {status && (!status.ok || !dirty) ? (
        <Txt variant="bodySmall" tone={status.ok ? 'green' : 'red'} align="center" style={{ marginTop: 8 }}>
          {status.ok ? '✓ ' : ''}
          {status.text}
        </Txt>
      ) : null}
    </View>
  );
}

function Field({ label, right, children }: { label: string; right?: ReactNode; children: ReactNode }) {
  return (
    <View style={{ marginTop: 20 }}>
      <View style={styles.fieldHead}>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label }}>
          {label}
        </Txt>
        {right}
      </View>
      {children}
    </View>
  );
}

function Counter({ n, max }: { n: number; max: number }) {
  return (
    <Txt variant="monoSmall" tone={n >= max ? 'red' : 'faint'}>
      {n}/{max}
    </Txt>
  );
}

const styles = StyleSheet.create({
  photoRow: { flexDirection: 'row', alignItems: 'center' },
  fieldHead: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  input: { fontFamily: fonts.serif, fontSize: 17, paddingVertical: 10, borderBottomWidth: StyleSheet.hairlineWidth },
  tap: { minHeight: 44 },
  swatches: { flexDirection: 'row' },
  swatchRing: { width: 46, height: 46, borderRadius: 23, borderWidth: 2, alignItems: 'center', justifyContent: 'center', marginRight: 10 },
  swatch: { width: 34, height: 34, borderRadius: 17 },
});
