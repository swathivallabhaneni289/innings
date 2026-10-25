// On-device preferences only. Everything else lives in the Go API; the session token lives in SecureStore
// (src/data/session.ts).
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useEffect, useState } from 'react';
import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

import type { ThemePreference } from '@/domain/types';

interface PrefsState {
  theme: ThemePreference;
  notificationsSeenAt: string | null; // ISO; the bell counts items newer than this
  setTheme: (theme: ThemePreference) => void;
  markNotificationsSeen: () => void;
}

export const usePrefs = create<PrefsState>()(
  persist(
    (set) => ({
      theme: 'auto',
      notificationsSeenAt: null,
      setTheme: (theme) => set({ theme }),
      markNotificationsSeen: () => set({ notificationsSeenAt: new Date().toISOString() }),
    }),
    {
      name: 'innings-prefs-v1',
      storage: createJSONStorage(() => AsyncStorage),
      // v0 also kept the old guest token here. It is dropped: sign-in replaced the guest account.
      version: 1,
      migrate: (persisted) => {
        const old = (persisted ?? {}) as Partial<PrefsState>;
        return { theme: old.theme ?? 'auto', notificationsSeenAt: old.notificationsSeenAt ?? null } as PrefsState;
      },
      partialize: ({ theme, notificationsSeenAt }) => ({ theme, notificationsSeenAt }),
    },
  ),
);

/** True once AsyncStorage has been read. */
export function usePrefsHydrated() {
  const [hydrated, setHydrated] = useState(usePrefs.persist.hasHydrated());
  useEffect(() => usePrefs.persist.onFinishHydration(() => setHydrated(true)), []);
  return hydrated;
}
