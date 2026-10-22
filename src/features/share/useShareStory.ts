import * as Sharing from 'expo-sharing';
import { useCallback, useRef, useState } from 'react';
import type { View } from 'react-native';
import { captureRef } from 'react-native-view-shot';

/**
 * Captures the off-screen <StoryCard ref={ref}> as a PNG and opens the share sheet ("Share to your story").
 * Mount the card first (it needs a layout pass), then call `share()`.
 */
export function useShareStory() {
  const ref = useRef<View>(null);
  const [sharing, setSharing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const share = useCallback(async () => {
    setError(null);
    if (!ref.current) {
      setError('The story card is still printing. Try again in a second.');
      return;
    }
    setSharing(true);
    try {
      if (!(await Sharing.isAvailableAsync())) {
        setError("Sharing isn't available on this device.");
        return;
      }
      const uri = await captureRef(ref, { format: 'png', quality: 1, result: 'tmpfile' });
      await Sharing.shareAsync(uri, { mimeType: 'image/png', UTI: 'public.png', dialogTitle: 'Share to your story' });
    } catch (e) {
      setError(e instanceof Error && e.message ? `Couldn't share the card: ${e.message}` : "Couldn't share the card.");
    } finally {
      setSharing(false);
    }
  }, []);

  return { ref, share, sharing, error };
}
