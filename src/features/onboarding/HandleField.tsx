import { Feather } from '@expo/vector-icons';
import { useEffect, useState, type Ref } from 'react';
import { ActivityIndicator, StyleSheet, TextInput, View, type ViewStyle } from 'react-native';

import { Txt } from '@/components';
import { checkHandle } from '@/data';
import { fonts, useTheme } from '@/theme';

export const HANDLE_MIN = 3;
export const HANDLE_MAX = 20;

export type HandleState =
  | { kind: 'empty' }
  | { kind: 'same' } // the fan's current handle, unchanged
  | { kind: 'short' } // still typing: under 3 characters
  | { kind: 'invalid'; message: string } // breaks a rule we can check here
  | { kind: 'checking' }
  | { kind: 'available'; handle: string }
  | { kind: 'unavailable'; message: string }
  | { kind: 'error'; message: string }; // couldn't ask the server

/** What the field holds: no "@", no spaces, lower-case as it's typed. */
export const cleanHandle = (s: string) => s.replace(/^@+/, '').replace(/\s+/g, '').toLowerCase();

/** The rules from docs/API.md that can be checked without asking: the server has the final word. */
function localProblem(h: string): string | null {
  if (/[^a-z0-9_.]/.test(h)) return 'Letters, numbers, _ and . only.';
  if (!/^[a-z]/.test(h)) return 'Start with a letter.';
  if (h.endsWith('.')) return "Can't end with a full stop.";
  if (h.length > HANDLE_MAX) return `${HANDLE_MAX} characters at most.`;
  return null;
}

/**
 * Is this @username free? Checks the rules locally, then asks GET /v1/handles/{h} once typing pauses
 * (~400 ms). An answer for an older value is dropped, so the state always describes what's in the field.
 */
export function useHandleCheck(handle: string, current?: string): HandleState {
  const [answer, setAnswer] = useState<{ handle: string; state: HandleState } | null>(null);
  const same = !!current && handle === current;
  const problem = handle ? localProblem(handle) : null;
  const ask = !!handle && !same && !problem && handle.length >= HANDLE_MIN;

  useEffect(() => {
    if (!ask) return;
    let live = true;
    const timer = setTimeout(async () => {
      try {
        const r = await checkHandle(handle);
        if (!live) return;
        setAnswer({
          handle,
          state: r.available
            ? { kind: 'available', handle: r.handle }
            : { kind: 'unavailable', message: r.message ?? `@${handle} is already taken.` },
        });
      } catch {
        if (live) setAnswer({ handle, state: { kind: 'error', message: "Couldn't check that one just now. Try again in a moment." } });
      }
    }, 400);
    return () => {
      live = false;
      clearTimeout(timer);
    };
  }, [handle, ask]);

  if (!handle) return { kind: 'empty' };
  if (same) return { kind: 'same' };
  if (problem) return { kind: 'invalid', message: problem };
  if (handle.length < HANDLE_MIN) return { kind: 'short' };
  return answer?.handle === handle ? answer.state : { kind: 'checking' };
}

/** True when the handle in the field can be saved as it stands. */
export const handleOk = (s: HandleState) => s.kind === 'available' || s.kind === 'same';

/**
 * The @username field, shared by onboarding and Settings: an "@" that can't be deleted, lower-casing as you
 * type, and the availability answer underneath. `serverError` shows a 409/400 from saving.
 */
export function HandleField({
  value,
  onChangeText,
  state,
  serverError,
  style,
  inputRef,
}: {
  value: string;
  onChangeText: (handle: string) => void;
  state: HandleState;
  serverError?: string | null;
  style?: ViewStyle;
  inputRef?: Ref<TextInput>; // so the field before can hand focus on with "next"
}) {
  const t = useTheme();
  const bad = !!serverError || state.kind === 'invalid' || state.kind === 'unavailable';
  const line = serverError
    ? { tone: 'red' as const, text: serverError }
    : state.kind === 'available'
      ? { tone: 'green' as const, text: `@${state.handle} is yours` }
      : state.kind === 'invalid' || state.kind === 'unavailable'
        ? { tone: 'red' as const, text: state.message }
        : state.kind === 'error'
          ? { tone: 'soft' as const, text: state.message }
          : state.kind === 'checking'
            ? { tone: 'faint' as const, text: `Checking @${value}…` }
            : state.kind === 'same'
              ? { tone: 'faint' as const, text: 'Your username now.' }
              : { tone: 'faint' as const, text: `${HANDLE_MIN} to ${HANDLE_MAX} letters, numbers, _ or .` };

  return (
    <View style={[{ marginTop: 20 }, style]}>
      <View style={styles.head}>
        <Txt variant="bodySmall" tone="soft" style={{ fontFamily: fonts.label }}>
          Username
        </Txt>
        <Txt variant="monoSmall" tone={value.length >= HANDLE_MAX ? 'red' : 'faint'}>
          {value.length}/{HANDLE_MAX}
        </Txt>
      </View>
      <View style={[styles.row, { borderBottomColor: bad ? t.red : t.inkFaint }]}>
        <Txt style={[styles.at, { color: value ? t.ink : t.inkFaint }]}>@</Txt>
        <TextInput
          ref={inputRef}
          value={value}
          onChangeText={(s) => onChangeText(cleanHandle(s))}
          maxLength={HANDLE_MAX}
          placeholder="yourname"
          placeholderTextColor={t.inkFaint}
          selectionColor={t.red}
          autoCapitalize="none"
          autoCorrect={false}
          autoComplete="off"
          spellCheck={false}
          returnKeyType="done"
          accessibilityLabel="Username"
          style={[styles.input, { color: t.ink }]}
        />
        <View style={styles.mark}>
          {state.kind === 'checking' ? (
            <ActivityIndicator size="small" color={t.inkSoft} />
          ) : state.kind === 'available' && !serverError ? (
            <Feather name="check" size={18} color={t.green} />
          ) : bad ? (
            <Feather name="x" size={18} color={t.red} />
          ) : null}
        </View>
      </View>
      <Txt variant="bodySmall" tone={line.tone} accessibilityLiveRegion="polite" style={{ marginTop: 5 }}>
        {line.text}
      </Txt>
    </View>
  );
}

const styles = StyleSheet.create({
  head: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  row: { flexDirection: 'row', alignItems: 'center', borderBottomWidth: StyleSheet.hairlineWidth },
  at: { fontFamily: fonts.serif, fontSize: 17, lineHeight: 22, marginRight: 1 },
  input: { flex: 1, fontFamily: fonts.serif, fontSize: 17, paddingVertical: 10 },
  mark: { width: 24, alignItems: 'flex-end' },
});
