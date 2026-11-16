import { QueryState, Screen } from '@/components';
import { useUserProfile } from '@/data';
import { Top5Editor } from '@/features/vault/Top5Editor';

/** /top5 — put the fan's Top 5 right: reorder it or take games out. Opened from Edit in the Vault and on the profile. */
export default function Top5Screen() {
  const me = useUserProfile('me');
  return (
    <Screen title="Edit your Top 5" back>
      <QueryState q={me} loadingLabel="Fetching your Top 5…">
        {(data) => <Top5Editor top5={data.top5} />}
      </QueryState>
    </Screen>
  );
}
