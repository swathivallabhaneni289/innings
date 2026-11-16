import { router } from 'expo-router';
import { useEffect, useState } from 'react';

import { WelcomeTicket } from '@/features/auth/WelcomeTicket';
import { useAuthFlow } from '@/features/auth/flow';
import { entranceLead } from '@/lib/splash';

/** The front gate: the whole screen is one match ticket (see WelcomeTicket), with the way in on its stub. */
export default function WelcomeScreen() {
  const clear = useAuthFlow((s) => s.clear);
  // On a cold launch the ticket waits for the splash to fade; after a sign-out it starts straight away.
  const [lead] = useState(entranceLead);
  // A fresh visit to the gate (first launch, or after signing out) starts from scratch.
  useEffect(() => clear(), [clear]);

  return <WelcomeTicket lead={lead} onStart={() => router.push('/sign-up')} onSignIn={() => router.push('/sign-in')} />;
}
