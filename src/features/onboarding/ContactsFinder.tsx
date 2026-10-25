import { Feather } from '@expo/vector-icons';
import { useState, type ComponentType } from 'react';
import { ActivityIndicator, Linking, Share, StyleSheet, View } from 'react-native';

import { Button, Panel, Rule, Txt } from '@/components';
import { useAccount } from '@/data';
import type { User } from '@/domain/types';
import { DEFAULT_COUNTRY } from '@/features/auth/countries';
import { errorText } from '@/features/auth/errors';
import { countryOf } from '@/lib/phone';
import { useTheme } from '@/theme';
import { findFriendsInContacts } from './contacts';

const INVITE =
  "I'm keeping my matchday diary on Innings: call it before the first ball, rate it after the last, keep the stub. Come and keep score with me.";

type Row = ComponentType<{ user: User; following: boolean; last?: boolean }>;

type ContactsState =
  | { kind: 'idle' }
  | { kind: 'working' }
  | { kind: 'denied'; canAskAgain: boolean }
  | { kind: 'done'; matches: { user: User; following: boolean }[] }
  | { kind: 'error'; message: string };

/**
 * "From your contacts": asks for the address book, lists the fans already on Innings and offers an invite. Used
 * in sign-up (rows that can't open a profile yet) and on the Friends screen (rows that can). `onMatched` hands the
 * matched ids back so a list below can leave them out.
 */
export function ContactsFinder({ Row, onMatched }: { Row: Row; onMatched?: (ids: Set<string>) => void }) {
  const t = useTheme();
  const account = useAccount();
  const [contacts, setContacts] = useState<ContactsState>({ kind: 'idle' });

  const search = async () => {
    setContacts({ kind: 'working' });
    try {
      const phone = account.data?.phone;
      const res = await findFriendsInContacts(countryOf(phone) ?? DEFAULT_COUNTRY, phone);
      if (res.kind === 'denied') {
        setContacts(res);
        return;
      }
      setContacts({ kind: 'done', matches: res.matches });
      onMatched?.(new Set(res.matches.map((m) => m.user.id)));
    } catch (e) {
      setContacts({ kind: 'error', message: errorText(e, "Couldn't read your contacts. Try again.") });
    }
  };

  const invite = () => {
    Share.share({ message: INVITE }).catch(() => {});
  };

  return (
    <Panel>
      {contacts.kind === 'idle' ? (
        <View>
          <Txt variant="body">See who&apos;s already here from your phone&apos;s contacts.</Txt>
          <Txt variant="monoSmall" tone="soft" style={{ marginTop: 6 }}>
            Numbers are matched as hashes and never stored.
          </Txt>
          <Button
            label="Find friends from contacts"
            kind="outline"
            icon={<Feather name="book" size={14} color={t.ink} />}
            onPress={search}
            style={styles.action}
          />
        </View>
      ) : contacts.kind === 'working' ? (
        <View style={styles.working}>
          <ActivityIndicator size="small" color={t.inkSoft} />
          <Txt variant="bodySmall" tone="soft" style={{ marginLeft: 10 }}>
            Checking the turnstiles…
          </Txt>
        </View>
      ) : contacts.kind === 'denied' ? (
        <View>
          <Txt variant="body">Innings can&apos;t see your contacts.</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
            Allow access in Settings and come back, or skip it: you can find friends later from the Friends screen.
          </Txt>
          <View style={styles.buttons}>
            <Button label="Open Settings" kind="outline" size="sm" onPress={() => Linking.openSettings().catch(() => {})} style={styles.tap} />
            {contacts.canAskAgain ? <Button label="Ask again" kind="ghost" size="sm" onPress={search} style={styles.tap} /> : null}
          </View>
        </View>
      ) : contacts.kind === 'error' ? (
        <View>
          <Txt variant="bodySmall" tone="red">
            {contacts.message}
          </Txt>
          <Button label="Try again" kind="outline" size="sm" onPress={search} style={styles.tapStart} />
        </View>
      ) : contacts.matches.length === 0 ? (
        <View>
          <Txt variant="h3">None of your contacts are on Innings yet.</Txt>
          <Txt variant="bodySmall" tone="soft" style={{ marginTop: 4 }}>
            Every good stand starts with one fan. Send them a ticket.
          </Txt>
          <Button
            label="Invite friends"
            kind="outline"
            size="sm"
            icon={<Feather name="share" size={13} color={t.ink} />}
            onPress={invite}
            style={styles.tapStart}
          />
        </View>
      ) : (
        <View>
          <Txt variant="h3">
            {contacts.matches.length === 1 ? '1 of your contacts is here' : `${contacts.matches.length} of your contacts are here`}
          </Txt>
          {contacts.matches.map((m, i) => (
            <Row key={m.user.id} user={m.user} following={m.following} last={i === contacts.matches.length - 1} />
          ))}
          <Rule />
          <Button label="Invite more friends" kind="ghost" size="sm" onPress={invite} style={{ ...styles.tapStart, marginTop: 4 }} />
        </View>
      )}
    </Panel>
  );
}

const styles = StyleSheet.create({
  action: { marginTop: 14, minHeight: 48 },
  working: { flexDirection: 'row', alignItems: 'center', minHeight: 48 },
  buttons: { flexDirection: 'row', alignItems: 'center', gap: 12, marginTop: 12 },
  tap: { minHeight: 44 },
  tapStart: { minHeight: 44, marginTop: 12, alignSelf: 'flex-start' },
});
