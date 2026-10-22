import { router } from 'expo-router';
import { Pressable, StyleSheet, useWindowDimensions, View } from 'react-native';

import { BallRating, Empty, Panel, Rule, Txt } from '@/components';
import type { HomeResponse } from '@/domain/api';
import { competitionShort } from '@/domain/constants';
import { fonts } from '@/theme';
import { TeamCodes } from './TeamCodes';

type Item = HomeResponse['justFinished'][number];

const VERDICT_W = 92;

/** Matches that ended in the last 24 h: the result, the fans' verdict and the review everyone's liking. */
export function JustFinished({ items }: { items: Item[] }) {
  if (items.length === 0) {
    return <Empty title="Nothing at the close yet" body="No matches have finished in the last 24 hours." />;
  }
  return (
    <View>
      {items.map((it) => (
        <FinishedRow key={it.match.id} item={it} />
      ))}
    </View>
  );
}

function FinishedRow({ item }: { item: Item }) {
  const { width } = useWindowDimensions();
  const { match, ratingAvg, ratingCount, topReview } = item;
  const rated = ratingCount > 0;
  return (
    <Pressable onPress={() => router.push('/match/' + match.id)} accessibilityRole="button" style={({ pressed }) => ({ opacity: pressed ? 0.9 : 1, marginBottom: 12 })}>
      <Panel style={{ padding: 14 }}>
        <View style={styles.top}>
          <View style={{ flex: 1, marginRight: 12 }}>
            <Txt variant="label" tone="soft" numberOfLines={1} style={{ fontSize: 10 }}>
              {competitionShort(match.competitionId)} · {match.stage}
            </Txt>
            <TeamCodes teamIds={match.teamIds} winnerId={match.result?.winnerId} width={width - 40 - 28 - VERDICT_W - 12} max={30} min={20} style={{ marginTop: 4 }} />
            <Txt variant="bodySmall" numberOfLines={2} style={{ fontFamily: fonts.ui, marginTop: 4 }}>
              {match.result?.text ?? 'Result to come'}
            </Txt>
          </View>
          <View style={styles.verdict}>
            <Txt variant="stat" style={styles.avg}>
              {rated ? ratingAvg.toFixed(1) : '–'}
            </Txt>
            <BallRating value={rated ? Math.round(ratingAvg * 2) / 2 : 0} size={12} gap={2} />
            <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4, fontSize: 12, lineHeight: 16 }}>
              {rated ? (
                <>
                  <Txt variant="mono" tone="soft" style={{ fontSize: 11.5 }}>
                    {ratingCount.toLocaleString()}
                  </Txt>{' '}
                  {ratingCount === 1 ? 'rating' : 'ratings'}
                </>
              ) : (
                'No ratings yet'
              )}
            </Txt>
          </View>
        </View>
        {topReview?.log.review ? (
          <>
            <Rule style={{ marginVertical: 12 }} />
            <Txt variant="body" numberOfLines={3}>
              “{topReview.log.review}”
            </Txt>
            <Txt variant="bodySmall" tone="soft" style={{ marginTop: 6 }}>
              <Txt variant="bodySmall" style={{ fontFamily: fonts.ui }}>
                {topReview.user.name}
              </Txt>
              {'  ·  '}
              {topReview.likes} {topReview.likes === 1 ? 'like' : 'likes'}
            </Txt>
          </>
        ) : null}
      </Panel>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  top: { flexDirection: 'row', alignItems: 'flex-start' },
  verdict: { width: VERDICT_W, alignItems: 'flex-end' },
  avg: { fontSize: 36, lineHeight: 38 },
});
