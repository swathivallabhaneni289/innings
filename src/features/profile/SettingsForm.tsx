import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import * as ImagePicker from 'expo-image-picker';
import { useMemo, useState, type ReactNode } from 'react';
import { Pressable, StyleSheet, TextInput, View } from 'react-native';

import { Avatar, Button, Chip, Panel, SectionHeader, TeamBadge, Txt } from '@/components';
import { getTeam, useActions } from '@/data';
import type { User } from '@/domain/types';
import { fonts, useTheme } from '@/theme';
import { AVATAR_SWATCHES, BIO_MAX, FAV_TEAMS, NAME_MAX } from './constants';

type Patch = Partial<Omit<User, 'id' | 'handle'>>;

/** Edit the fan card: photo, name, one-line bio, favourite team and avatar colour. Saved with updateMe. */
export function SettingsForm({ me }: { me: User }) {
  const t = useTheme();
  const { updateMe } = useActions();
  const [name, setName] = useState(me.name);
  const [bio, setBio] = useState(me.bio ?? '');
  const [teamId, setTeamId] = useState(me.favouriteTeamId);
  const [colour, setColour] = useState(me.avatarColour);
  const [photoUri, setPhotoUri] = useState(me.photoUri ?? '');
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState<{ ok: boolean; text: string } | null>(null);

  const patch = useMemo(() => {
    const p: Patch = {};
    if (name.trim() !== me.name) p.name = name.trim();
    if (bio.trim() !== (me.bio ?? '')) p.bio = bio.trim();
    if (teamId !== me.favouriteTeamId) p.favouriteTeamId = teamId;
    if (colour.toLowerCase() !== me.avatarColour.toLowerCase()) p.avatarColour = colour;
    if (photoUri !== (me.photoUri ?? '')) p.photoUri = photoUri;
    return p;
  }, [name, bio, teamId, colour, photoUri, me]);
  const dirty = Object.keys(patch).length > 0;
  const nameError = name.trim().length === 0 ? 'Every fan needs a name on the scorecard.' : null;

  // A team picked before (or by the server default) that isn't on the fixed list still gets a chip.
  const listed = [...FAV_TEAMS.international, ...FAV_TEAMS.franchise] as string[];
  const international: string[] = listed.includes(me.favouriteTeamId)
    ? [...FAV_TEAMS.international]
    : [me.favouriteTeamId, ...FAV_TEAMS.international];

  const pickPhoto = async () => {
    setStatus(null);
    try {
      const res = await ImagePicker.launchImageLibraryAsync({ mediaTypes: ['images'], allowsEditing: true, aspect: [1, 1], quality: 0.7 });
      const uri = res.canceled ? undefined : res.assets?.[0]?.uri;
      if (uri) setPhotoUri(uri);
    } catch (e) {
      setStatus({ ok: false, text: e instanceof Error ? e.message : "Couldn't open your photos." });
    }
  };

  const save = async () => {
    if (!dirty || nameError) return;
    setSaving(true);
    setStatus(null);
    try {
      await updateMe(patch);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      setStatus({ ok: true, text: 'Saved. Your fan card is up to date.' });
    } catch (e) {
      setStatus({ ok: false, text: e instanceof Error ? e.message : "Couldn't save. Try again." });
    } finally {
      setSaving(false);
    }
  };

  const input = [styles.input, { color: t.ink, borderBottomColor: t.inkFaint }];
  const team = getTeam(teamId);

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

      <SectionHeader title="Favourite team" />
      <Panel>
        <Txt variant="bodySmall" tone="soft" style={styles.groupLabel}>
          International
        </Txt>
        <View style={styles.chips}>
          {international.map((id) => (
            <View key={id} style={styles.chipCell}>
              <Chip label={getTeam(id).name} selected={teamId === id} onPress={() => setTeamId(id)} />
            </View>
          ))}
        </View>
        <Txt variant="bodySmall" tone="soft" style={[styles.groupLabel, { marginTop: 6 }]}>
          Franchise
        </Txt>
        <View style={styles.chips}>
          {FAV_TEAMS.franchise.map((id) => (
            <View key={id} style={styles.chipCell}>
              <Chip label={getTeam(id).short} selected={teamId === id} onPress={() => setTeamId(id)} />
            </View>
          ))}
        </View>
        <View style={[styles.backing, { borderTopColor: t.rule }]}>
          <Txt variant="bodySmall" tone="soft">
            Backing
          </Txt>
          <View style={{ marginLeft: 10 }}>
            <TeamBadge teamId={teamId} size="sm" />
          </View>
          <Txt variant="bodySmall" numberOfLines={1} style={{ marginLeft: 8, flexShrink: 1, fontFamily: fonts.uiBold }}>
            {team.name}
          </Txt>
        </View>
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

      <Button label="Save changes" size="lg" onPress={save} loading={saving} disabled={!dirty || !!nameError} style={{ marginTop: 26 }} />
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
  groupLabel: { fontFamily: fonts.label, marginBottom: 8 },
  tap: { minHeight: 44 },
  chips: { flexDirection: 'row', flexWrap: 'wrap', marginBottom: 6 },
  chipCell: { marginBottom: 8 },
  backing: { flexDirection: 'row', alignItems: 'center', marginTop: 8, paddingTop: 12, borderTopWidth: StyleSheet.hairlineWidth },
  swatches: { flexDirection: 'row' },
  swatchRing: { width: 46, height: 46, borderRadius: 23, borderWidth: 2, alignItems: 'center', justifyContent: 'center', marginRight: 10 },
  swatch: { width: 34, height: 34, borderRadius: 17 },
});
