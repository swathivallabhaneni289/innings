import Svg, { Circle, ClipPath, Defs, G, Path, Rect } from 'react-native-svg';

import { useTheme } from '@/theme';

export type BallFill = 'full' | 'half' | 'empty';

/** A small leather cricket ball with its seam. `half` fills the left half; `empty` is an inked outline. */
export function CricketBall({ size = 20, fill = 'full', color }: { size?: number; fill?: BallFill; color?: string }) {
  const t = useTheme();
  const red = color ?? t.red;
  const r = 10;
  const clipId = `half-${size}`;
  const ball = (
    <G>
      <Circle cx={12} cy={12} r={r} fill={red} />
      {/* seam: two parallel arcs with stitches */}
      <Path d="M7.2 3.6 C11.6 8.4 11.6 15.6 7.2 20.4" stroke={t.seam} strokeWidth={0.9} fill="none" opacity={0.95} />
      <Path d="M9.4 2.6 C13.8 7.8 13.8 16.2 9.4 21.4" stroke={t.seam} strokeWidth={0.9} fill="none" opacity={0.95} />
      <Path
        d="M8 5.6 l1.2 -0.6 M9.4 8 l1.3 -0.4 M10.1 10.6 l1.3 -0.2 M10.2 13.4 l1.3 0.2 M9.6 16 l1.3 0.4 M8.4 18.6 l1.2 0.6"
        stroke={t.seam}
        strokeWidth={0.7}
        opacity={0.85}
      />
      <Circle cx={15.5} cy={7.5} r={3.2} fill="#fff" opacity={0.12} />
    </G>
  );
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24">
      {fill === 'empty' ? (
        <Circle cx={12} cy={12} r={r - 0.6} fill="none" stroke={t.inkFaint} strokeWidth={1.2} strokeDasharray="2.2 1.8" />
      ) : null}
      {fill === 'full' ? ball : null}
      {fill === 'half' ? (
        <>
          <Defs>
            <ClipPath id={clipId}>
              <Rect x={0} y={0} width={12} height={24} />
            </ClipPath>
          </Defs>
          <Circle cx={12} cy={12} r={r - 0.6} fill="none" stroke={t.inkFaint} strokeWidth={1.2} strokeDasharray="2.2 1.8" />
          <G clipPath={`url(#${clipId})`}>{ball}</G>
        </>
      ) : null}
    </Svg>
  );
}
