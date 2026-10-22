import { Feather } from '@expo/vector-icons';
import type { Ref } from 'react';
import { StyleSheet, View } from 'react-native';

import { BallRating, MatchTicket, Paper, Rule, Txt } from '@/components';
import { playerName } from '@/data';
import type { Match, MatchLog } from '@/domain/types';
import { placeLine } from '@/features/log/draft';
import { RatingLine } from '@/features/log/RatingLine';
import { stubNo as fmtStubNo } from '@/lib/format';
import { fonts, useTheme } from '@/theme';

/** Size of the card in points; captured at the device's pixel ratio (1080×1920 on a 3× phone). */
export const STORY_W = 360;
export const STORY_H = 640;

/**
 * The 9:16 Instagram-story card on the concrete ground: the lowercase innings wordmark, the stamped ticket,
 * your balls and verdict, where you watched, a line of your review, your Player of the Match, and the
 * innings.app footer. Rendered off-screen and captured.
 */
export function StoryCard({ match, log, ref }: { match: Match; log: MatchLog; ref?: Ref<View> }) {
  const t = useTheme();
  const review = log.review?.trim();
  return (
    <View ref={ref} collapsable={false} style={styles.frame}>
      <Paper style={styles.paper}>
        <View style={styles.body}>
          <View style={styles.brandRow}>
            <Txt style={[styles.wordmark, { color: t.ink }]}>innings</Txt>
            <Txt variant="mono" tone="red">
              {fmtStubNo(log.stubNo)}
            </Txt>
          </View>

          <Txt variant="h2" style={{ marginTop: 18, marginBottom: 10 }}>
            I watched this one
          </Txt>
          <MatchTicket match={match} size="lg" rating={log.rating} stubNo={log.stubNo} instantClassic={log.instantClassic} stamped />

          <View style={[styles.card, { backgroundColor: t.card }]}>
            <BallRating value={log.rating} size={22} gap={5} />
            <RatingLine
              rating={log.rating}
              size="sm"
              word={log.instantClassic ? '❤️ Instant classic' : undefined}
              style={{ marginTop: 8 }}
            />
            <Txt variant="body" numberOfLines={2} style={{ marginTop: 6 }}>
              {placeLine(log)}
            </Txt>

            {review ? (
              <>
                <Rule style={{ marginVertical: 12 }} />
                <Txt variant="review" numberOfLines={2}>
                  “{review}”
                </Txt>
              </>
            ) : null}

            {log.fanPotmId ? (
              <>
                <Rule style={{ marginVertical: 12 }} />
                <View style={styles.potm}>
                  <Feather name="award" size={18} color={t.red} />
                  <View style={{ marginLeft: 10, flex: 1 }}>
                    <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
                      My player of the match
                    </Txt>
                    <Txt variant="h3" numberOfLines={1} style={{ fontSize: 16 }}>
                      {playerName(log.fanPotmId)}
                    </Txt>
                  </View>
                </View>
              </>
            ) : null}
          </View>

          <View style={{ flex: 1 }} />
          <View style={styles.footer}>
            <Txt variant="mono" tone="ink">
              innings.app
            </Txt>
            <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.ui }}>
              Log every match you watch
            </Txt>
          </View>
        </View>
      </Paper>
    </View>
  );
}

const styles = StyleSheet.create({
  frame: { width: STORY_W, height: STORY_H, overflow: 'hidden' },
  paper: { width: STORY_W, height: STORY_H, flex: 0 },
  body: { flex: 1, paddingHorizontal: 22, paddingTop: 26, paddingBottom: 20 },
  brandRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  wordmark: { fontFamily: fonts.heading, fontSize: 30, lineHeight: 36, letterSpacing: -0.6 },
  card: { padding: 16, marginTop: 14 },
  potm: { flexDirection: 'row', alignItems: 'center' },
  footer: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
});
