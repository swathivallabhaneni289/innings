import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useColorScheme } from 'react-native';

import { usePrefs } from '@/data/store';
import { floodlights, paper, type Palette } from './tokens';

export * from './tokens';

const ThemeContext = createContext<Palette>(paper);

/** Paper by day; "floodlights" when the system is dark or the fan forces it in Settings. */
export function ThemeProvider({ children }: { children: ReactNode }) {
  const pref = usePrefs((s) => s.theme);
  const scheme = useColorScheme();
  const palette = useMemo(() => {
    if (pref === 'floodlights') return floodlights;
    if (pref === 'paper') return paper;
    return scheme === 'dark' ? floodlights : paper;
  }, [pref, scheme]);
  return <ThemeContext.Provider value={palette}>{children}</ThemeContext.Provider>;
}

export const useTheme = () => useContext(ThemeContext);

/** Renders children in a fixed theme, e.g. the shareable story card, which always prints in daylight. */
export function ForceTheme({ mode, children }: { mode: 'paper' | 'floodlights'; children: ReactNode }) {
  return <ThemeContext.Provider value={mode === 'paper' ? paper : floodlights}>{children}</ThemeContext.Provider>;
}
