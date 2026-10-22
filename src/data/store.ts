// On-device preferences only. Everything else lives in the Go API.
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useEffect, useState } from 'react';
import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

import type { ThemePreference } from '@/domain/types';

interface PrefsState {
  token: string | null; // guest session token from POST /v1/bootstrap
  theme: ThemePreference;
  notificationsSeenAt: string | null; // ISO; the bell counts items newer than this
  setToken: (token: string | null) => void;
  setTheme: (theme: ThemePreference) => void;
  markNotificationsSeen: () => void;
}

export const usePrefs = create<PrefsState>()(
  persist(
    (set) => ({
      token: null,
      theme: 'auto',
      notificationsSeenAt: null,
      setToken: (token) => set({ token }),
      setTheme: (theme) => set({ theme }),
      markNotificationsSeen: () => set({ notificationsSeenAt: new Date().toISOString() }),
    }),
    {
      name: 'innings-prefs-v1',
      storage: createJSONStorage(() => AsyncStorage),
      partialize: ({ token, theme, notificationsSeenAt }) => ({ token, theme, notificationsSeenAt }),
    },
  ),
);

/** True once AsyncStorage has been read. */
export function usePrefsHydrated() {
  const [hydrated, setHydrated] = useState(usePrefs.persist.hasHydrated());
  useEffect(() => usePrefs.persist.onFinishHydration(() => setHydrated(true)), []);
  return hydrated;
}
