import { QueryClient } from '@tanstack/react-query';

import { ApiError } from './client';

/** The app's one react-query cache. A module singleton so signing out can clear it from outside React. */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 10_000,
      // Retry once, except when the session is gone: a 401 won't get better by asking again.
      retry: (failures, e) => failures < 1 && !(e instanceof ApiError && e.status === 401),
    },
  },
});
