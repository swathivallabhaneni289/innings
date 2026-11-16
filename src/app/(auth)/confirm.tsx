import { Redirect, router } from 'expo-router';

import { confirmEmail, resendConfirmation } from '@/data';
import { CodeStep } from '@/features/auth/CodeStep';
import { useAuthFlow } from '@/features/auth/flow';

/**
 * "Check your email": the code that confirms a new account's address. The right one signs the device in, and the
 * root guard opens the fan card (onboarding); an account that's already onboarded goes straight Home.
 */
export default function ConfirmScreen() {
  const sent = useAuthFlow((s) => s.sent);
  const setSent = useAuthFlow((s) => s.setSent);
  if (!sent || sent.purpose !== 'confirm') return <Redirect href="/sign-up" />;

  return (
    <CodeStep
      title="Check your email"
      lead="We sent a 6-digit code to"
      sent={sent}
      onCheck={async (code) => {
        const account = await confirmEmail(sent.email, code);
        if (account.onboarded) router.replace('/');
      }}
      onResend={async () => {
        const res = await resendConfirmation(sent.email);
        setSent(res, 'confirm');
        return res;
      }}
      onChangeEmail={() => (router.canGoBack() ? router.back() : router.replace('/sign-up'))}
    />
  );
}
