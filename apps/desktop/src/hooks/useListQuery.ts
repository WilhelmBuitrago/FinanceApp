import { useEffect, useState } from 'react';

import type { UseQueryState } from '../types/api';

/* The one list-fetching hook in the app. Each fetch resolves loading,
   data+empty or error on its own; a failure in one section never blanks
   the others. Pass a refreshKey to re-run on demand (pages bump it after
   a mutation); without it the fetch runs on mount.

   It replaced four copies of the same effect (useAccounts, useMovements,
   useDocuments and a private twin inside useDashboard). */

export function useListQuery<T>(fetcher: () => Promise<T>, refreshKey = 0): UseQueryState<T> {
  const [state, setState] = useState<UseQueryState<T>>({ loading: true });

  useEffect(() => {
    let cancelled = false;
    setState({ loading: true });
    fetcher()
      .then((data) => {
        if (cancelled) return;
        setState({ loading: false, data, empty: Array.isArray(data) ? data.length === 0 : false });
      })
      .catch((err) => {
        if (cancelled) return;
        setState({ loading: false, error: err instanceof Error ? err : new Error(String(err)) });
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshKey]);

  return state;
}
