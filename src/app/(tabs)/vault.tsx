import { router } from 'expo-router';
import { useMemo } from 'react';
import { ScrollView, View } from 'react-native';

import { Empty, Panel, QueryState, Screen, SectionHeader, Txt } from '@/components';
import { useRankings, useVault } from '@/data';
import type { VaultResponse } from '@/domain/api';
import { ClassicCard } from '@/features/vault/ClassicCard';
import { OnThisDay } from '@/features/vault/OnThisDay';
import { RankRow } from '@/features/vault/RankRow';
import { Top5Pager } from '@/features/vault/Top5Pager';

export default function VaultScreen() {
  const vault = useVault();
  const rankings = useRankings({ format: 'all', era: 'all' });
  return (
    <Screen
      title="The Vault"
      onRefresh={() => {
        vault.refetch();
        rankings.refetch();
      }}
      refreshing={vault.isRefetching}>
      <Txt variant="body" tone="soft" style={{ marginBottom: 4 }}>
        Every great game you were part of.
      </Txt>
      <QueryState q={vault} loadingLabel="Opening the archive…">
        {(v) => <VaultBody vault={v} />}
      </QueryState>
      {vault.data ? (
        <>
          <SectionHeader title="All-time fan rankings" action="Full rankings" onAction={() => router.push('/vault/rankings')} />
          <QueryState q={rankings}>
            {(r) =>
              r.matches.length === 0 ? (
                <Panel>
                  <Empty title="No ratings yet" body="The jury's still out. Rate a classic to get the table started." />
                </Panel>
              ) : (
                <Panel style={{ paddingVertical: 2 }}>
                  {r.matches.slice(0, 5).map((item, i, top) => (
                    <RankRow key={item.match.id} rank={i + 1} item={item} compact last={i === top.length - 1} />
                  ))}
                </Panel>
              )
            }
          </QueryState>
        </>
      ) : null}
    </Screen>
  );
}

function VaultBody({ vault }: { vault: VaultResponse }) {
  // Average fan ratings by match id, so pinned tickets can print the crowd's verdict.
  const ratings = useMemo(() => {
    const out: Record<string, number> = {};
    for (const shelf of vault.collections) for (const r of shelf.matches) if (r.ratingCount > 0) out[r.match.id] = r.ratingAvg;
    for (const [id, r] of Object.entries(vault.fanRatings ?? {})) if (r.count > 0) out[id] = r.avg;
    return out;
  }, [vault.collections, vault.fanRatings]);

  return (
    <View>
      <SectionHeader title="On this day" />
      {vault.onThisDay ? (
        <OnThisDay data={vault.onThisDay} />
      ) : (
        <Panel>
          <Empty title="No play today" body="Nothing in the archive for this date. Check back tomorrow." />
        </Panel>
      )}

      <SectionHeader title="Your Top 5" />
      <Top5Pager top5={vault.top5} ratings={ratings} />

      {vault.collections.map(({ collection, matches }) => (
        <View key={collection.id}>
          <SectionHeader
            title={collection.title}
            action={matches.length > 0 ? 'See all' : undefined}
            onAction={() => router.push(`/vault/collection/${collection.id}`)}
            style={{ marginBottom: 4 }}
          />
          <Txt variant="bodySmall" tone="soft" style={{ marginBottom: 12 }}>
            {collection.blurb}
          </Txt>
          {matches.length === 0 ? (
            <Panel>
              <Empty title="Covers on" body="No matches in this collection yet." />
            </Panel>
          ) : (
            <ScrollView
              horizontal
              showsHorizontalScrollIndicator={false}
              style={{ marginHorizontal: -20 }}
              contentContainerStyle={{ paddingHorizontal: 20 }}>
              {matches.map((item) => (
                <ClassicCard key={item.match.id} item={item} />
              ))}
            </ScrollView>
          )}
        </View>
      ))}
    </View>
  );
}
