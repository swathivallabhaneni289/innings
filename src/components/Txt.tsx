import { StyleSheet, Text, type TextProps, type TextStyle } from 'react-native';

import { type, useTheme, type TypeVariant } from '@/theme';

export type Tone = 'ink' | 'soft' | 'faint' | 'red' | 'green' | 'onRed';

export interface TxtProps extends TextProps {
  variant?: TypeVariant;
  tone?: Tone;
  align?: TextStyle['textAlign'];
  upper?: boolean;
}

/**
 * Every piece of text in the app goes through this, so the type scale stays consistent.
 * A custom fontSize without a lineHeight keeps the variant's proportions, so big numbers never clip.
 */
export function Txt({ variant = 'body', tone = 'ink', align, upper, style, ...rest }: TxtProps) {
  const t = useTheme();
  const color = { ink: t.ink, soft: t.inkSoft, faint: t.inkFaint, red: t.red, green: t.green, onRed: t.onRed }[tone];
  const base = type[variant];
  const flat = StyleSheet.flatten(style) as TextStyle | undefined;
  const fitted =
    flat?.fontSize && flat.lineHeight === undefined ? { lineHeight: Math.ceil(flat.fontSize * Math.max(1.12, base.lineHeight / base.fontSize)) } : null;
  return (
    <Text
      {...rest}
      style={[base, { color }, align ? { textAlign: align } : null, upper ? { textTransform: 'uppercase' } : null, style, fitted]}
    />
  );
}
