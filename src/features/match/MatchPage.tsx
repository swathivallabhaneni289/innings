import { View } from 'react-native';

import { SectionHeader } from '@/components';
import { useMe, type MatchPageResponse } from '@/data';
import { ChatEntry } from './ChatEntry';
import { LiveCalls } from './LiveCalls';
import { FanRating } from './FanRating';
import { Header } from './Header';
import { FriendsWatching, LiveScoreboard, ResultCard } from './LiveScoreboard';
import { ReviewBox } from './ReviewBox';
import { PlayerRatings } from './PlayerRatings';
import { PredictionResults } from './PredictionResults';
import { PredictionSection, YourPicks } from './PredictionSection';
import { Reviews } from './Reviews';
import { WatchingButton } from './Watching';
import { WatchSummary } from './WatchSummary';
import { YourTake } from './YourTake';

/** The Match Page body. Its layout follows the match: BEFORE (upcoming), DURING (live), AFTER (completed and every Vault classic). */
export function MatchPage({ data }: { data: MatchPageResponse }) {
  const { match } = data;
  const phase = match.kind === 'classic' || match.state === 'completed' ? 'after' : match.state === 'live' ? 'live' : 'before';
  return (
    <View>
      <Header match={match} />
      {phase === 'before' ? <Before data={data} /> : null}
      {phase === 'live' ? <Live data={data} /> : null}
      {phase === 'after' ? <After data={data} /> : null}
    </View>
  );
}

function Before({ data }: { data: MatchPageResponse }) {
  return (
    <View>
      <PredictionSection data={data} />
      <ChatEntry match={data.match} />
      <SectionHeader title="Where you're watching" />
      <WatchingButton matchId={data.match.id} watching={data.mine.watching} />
      {data.friendsWatching.length > 0 ? (
        <View style={{ marginTop: 18 }}>
          <SectionHeader title="Friends tuning in" style={{ marginTop: 0 }} />
          <FriendsWatching users={data.friendsWatching} />
        </View>
      ) : null}
    </View>
  );
}

function Live({ data }: { data: MatchPageResponse }) {
  return (
    <View>
      <LiveScoreboard match={data.match} />
      {data.calls?.length ? (
        <>
          <SectionHeader title="Call it" />
          <LiveCalls calls={data.calls} match={data.match} />
        </>
      ) : null}
      <ChatEntry match={data.match} />
      <SectionHeader title="Friends watching now" />
      <FriendsWatching users={data.friendsWatching} emptyText="None of your friends are watching this one. Yet." />
      <View style={{ marginTop: 12 }}>
        <WatchingButton matchId={data.match.id} watching={data.mine.watching} />
      </View>
      <SectionHeader title="Your picks" />
      <YourPicks match={data.match} mine={data.mine.prediction} />
    </View>
  );
}

function After({ data }: { data: MatchPageResponse }) {
  const me = useMe();
  const { match, fans, results, mine } = data;
  const classic = match.kind === 'classic';
  return (
    <View>
      <ResultCard match={match} />
      <View style={{ marginTop: 12 }}>
        <YourTake match={match} log={mine.log} />
      </View>
      <ChatEntry match={match} />
      {data.calls?.length ? (
        <>
          <SectionHeader title="Calls during the match" />
          <LiveCalls calls={data.calls} match={data.match} />
        </>
      ) : null}

      {fans ? (
        <>
          <SectionHeader title="Fan rating" />
          <FanRating fans={fans} />
          <SectionHeader title="Player ratings" />
          <PlayerRatings match={match} fans={fans} />
        </>
      ) : null}

      {!classic || results.length > 0 ? (
        <>
          <SectionHeader title="Who called it" />
          <PredictionResults results={results} meId={me.data?.id} />
        </>
      ) : null}

      {fans ? (
        <>
          <SectionHeader title="Reviews" />
          <ReviewBox match={match} log={mine.log} />
          <Reviews fans={fans} />
          <SectionHeader title="Where fans watched" />
          <WatchSummary watch={fans.watch} />
        </>
      ) : null}
    </View>
  );
}
