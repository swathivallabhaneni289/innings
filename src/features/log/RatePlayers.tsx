import { Feather } from '@expo/vector-icons';
import Slider from '@react-native-community/slider';
import * as Haptics from 'expo-haptics';
import { memo, useRef, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, useWindowDimensions, View } from 'react-native';

import { Panel, Txt } from '@/components';
import { getPlayer, squads } from '@/data';
import type { Match, Player, PlayerRole } from '@/domain/types';
import { fonts, useTheme } from '@/theme';

const GUTTER = 20; // the flow's horizontal padding: the pager bleeds to the screen edges
const ROLE: Record<PlayerRole, string> = { BAT: 'Batter', BOWL: 'Bowler', AR: 'All-rounder', WK: 'Wicketkeeper' };

/** One player: crown them, slide 1–10 (0 = skipped), or skip. */
const PlayerRow = memo(function PlayerRow({
  player,
  value,
  potm,
  last,
  onRate,
  onPotm,
}: {
  player: Player;
  value: number;
  potm: boolean;
  last: boolean;
  onRate: (playerId: string, v: number) => void;
  onPotm: (playerId: string) => void;
}) {
  const t = useTheme();
  const rated = value >= 1;
  return (
    <View style={[styles.player, last ? null : { borderBottomColor: t.rule, borderBottomWidth: StyleSheet.hairlineWidth }]}>
      <View style={styles.playerTop}>
        <Pressable
          onPress={() => {
            Haptics.selectionAsync();
            onPotm(player.id);
          }}
          hitSlop={4}
          accessibilityRole="button"
          accessibilityState={{ selected: potm }}
          accessibilityLabel={potm ? `${player.name} is your Player of the Match` : `Make ${player.name} your Player of the Match`}
          style={[styles.crown, { borderColor: potm ? t.ink : t.rule, backgroundColor: potm ? t.ink : 'transparent' }]}>
          <Feather name="award" size={17} color={potm ? t.onInk : t.inkFaint} />
        </Pressable>
        <View style={{ flex: 1, marginLeft: 12 }}>
          <Txt variant="h3" numberOfLines={1}>
            {player.name}
          </Txt>
          <Txt variant="bodySmall" tone={potm ? 'ink' : 'faint'} numberOfLines={1} style={{ fontSize: 12, lineHeight: 16 }}>
            {ROLE[player.role] ?? player.role}
            {potm ? <Txt style={{ fontFamily: fonts.uiBold, fontSize: 12, lineHeight: 16 }}> · your player of the match</Txt> : null}
          </Txt>
        </View>
        <Txt variant="stat" tone={rated ? 'ink' : 'faint'} style={styles.value}>
          {rated ? String(value) : '–'}
        </Txt>
        <Pressable
          onPress={() => onRate(player.id, 0)}
          disabled={!rated}
          accessibilityRole="button"
          accessibilityLabel={`Skip ${player.name}`}
          style={styles.skip}>
          <Txt tone={rated ? 'soft' : 'faint'} style={{ fontFamily: rated ? fonts.uiBold : fonts.ui, fontSize: 13 }}>
            {rated ? 'Skip' : 'Skipped'}
          </Txt>
        </Pressable>
      </View>
      <Slider
        style={styles.slider}
        minimumValue={0}
        maximumValue={10}
        step={1}
        value={value}
        onValueChange={(v) => onRate(player.id, Math.round(v))}
        minimumTrackTintColor={rated ? t.ink : t.inkFaint}
        maximumTrackTintColor={t.rule}
        thumbTintColor={rated ? t.ink : t.inkFaint}
        accessibilityLabel={`Rate ${player.name} out of 10`}
      />
    </View>
  );
});

/**
 * Players step: both XIs as a two-page horizontal pager (one page per team), a 1–10 slider per player, and an optional
 * fan Player of the Match.
 */
export function RatePlayers({
  match,
  ratings,
  potmId,
  onRate,
  onPotm,
}: {
  match: Match;
  ratings: Record<string, number>;
  potmId: string | null;
  onRate: (playerId: string, v: number) => void;
  onPotm: (playerId: string) => void;
}) {
  const t = useTheme();
  const { width } = useWindowDimensions();
  const teams = squads(match).filter((s) => s.players.length > 0);
  const pager = useRef<ScrollView>(null);
  const [page, setPage] = useState(0);
  const potm = getPlayer(potmId);
  const rated = Object.values(ratings).filter((v) => v >= 1).length;

  const goTo = (i: number) => {
    setPage(i);
    pager.current?.scrollTo({ x: i * width, animated: true });
  };

  return (
    <View>
      <Panel style={styles.potmPanel}>
        <Feather name="award" size={20} color={potm ? t.red : t.inkFaint} />
        <View style={{ flex: 1, marginLeft: 12 }}>
          <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
            Your player of the match
          </Txt>
          {potm ? (
            <Txt variant="h3">{potm.name}</Txt>
          ) : (
            <Txt variant="bodySmall" tone="faint">
              Optional. Tap the medal next to a player.
            </Txt>
          )}
        </View>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
          {rated} rated
        </Txt>
      </Panel>

      <View style={styles.tabs}>
        {teams.map(({ team, players }, i) => {
          const on = i === page;
          const n = players.filter((p) => (ratings[p.id] ?? 0) >= 1).length;
          return (
            <Pressable
              key={team.id}
              onPress={() => goTo(i)}
              accessibilityRole="tab"
              accessibilityState={{ selected: on }}
              accessibilityLabel={`${team.name}, ${n} of ${players.length} rated`}
              style={[styles.tab, { borderColor: on ? t.ink : t.rule, backgroundColor: on ? t.ink : t.card, marginLeft: i === 0 ? 0 : 8 }]}>
              <View>
                <Txt variant="score" numberOfLines={1} style={{ fontSize: 20, lineHeight: 22, color: on ? t.onInk : t.ink }}>
                  {team.short}
                </Txt>
                <View style={[styles.tabBar, { backgroundColor: team.colour }]} />
              </View>
              <Txt variant="monoSmall" style={{ color: on ? t.onInk : t.inkSoft, marginLeft: 10 }}>
                {n}/{players.length}
              </Txt>
            </Pressable>
          );
        })}
      </View>

      <ScrollView
        ref={pager}
        horizontal
        pagingEnabled
        showsHorizontalScrollIndicator={false}
        style={{ marginHorizontal: -GUTTER }}
        onMomentumScrollEnd={(e) => setPage(Math.round(e.nativeEvent.contentOffset.x / width))}>
        {teams.map(({ team, players }, i) => (
          <View key={team.id} style={{ width, paddingHorizontal: GUTTER }}>
            <View style={[styles.list, { backgroundColor: t.card }]}>
              {players.map((p, j) => (
                <PlayerRow
                  key={p.id}
                  player={p}
                  value={ratings[p.id] ?? 0}
                  potm={potmId === p.id}
                  last={j === players.length - 1}
                  onRate={onRate}
                  onPotm={onPotm}
                />
              ))}
            </View>
            {i < teams.length - 1 ? (
              <Pressable onPress={() => goTo(i + 1)} style={styles.swipeHint} accessibilityRole="button">
                <Txt tone="red" style={{ fontFamily: fonts.uiBold, fontSize: 14 }}>
                  Swipe for {teams[i + 1].team.short} →
                </Txt>
              </Pressable>
            ) : i > 0 ? (
              <Pressable onPress={() => goTo(i - 1)} style={styles.swipeHint} accessibilityRole="button">
                <Txt tone="soft" style={{ fontFamily: fonts.uiBold, fontSize: 14 }}>
                  ← Back to {teams[i - 1].team.short}
                </Txt>
              </Pressable>
            ) : null}
          </View>
        ))}
      </ScrollView>
      <Txt variant="bodySmall" tone="faint" align="center" style={{ marginTop: 4 }}>
        Untouched players are skipped. Slide to 0 to skip again.
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  potmPanel: { flexDirection: 'row', alignItems: 'center', marginTop: 12, paddingVertical: 14 },
  tabs: { flexDirection: 'row', marginTop: 12, marginBottom: 12 },
  tab: { flex: 1, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', borderWidth: 1, height: 48, paddingHorizontal: 10 },
  tabBar: { height: 3, marginTop: 1 },
  list: { paddingHorizontal: 14 },
  player: { paddingTop: 14, paddingBottom: 6 },
  playerTop: { flexDirection: 'row', alignItems: 'center' },
  crown: { width: 40, height: 40, borderWidth: 1, alignItems: 'center', justifyContent: 'center' },
  value: { width: 40, textAlign: 'right', fontSize: 24, lineHeight: 28 },
  skip: { width: 64, height: 44, alignItems: 'flex-end', justifyContent: 'center' },
  slider: { height: 36, marginTop: 2 },
  swipeHint: { alignSelf: 'flex-end', justifyContent: 'center', minHeight: 44, paddingTop: 4 },
});
