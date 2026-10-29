import { Feather } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { router, useLocalSearchParams } from 'expo-router';
import { useMemo, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Button, Paper, QueryState, Txt } from '@/components';
import { ApiError, getTeam, matchTitle, squads, useActions, useMatchPage } from '@/data';
import type { MatchPageResponse } from '@/domain/api';
import type { CallSpec, Match } from '@/domain/types';
import { fonts, useTheme } from '@/theme';

/** /match/call/[id] — a fan makes their own call on a match, which goes into its chat for everyone to answer. */
export default function MakeCallScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const q = useMatchPage(id);
  const insets = useSafeAreaInsets();
  const t = useTheme();
  return (
    <Paper>
      <View style={[styles.top, { paddingTop: Math.max(insets.top, 16) }]}>
        <View style={{ flex: 1 }}>
          <Txt variant="label" tone="red">
            Make a call
          </Txt>
          <Txt variant="h1">{q.data ? matchTitle(q.data.match) : 'Your call'}</Txt>
        </View>
        <Pressable onPress={() => router.back()} hitSlop={12} accessibilityRole="button" accessibilityLabel="Close">
          <Feather name="x" size={24} color={t.ink} />
        </Pressable>
      </View>
      <QueryState q={q}>{(d) => <Builder data={d} />}</QueryState>
    </Paper>
  );
}

type Kind = CallSpec['kind'];
const KINDS: { kind: Kind; title: string; line: string }[] = [
  { kind: 'runs', title: 'Batter', line: 'Will he make 50?' },
  { kind: 'teamTotal', title: 'Total', line: 'Will they get to 180?' },
  { kind: 'win', title: 'Result', line: 'Who takes it?' },
];
const RUN_TARGETS = [30, 50, 100];

function Builder({ data }: { data: MatchPageResponse }) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const { makeCall } = useActions();
  const m = data.match;
  const [kind, setKind] = useState<Kind>('runs');
  const [team, setTeam] = useState(m.teamIds[0]);
  const [player, setPlayer] = useState<string | null>(null);
  const [target, setTarget] = useState(50);
  const [total, setTotal] = useState(defaultTotal(m));
  const [side, setSide] = useState<'yes' | 'no'>('yes');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const groups = useMemo(() => squads(m), [m]);
  const players = groups.find((g) => g.team.id === team)?.players ?? [];
  const playerName = players.find((p) => p.id === player)?.name;

  const question =
    kind === 'runs'
      ? `Will ${playerName ?? '…'} make ${target}?`
      : kind === 'teamTotal'
        ? `Will ${getTeam(team).name} score ${total} or more?`
        : `Will ${getTeam(team).name} win?`;
  const ready = kind !== 'runs' || !!player;
  const limited = m.format !== 'Test';

  const post = async () => {
    if (!ready || busy) return;
    setBusy(true);
    setError(null);
    const spec: CallSpec =
      kind === 'runs' ? { kind, playerId: player!, target, side } : kind === 'teamTotal' ? { kind, teamId: team, target: total, side } : { kind, teamId: team, side };
    try {
      await makeCall(m.id, spec);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => {});
      router.back();
    } catch (e) {
      const code = e instanceof ApiError ? e.code : '';
      setError(
        code === 'call-too-late'
          ? 'Too late for that one: it’s as good as decided. Try a bigger target or another player.'
          : code === 'too-many-calls'
            ? 'That’s five calls on this match already.'
            : code === 'calls-off'
              ? 'This match doesn’t take calls.'
              : 'That call couldn’t be made. Try again.',
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <View style={{ flex: 1 }}>
      <ScrollView contentContainerStyle={styles.body} keyboardShouldPersistTaps="handled">
        <Step n={1} title="What are you calling?" />
        <View style={styles.kinds}>
          {KINDS.filter((k) => limited || k.kind !== 'teamTotal').map((k) => {
            const on = kind === k.kind;
            return (
              <Pressable
                key={k.kind}
                onPress={() => setKind(k.kind)}
                accessibilityRole="button"
                accessibilityState={{ selected: on }}
                style={({ pressed }) => [styles.kind, { borderColor: on ? t.ink : t.rule, backgroundColor: on ? t.ink : t.card, transform: [{ scale: pressed ? 0.97 : 1 }] }]}>
                <Txt style={{ fontFamily: fonts.display, fontSize: 22, lineHeight: 24, color: on ? t.onInk : t.ink }} allowFontScaling={false}>
                  {k.title.toUpperCase()}
                </Txt>
                <Txt variant="bodySmall" style={{ color: on ? t.onInk : t.inkSoft, marginTop: 4, fontSize: 12 }} numberOfLines={2}>
                  {k.line}
                </Txt>
              </Pressable>
            );
          })}
        </View>

        <Step n={2} title={kind === 'runs' ? 'Which batter?' : 'Which team?'} />
        <View style={styles.teams}>
          {m.teamIds.map((id) => {
            const tm = getTeam(id);
            const on = team === id;
            return (
              <Pressable
                key={id}
                onPress={() => {
                  setTeam(id);
                  setPlayer(null);
                }}
                accessibilityRole="button"
                accessibilityState={{ selected: on }}
                accessibilityLabel={tm.name}
                style={({ pressed }) => [styles.team, { borderColor: on ? t.ink : t.rule, backgroundColor: on ? t.card : 'transparent', opacity: pressed ? 0.85 : 1 }]}>
                <Txt style={{ fontFamily: fonts.display, fontSize: 30, lineHeight: 32, color: t.ink, opacity: on ? 1 : 0.45 }} allowFontScaling={false}>
                  {tm.short}
                </Txt>
                <View style={[styles.kit, { backgroundColor: tm.colour, opacity: on ? 1 : 0.35 }]} />
              </Pressable>
            );
          })}
        </View>
        {kind === 'runs' ? (
          players.length ? (
            <View style={styles.players}>
              {players.map((p) => {
                const on = player === p.id;
                return (
                  <Pressable
                    key={p.id}
                    onPress={() => setPlayer(p.id)}
                    accessibilityRole="button"
                    accessibilityState={{ selected: on }}
                    style={({ pressed }) => [styles.player, { borderColor: on ? t.ink : t.rule, backgroundColor: on ? t.ink : t.card, opacity: pressed ? 0.85 : 1 }]}>
                    <Txt style={{ fontFamily: fonts.label, fontSize: 14, color: on ? t.onInk : t.ink }} numberOfLines={1}>
                      {p.name}
                    </Txt>
                    <Txt variant="label" style={{ fontSize: 8.5, color: on ? t.onInk : t.inkFaint, marginLeft: 6 }}>
                      {p.role}
                    </Txt>
                  </Pressable>
                );
              })}
            </View>
          ) : (
            <Txt variant="bodySmall" tone="soft" style={{ marginTop: 10 }}>
              The squads aren’t out yet for this one.
            </Txt>
          )
        ) : null}

        {kind !== 'win' ? <Step n={3} title={kind === 'runs' ? 'How many?' : 'What total?'} /> : null}
        {kind === 'runs' ? (
          <View style={styles.targets}>
            {RUN_TARGETS.map((n) => {
              const on = target === n;
              return (
                <Pressable
                  key={n}
                  onPress={() => setTarget(n)}
                  accessibilityRole="button"
                  accessibilityState={{ selected: on }}
                  style={({ pressed }) => [styles.target, { borderColor: on ? t.ink : t.rule, backgroundColor: on ? t.ink : t.card, transform: [{ scale: pressed ? 0.97 : 1 }] }]}>
                  <Txt style={{ fontFamily: fonts.display, fontSize: 40, lineHeight: 42, color: on ? t.onInk : t.ink }} allowFontScaling={false}>
                    {n}
                  </Txt>
                  <Txt variant="label" style={{ fontSize: 9, color: on ? t.onInk : t.inkSoft }}>
                    {n === 100 ? 'A hundred' : n === 50 ? 'Fifty' : 'Thirty'}
                  </Txt>
                </Pressable>
              );
            })}
          </View>
        ) : kind === 'teamTotal' ? (
          <View style={[styles.stepper, { borderColor: t.ink }]}>
            <StepButton icon="minus" onPress={() => setTotal((v) => Math.max(50, v - 5))} label="Five fewer" />
            <View style={{ flex: 1, alignItems: 'center' }}>
              <Txt style={{ fontFamily: fonts.display, fontSize: 56, lineHeight: 58, color: t.ink }} allowFontScaling={false}>
                {total}
              </Txt>
              <Txt variant="label" tone="soft" style={{ fontSize: 9 }}>
                or more
              </Txt>
            </View>
            <StepButton icon="plus" onPress={() => setTotal((v) => Math.min(500, v + 5))} label="Five more" />
          </View>
        ) : null}

        <Step n={kind === 'win' ? 3 : 4} title="Your side" />
        <View style={[styles.sides, { borderColor: t.ink }]}>
          {(['yes', 'no'] as const).map((s, k) => {
            const on = side === s;
            return (
              <Pressable
                key={s}
                onPress={() => setSide(s)}
                accessibilityRole="button"
                accessibilityState={{ selected: on }}
                style={[styles.side, { backgroundColor: on ? t.ink : 'transparent', borderLeftWidth: k ? 1.5 : 0, borderLeftColor: t.ink }]}>
                <Txt style={{ fontFamily: fonts.display, fontSize: 28, lineHeight: 30, color: on ? t.onInk : t.ink }} allowFontScaling={false}>
                  {s.toUpperCase()}
                </Txt>
              </Pressable>
            );
          })}
        </View>

        <View style={[styles.preview, { borderColor: t.rule, backgroundColor: t.card }]}>
          <Txt variant="label" tone="soft">
            Goes into the chat as
          </Txt>
          <Txt style={{ fontFamily: fonts.heading, fontSize: 20, lineHeight: 25, color: t.ink, marginTop: 6 }}>{question}</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 6 }}>
            Everyone can answer until it locks. It settles itself from the live score, 5 points if you’re right.
          </Txt>
        </View>
        {error ? (
          <Txt variant="bodySmall" tone="red" style={{ marginTop: 12 }}>
            {error}
          </Txt>
        ) : null}
      </ScrollView>
      <View style={[styles.bar, { borderTopColor: t.rule, paddingBottom: Math.max(insets.bottom, 14) }]}>
        <Button label="Post the call" size="lg" onPress={post} loading={busy} disabled={!ready} style={{ minHeight: 54 }} />
      </View>
    </View>
  );
}

function defaultTotal(m: Match): number {
  const women = m.gender === 'women';
  if (m.format === 'ODI') return women ? 250 : 280;
  if (m.format === 'The Hundred') return women ? 130 : 150;
  return women ? 140 : 160;
}

function Step({ n, title }: { n: number; title: string }) {
  const t = useTheme();
  return (
    <View style={styles.step}>
      <View style={[styles.stepNo, { borderColor: t.red }]}>
        <Txt style={{ fontFamily: fonts.monoMedium, fontSize: 11, color: t.red }}>{n}</Txt>
      </View>
      <Txt variant="h3">{title}</Txt>
    </View>
  );
}

function StepButton({ icon, onPress, label }: { icon: 'minus' | 'plus'; onPress: () => void; label: string }) {
  const t = useTheme();
  return (
    <Pressable
      onPress={() => {
        onPress();
        Haptics.selectionAsync().catch(() => {});
      }}
      accessibilityRole="button"
      accessibilityLabel={label}
      hitSlop={8}
      style={({ pressed }) => [styles.stepBtn, { borderColor: t.ink, backgroundColor: pressed ? t.ink : 'transparent' }]}>
      {({ pressed }) => <Feather name={icon} size={22} color={pressed ? t.onInk : t.ink} />}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  top: { flexDirection: 'row', alignItems: 'flex-start', paddingHorizontal: 20, paddingBottom: 6 },
  body: { paddingHorizontal: 20, paddingBottom: 30 },
  step: { flexDirection: 'row', alignItems: 'center', marginTop: 24, marginBottom: 10 },
  stepNo: { width: 22, height: 22, borderRadius: 11, borderWidth: 1.5, alignItems: 'center', justifyContent: 'center', marginRight: 10 },
  kinds: { flexDirection: 'row', gap: 8 },
  kind: { flex: 1, borderWidth: 1.5, paddingHorizontal: 10, paddingVertical: 12, minHeight: 88 },
  teams: { flexDirection: 'row', gap: 8 },
  team: { flex: 1, borderWidth: 1.5, alignItems: 'center', paddingTop: 10, paddingBottom: 0, overflow: 'hidden' },
  kit: { alignSelf: 'stretch', height: 5, marginTop: 8 },
  players: { flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginTop: 12 },
  player: { flexDirection: 'row', alignItems: 'center', borderWidth: 1, paddingHorizontal: 12, minHeight: 40 },
  targets: { flexDirection: 'row', gap: 8 },
  target: { flex: 1, borderWidth: 1.5, alignItems: 'center', paddingVertical: 10 },
  stepper: { flexDirection: 'row', alignItems: 'center', borderWidth: 1.5, paddingHorizontal: 12, paddingVertical: 10 },
  stepBtn: { width: 48, height: 48, borderWidth: 1.5, alignItems: 'center', justifyContent: 'center' },
  sides: { flexDirection: 'row', borderWidth: 1.5, height: 64 },
  side: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  preview: { borderWidth: StyleSheet.hairlineWidth, padding: 16, marginTop: 24 },
  bar: { borderTopWidth: StyleSheet.hairlineWidth, paddingHorizontal: 20, paddingTop: 12 },
});
