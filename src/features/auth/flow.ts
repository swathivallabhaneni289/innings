// The sign-up or sign-in in progress, shared by the screens in (auth). In memory only: an app restart starts again,
// which is what you'd want for a 10-minute code anyway.
import { create } from 'zustand';

import type { AuthCodeResponse } from '@/data';

/** What the last emailed code was for: confirming a new account, or resetting a password. */
export type CodePurpose = 'confirm' | 'reset';

export type SentCode = AuthCodeResponse & { purpose: CodePurpose; sentAt: number };

interface AuthFlow {
  /** The address typed last, carried between sign-in, sign-up and forgot password so nobody types it twice. */
  email: string;
  /** The last code emailed: where, what for, when, and (dev builds only) the code itself. */
  sent: SentCode | null;
  /** From the right reset code: lets the new-password screen set one, until it expires. */
  reset: { token: string; expiresAt: number } | null;
  setEmail: (email: string) => void;
  setSent: (res: AuthCodeResponse, purpose: CodePurpose) => void;
  setReset: (token: string, expiresInSeconds: number) => void;
  clear: () => void;
}

export const useAuthFlow = create<AuthFlow>((set, get) => ({
  email: '',
  sent: null,
  reset: null,
  setEmail: (email) => set({ email }),
  setSent: (res, purpose) => {
    const prev = get().sent;
    // Asking again inside the resend gap sends nothing new: the earlier code (and, in dev, its devCode) still works.
    const devCode = res.devCode ?? (prev?.email === res.email && prev.purpose === purpose ? prev.devCode : undefined);
    set({ email: res.email, sent: { ...res, devCode, purpose, sentAt: Date.now() } });
  },
  setReset: (token, expiresInSeconds) => set({ reset: { token, expiresAt: Date.now() + expiresInSeconds * 1000 } }),
  clear: () => set({ email: '', sent: null, reset: null }),
}));

/** A code for this address and purpose that's still within its 10 minutes, so there's no need to ask for another. */
export function liveCode(sent: SentCode | null, email: string, purpose: CodePurpose): boolean {
  return !!sent && sent.email === email && sent.purpose === purpose && Date.now() < sent.sentAt + sent.expiresInSeconds * 1000;
}
