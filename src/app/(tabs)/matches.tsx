import { useRef, useState } from 'react';
import { View } from 'react-native';

import { Button, Empty, QueryState, Rule, Screen, Txt } from '@/components';
import { useMatches, type MatchesResponse, type MatchFilter } from '@/data';
import { MatchFilters, SCOPE_CHIPS, type ScopeChip, type StateFilter } from '@/features/matches/MatchFilters';
import { MatchList } from '@/features/matches/MatchList';

function emptyCopy(state: StateFilter | null, scope: ScopeChip | undefined): { title: string; body: string } {
  const where = scope ? ` ${scope.phrase}` : '';
  switch (state) {
    case 'live':
      return { title: 'No play right now', body: `Nothing live${where} at the moment. The covers are on.` };
    case 'upcoming':
      return { title: 'Nothing on the fixture list', body: `No upcoming matches${where} yet. Check back after the next announcement.` };
    case 'recent':
      return { title: 'No recent results', body: `No finished matches${where} to rate just now.` };
    default:
      return { title: 'No play today', body: scope ? `No matches${where} around right now.` : 'No fixtures around right now. Check back after the toss.' };
  }
}

export default function MatchesScreen() {
  const [state, setState] = useState<StateFilter | null>(null);
  const [scopeId, setScopeId] = useState<string | null>(null);
  const [pulling, setPulling] = useState(false);
  const scope = SCOPE_CHIPS.find((c) => c.id === scopeId);
  const filter: MatchFilter = { ...(state ? { state } : null), ...scope?.filter };
  const q = useMatches(filter);

  // Keep the previous list on screen (dimmed) while a new filter loads, instead of flashing a spinner.
  const last = useRef<MatchesResponse | undefined>(undefined);
  if (q.data) last.current = q.data;
  const stale = q.data === undefined && !q.error ? last.current : undefined;
  const shown = { data: q.data ?? stale, isLoading: q.isLoading, error: q.error, refetch: q.refetch };

  const refresh = async () => {
    setPulling(true);
    try {
      await q.refetch();
    } finally {
      setPulling(false);
    }
  };
  const clear = () => {
    setState(null);
    setScopeId(null);
  };

  const header = (
    <View style={{ paddingTop: 8 }}>
      <View style={{ paddingHorizontal: 20 }}>
        <Txt variant="h1">Matches</Txt>
      </View>
      <MatchFilters state={state} scope={scopeId} onState={setState} onScope={setScopeId} />
      <Rule />
    </View>
  );

  return (
    <Screen header={header} onRefresh={refresh} refreshing={pulling}>
      <QueryState q={shown} loadingLabel="Reading the fixture list…">
        {(data) =>
          data.matches.length === 0 ? (
            <Empty
              {...emptyCopy(state, scope)}
              action={state || scope ? <Button label="Clear filters" kind="outline" size="sm" onPress={clear} /> : undefined}
            />
          ) : (
            <View style={{ opacity: stale ? 0.45 : 1 }}>
              <MatchList data={data} />
            </View>
          )
        }
      </QueryState>
    </Screen>
  );
}
