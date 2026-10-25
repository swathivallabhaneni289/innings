// The sign-in in progress, shared by the phone, country and code screens. In memory only: an app restart
// starts sign-in again, which is what you'd want for a 10-minute code anyway.
import type { CountryCode } from 'libphonenumber-js';
import { create } from 'zustand';

import type { AuthStartResponse } from '@/data';
import { DEFAULT_COUNTRY } from './countries';

interface SignInFlow {
  country: CountryCode;
  /** The last code sent: the number, when it was sent and (dev only) the code itself. */
  sent: (AuthStartResponse & { sentAt: number }) | null;
  setCountry: (c: CountryCode) => void;
  setSent: (res: AuthStartResponse) => void;
  reset: () => void;
}

export const useSignInFlow = create<SignInFlow>((set) => ({
  country: DEFAULT_COUNTRY,
  sent: null,
  setCountry: (country) => set({ country }),
  setSent: (res) => set({ sent: { ...res, sentAt: Date.now() } }),
  reset: () => set({ country: DEFAULT_COUNTRY, sent: null }),
}));
