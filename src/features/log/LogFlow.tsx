import * as Haptics from 'expo-haptics';
import { router } from 'expo-router';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Alert, ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { Button, Empty, KeyboardAware, Loading, QueryState, Rule, Screen, Txt } from '@/components';
import { squads, useActions, useMatchPage, useNow, useUserProfile, type FeedItem } from '@/data';
import type { Match, MatchLog } from '@/domain/types';
import { StoryCard, STORY_H } from '@/features/share/StoryCard';
import { ForceTheme } from '@/theme';
import { useShareStory } from '@/features/share/useShareStory';
import { stubNo as fmtStubNo } from '@/lib/format';
import { useTheme } from '@/theme';
import { draftFromLog, emptyDraft, toNewLog, type Draft, type StepId } from './draft';
import { LogHeader } from './LogHeader';
import { MatchLine } from './MatchLine';
import { PickMatch } from './PickMatch';
import { Printed } from './Printed';
import { RateMatch } from './RateMatch';
import { RatePlayers } from './RatePlayers';
import { WhereWatched, type WhereValue } from './WhereWatched';

const hasXI = (m: Match) => squads(m).some((s) => s.players.length > 0);
const leave = () => (router.canGoBack() ? router.back() : router.replace('/'));

/**
 * The centre-button Log flow, four quick steps: pick the match → rate it → rate the players → where I watched it.
 * Save prints the ticket stub with the LOGGED stamp. A preset `matchId` (Match Page, Vault) skips step 1;
 * classics have no XI, so the players step is skipped too. Re-logging pre-fills from the existing stub.
 */
export function LogFlow({ presetMatchId }: { presetMatchId?: string }) {
  const t = useTheme();
  const insets = useSafeAreaInsets();
  const now = useNow(60_000);
  const actions = useActions();
  const me = useUserProfile('me');
  const scroll = useRef<ScrollView>(null);
  const story = useShareStory();

  const [skipPick, setSkipPick] = useState(!!presetMatchId);
  const [step, setStep] = useState<StepId | 'printed'>(presetMatchId ? 'rating' : 'match');
  const [draft, setDraft] = useState<Draft>(() => emptyDraft(presetMatchId ?? null));
  const [prefilledFor, setPrefilledFor] = useState<string | null>(null);
  const [touched, setTouched] = useState(false);
  const [errors, setErrors] = useState<{ rating?: string; place?: string; save?: string }>({});
  const [saved, setSaved] = useState<{ item: FeedItem; relogged: boolean } | null>(null);

  const page = useMatchPage(draft.matchId ?? undefined);
  const pageData = page.data && page.data.match.id === draft.matchId ? page.data : undefined;
  const match = pageData?.match;
  const existing = pageData?.mine.log ?? null;
  const ready = !!pageData && prefilledFor === draft.matchId;

  // Load the chosen match once: re-logging starts from the old stub, a new log starts blank.
  useEffect(() => {
    if (!pageData || prefilledFor === pageData.match.id) return;
    const log = pageData.mine.log;
    setDraft(log ? draftFromLog(log) : emptyDraft(pageData.match.id));
    setPrefilledFor(pageData.match.id);
  }, [pageData, prefilledFor]);

  // Default place: wherever the fan logged from last, else Home.
  const lastPlace = useMemo(() => {
    let last: MatchLog | null = null;
    for (const fi of me.data?.logs ?? []) if (!last || fi.log.createdAt > last.createdAt) last = fi.log;
    return last?.watchedAt ?? null;
  }, [me.data]);
  const place = draft.watchedAt ?? lastPlace ?? 'home';

  const steps: StepId[] | null =
    step !== 'match' && draft.matchId && !match
      ? null
      : [...(skipPick ? [] : (['match'] as StepId[])), 'rating', ...(match && !hasXI(match) ? [] : (['players'] as StepId[])), 'place'];
  const idx = steps && step !== 'printed' ? steps.indexOf(step) : -1;

  const go = (s: StepId | 'printed') => {
    setErrors({});
    setStep(s);
    scroll.current?.scrollTo({ y: 0, animated: false });
  };

  const edit = useCallback((fn: (d: Draft) => Draft) => {
    setTouched(true);
    setDraft(fn);
  }, []);

  const onRatePlayer = useCallback(
    (pid: string, v: number) => edit((d) => ((d.playerRatings[pid] ?? 0) === v ? d : { ...d, playerRatings: { ...d.playerRatings, [pid]: v } })),
    [edit],
  );
  const onPotm = useCallback((pid: string) => edit((d) => ({ ...d, fanPotmId: d.fanPotmId === pid ? null : pid })), [edit]);

  const onWhere = (patch: Partial<WhereValue>) => {
    const { place: p, ...rest } = patch;
    edit((d) => ({ ...d, ...rest, ...(p ? { watchedAt: p } : null) }));
    if (p) setErrors((e) => ({ ...e, place: undefined }));
  };

  const pick = (m: Match) => {
    Haptics.selectionAsync();
    setDraft((d) => (d.matchId === m.id ? d : { ...d, matchId: m.id }));
    go('rating');
  };

  const save = async () => {
    if (!match) return;
    if (draft.rating <= 0) {
      setStep('rating');
      setErrors({ rating: 'Give it at least half a ball before you save.' });
      return;
    }
    if (!place) {
      setErrors({ place: 'Pick where you watched it.' });
      return;
    }
    const xi = new Set(Object.values(match.xi ?? {}).flat());
    const relogged = !!existing;
    try {
      const item = await actions.saveLog(toNewLog(draft, match.id, place, xi));
      setSaved({ item, relogged });
      go('printed');
    } catch (e) {
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Error);
      setErrors({ save: e instanceof Error ? e.message : "Couldn't save your log. Try again." });
    }
  };

  const next = () => {
    if (!steps) return;
    if (step === 'rating' && draft.rating <= 0) {
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Warning);
      setErrors({ rating: 'Give it at least half a ball.' });
      return;
    }
    const n = steps[idx + 1];
    if (n) {
      Haptics.selectionAsync();
      go(n);
    } else save();
  };

  const back = () => {
    const p = steps?.[idx - 1];
    if (p) go(p);
  };

  const close = () => {
    if (step === 'printed' || !touched) return leave();
    Alert.alert('Tear up this log?', "Your rating and notes won't be saved.", [
      { text: 'Keep logging', style: 'cancel' },
      { text: 'Discard', style: 'destructive', onPress: leave },
    ]);
  };

  const notStarted = match?.state === 'upcoming';
  const showFooter = step !== 'match' && step !== 'printed' && ready && !notStarted;
  const last = steps ? idx === steps.length - 1 : false;
  const ratedPlayers = Object.values(draft.playerRatings).filter((v) => v >= 1).length;
  const nextLabel = last ? 'Save & print' : step === 'players' && ratedPlayers === 0 && !draft.fanPotmId ? 'Skip players' : 'Next';

  const body = () => {
    if (step === 'printed' && saved) return <Printed item={saved.item} relogged={saved.relogged} story={story} onDone={leave} />;
    if (step === 'match' || !draft.matchId) return <PickMatch selectedId={draft.matchId} onPick={pick} />;
    return (
      <QueryState q={page} loadingLabel="Fetching the scorecard…">
        {() => {
          if (!ready || !match) return <Loading label="Fetching the scorecard…" />;
          if (notStarted)
            return (
              <Empty
                title="Not started yet"
                body="You can log this one once the first ball is bowled."
                action={
                  <Button
                    label="Pick another match"
                    kind="outline"
                    onPress={() => {
                      setSkipPick(false);
                      go('match');
                    }}
                  />
                }
              />
            );
          return (
            <View>
              <MatchLine m={match} now={now} stubNo={existing?.stubNo} />
              {step === 'rating' && existing ? (
                <Txt variant="bodySmall" tone="soft" style={{ marginTop: 10 }}>
                  You logged this one before. Your take is filled in, and saving keeps {fmtStubNo(existing.stubNo)}.
                </Txt>
              ) : null}
              {step === 'rating' ? (
                <RateMatch
                  rating={draft.rating}
                  instantClassic={draft.instantClassic}
                  onRating={(rating) => {
                    edit((d) => ({ ...d, rating }));
                    setErrors((e) => ({ ...e, rating: undefined }));
                  }}
                  onInstantClassic={(instantClassic) => edit((d) => ({ ...d, instantClassic }))}
                  error={errors.rating}
                />
              ) : null}
              {step === 'players' ? (
                <RatePlayers match={match} ratings={draft.playerRatings} potmId={draft.fanPotmId} onRate={onRatePlayer} onPotm={onPotm} />
              ) : null}
              {step === 'place' ? (
                <WhereWatched
                  value={{
                    place,
                    placeNote: draft.placeNote,
                    companions: draft.companions,
                    watchedWith: draft.watchedWith,
                    photoUri: draft.photoUri,
                    review: draft.review,
                  }}
                  defaulted={draft.watchedAt ? null : lastPlace ? 'last' : 'home'}
                  onChange={onWhere}
                  onReviewFocus={() => setTimeout(() => scroll.current?.scrollToEnd({ animated: true }), 250)}
                  error={errors.place}
                />
              ) : null}
            </View>
          );
        }}
      </QueryState>
    );
  };

  return (
    <Screen
      scroll={false}
      header={
        <LogHeader
          steps={steps}
          current={draft.matchId || step === 'printed' ? step : 'match'}
          onClose={close}
          onJump={go}
        />
      }>
      <KeyboardAware>
        <ScrollView
          ref={scroll}
          style={{ flex: 1 }}
          contentContainerStyle={[styles.content, { paddingBottom: showFooter ? 24 : insets.bottom + 32 }]}
          keyboardShouldPersistTaps="handled"
          keyboardDismissMode="on-drag"
          showsVerticalScrollIndicator={false}>
          {body()}
        </ScrollView>
        {showFooter ? (
          <View style={[styles.footer, { paddingBottom: Math.max(insets.bottom, 12), backgroundColor: t.paper }]}>
            <Rule />
            {errors.save ? (
              <Txt variant="bodySmall" tone="red" style={{ marginTop: 8 }}>
                {errors.save}
              </Txt>
            ) : null}
            <View style={styles.footerRow}>
              {idx > 0 ? <Button label="Back" kind="outline" size="lg" onPress={back} style={{ minWidth: 96 }} /> : null}
              <Button
                label={nextLabel}
                size="lg"
                onPress={next}
                loading={last && actions.pending}
                style={{ flex: 1, marginLeft: idx > 0 ? 10 : 0 }}
              />
            </View>
          </View>
        ) : null}
      </KeyboardAware>
      {saved ? (
        // The 9:16 story card, mounted off-screen (outside the scroll view) so it's laid out and stamped before Share.
        <View pointerEvents="none" style={styles.offscreen}>
          <ForceTheme mode="paper">
            <StoryCard ref={story.ref} match={saved.item.match} log={saved.item.log} />
          </ForceTheme>
        </View>
      ) : null}
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { paddingHorizontal: 20, paddingTop: 4 },
  footer: { paddingHorizontal: 20 },
  footerRow: { flexDirection: 'row', marginTop: 10 },
  offscreen: { position: 'absolute', left: -10000, top: -STORY_H - 100 },
});
