import { useLocalSearchParams } from 'expo-router';
import { useState } from 'react';

import { KeyboardAware, QueryState, Screen } from '@/components';
import { useMatchPage } from '@/data';
import { MatchPage } from '@/features/match/MatchPage';

/**
 * /match/[id] — fixtures and Vault classics alike. The page refreshes itself every 20 s for live scores. Its review box
 * sits near the bottom, so the page scrolls it above the keyboard.
 */
export default function MatchScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const q = useMatchPage(id);
  const [pulling, setPulling] = useState(false);
  const refresh = async () => {
    setPulling(true);
    try {
      await q.refetch();
    } finally {
      setPulling(false);
    }
  };
  return (
    <KeyboardAware scrollInsets>
      <Screen back onRefresh={refresh} refreshing={pulling} keyboardInsets>
        <QueryState q={q} loadingLabel="Walking out to the middle…">
          {(data) => <MatchPage data={data} />}
        </QueryState>
      </Screen>
    </KeyboardAware>
  );
}
