import { useLocalSearchParams } from 'expo-router';

import { LogFlow } from '@/features/log/LogFlow';

/** Full-screen modal (see src/app/_layout.tsx). `matchId` (from the Match Page or the Vault) skips "Pick the match". */
export default function LogScreen() {
  const { matchId } = useLocalSearchParams<{ matchId?: string }>();
  const id = Array.isArray(matchId) ? matchId[0] : matchId;
  return <LogFlow key={id ?? 'pick'} presetMatchId={id || undefined} />;
}
