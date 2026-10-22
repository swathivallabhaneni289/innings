// Dev-only QA driver: when Metro is started with EXPO_PUBLIC_QA_DRIVER=<url>, the app polls that URL for a
// route (plain text, e.g. "/match/ind-aus-t20i-3") and navigates to it. Used to screenshot every screen from the
// command line without tapping the simulator. Inert in normal runs.
import { router } from 'expo-router';
import { useEffect } from 'react';
import { LogBox } from 'react-native';

import { usePrefs } from '@/data/store';

const DRIVER = process.env.EXPO_PUBLIC_QA_DRIVER;

// Screenshots should show the app, not dev-only warning toasts.
if (__DEV__ && DRIVER) LogBox.ignoreAllLogs();

export function useQaDriver() {
  useEffect(() => {
    if (!__DEV__ || !DRIVER) return;
    let last = '';
    const id = setInterval(async () => {
      try {
        const res = await fetch(DRIVER, { cache: 'no-store' });
        const target = (await res.text()).trim();
        if (!target || target === last) return;
        last = target;
        if (target.startsWith('theme:')) {
          usePrefs.getState().setTheme(target.slice(6) as 'auto' | 'paper' | 'floodlights');
        } else if (target.startsWith('/(tabs)') || target === '/') {
          if (router.canDismiss()) router.dismissAll();
          router.replace(target as never);
        } else {
          router.push(target as never);
        }
      } catch {
        // driver not running; ignore
      }
    }, 1000);
    return () => clearInterval(id);
  }, []);
}
