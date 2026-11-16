import { Redirect, router } from 'expo-router';

import { checkResetCode, forgotPassword } from '@/data';
import { CodeStep } from '@/features/auth/CodeStep';
import { useAuthFlow } from '@/features/auth/flow';

/** The emailed reset code. The right one leads on to choosing a new password. */
export default function ResetCodeScreen() {
  const sent = useAuthFlow((s) => s.sent);
  const setSent = useAuthFlow((s) => s.setSent);
  const setReset = useAuthFlow((s) => s.setReset);
  if (!sent || sent.purpose !== 'reset') return <Redirect href="/forgot" />;

  return (
    <CodeStep
      title="Enter the code"
      lead="If it has an account, we've sent a 6-digit code to"
      sent={sent}
      onCheck={async (code) => {
        const res = await checkResetCode(sent.email, code);
        setReset(res.resetToken, res.expiresInSeconds);
        router.replace('/new-password');
      }}
      onResend={async () => {
        const res = await forgotPassword(sent.email);
        setSent(res, 'reset');
        return res;
      }}
      onChangeEmail={() => (router.canGoBack() ? router.back() : router.replace('/forgot'))}
    />
  );
}
