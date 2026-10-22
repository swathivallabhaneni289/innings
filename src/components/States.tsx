import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { SpinningBall } from './Motion';
import { Txt } from './Txt';
import { Button } from './ui';

export function Loading({ label = 'Rolling the arm over…' }: { label?: string }) {
  return (
    <View style={styles.center}>
      <SpinningBall spinKey={1} size={30} />
      <Txt variant="bodySmall" tone="soft" style={{ marginTop: 12 }}>
        {label}
      </Txt>
    </View>
  );
}

export function ErrorState({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  const message = error instanceof Error ? error.message : 'Something went wrong.';
  return (
    <View style={styles.center}>
      <Txt variant="h2">Rain stopped play</Txt>
      <Txt variant="bodySmall" tone="soft" align="center" style={{ marginTop: 8, maxWidth: 300 }}>
        {message}
      </Txt>
      {onRetry ? <Button label="Try again" kind="outline" size="sm" onPress={onRetry} style={{ marginTop: 16 }} /> : null}
    </View>
  );
}

/** Empty state in the voice of the app ("No play today"). */
export function Empty({ title, body, action }: { title: string; body?: string; action?: ReactNode }) {
  return (
    <View style={[styles.center, { paddingVertical: 28 }]}>
      <Txt variant="h3" align="center">
        {title}
      </Txt>
      {body ? (
        <Txt variant="review" tone="soft" align="center" style={{ marginTop: 6, maxWidth: 300 }}>
          {body}
        </Txt>
      ) : null}
      {action ? <View style={{ marginTop: 14 }}>{action}</View> : null}
    </View>
  );
}

/** Renders loading / error for a react-query result, then `children(data)`. */
export function QueryState<T>({
  q,
  children,
  loadingLabel,
}: {
  q: { data?: T; isLoading: boolean; error: unknown; refetch: () => unknown };
  children: (data: T) => ReactNode;
  loadingLabel?: string;
}) {
  if (q.data !== undefined) return <>{children(q.data)}</>;
  if (q.error) return <ErrorState error={q.error} onRetry={() => q.refetch()} />;
  return <Loading label={loadingLabel} />;
}

const styles = StyleSheet.create({
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 24, minHeight: 160 },
});
