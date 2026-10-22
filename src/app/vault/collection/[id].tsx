import { useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { Empty, Panel, QueryState, Screen, Txt } from '@/components';
import { useVault } from '@/data';
import type { VaultResponse } from '@/domain/api';
import { COLLECTIONS, MAX_TOP5 } from '@/domain/constants';
import { ClassicRow } from '@/features/vault/ClassicRow';
import { useTop5 } from '@/features/vault/top5';
import { fonts } from '@/theme';

export default function CollectionScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const vault = useVault();
  const title = vault.data?.collections.find((s) => s.collection.id === id)?.collection.title ?? COLLECTIONS.find((c) => c.id === id)?.title ?? 'Collection';
  // The spinner shows only while the fan pulls, never over loaded tickets during a background refetch.
  const [pulling, setPulling] = useState(false);

  const refresh = async () => {
    setPulling(true);
    try {
      await vault.refetch();
    } finally {
      setPulling(false);
    }
  };

  return (
    <Screen back title={title} onRefresh={refresh} refreshing={pulling}>
      <QueryState q={vault} loadingLabel="Opening the archive…">
        {(v) => <CollectionBody vault={v} id={id} />}
      </QueryState>
    </Screen>
  );
}

function CollectionBody({ vault, id }: { vault: VaultResponse; id: string | undefined }) {
  const shelf = vault.collections.find((s) => s.collection.id === id);
  const pin = useTop5(vault.top5.map((m) => m.id));
  if (!shelf) return <Empty title="Lost ball" body="This collection isn't in the Vault." />;
  const { collection, matches } = shelf;
  return (
    <View>
      <Txt variant="body" tone="soft">
        {collection.blurb}
      </Txt>
      <View style={styles.meta}>
        <Txt variant="bodySmall" tone="soft" style={styles.metaText}>
          {matches.length} {matches.length === 1 ? 'match' : 'matches'}
        </Txt>
        <Txt variant="bodySmall" tone={pin.full ? 'red' : 'soft'} style={styles.metaText}>
          {pin.full ? `Top ${MAX_TOP5} is full` : `${vault.top5.length} of ${MAX_TOP5} pinned to your Top ${MAX_TOP5}`}
        </Txt>
      </View>
      {matches.length === 0 ? (
        <Panel>
          <Empty title="Covers on" body="No matches in this collection yet." />
        </Panel>
      ) : (
        matches.map((item) => (
          <ClassicRow
            key={item.match.id}
            item={item}
            pinned={pin.isPinned(item.match.id)}
            busy={pin.busyId === item.match.id}
            message={pin.message?.matchId === item.match.id ? pin.message.text : undefined}
            onTogglePin={() => pin.toggle(item.match.id)}
          />
        ))
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  meta: { flexDirection: 'row', justifyContent: 'space-between', marginTop: 14, marginBottom: 14 },
  metaText: { fontFamily: fonts.label },
});
