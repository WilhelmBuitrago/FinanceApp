import { useCallback, useEffect, useState } from 'react';

import { ApiError } from '../api/client';
import { useCollectionVersion, type DataCollection } from '../lib/dataBus';

/* Per-section resource state. Each fetch resolves, fails or retries on its
   own — a failure in one section never blanks the others. Errors carry the
   HTTP status code so callers can distinguish 404 (honest absence) from
   transport failures. Pass collections to refetch when a mutation
   invalidates them through the shared bus. */

export type ResourceState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'error'; code?: number }
  | { status: 'ready'; data: T };

/* What consumers actually hold: the state plus its retry handle. */
export type ResourceHandle<T> = ResourceState<T> & { retry: () => void };

export function useResource<T>(
  id: string | null | undefined,
  fetcher: (id: string) => Promise<T>,
  collections: DataCollection[] = [],
): ResourceHandle<T> {
  const [state, setState] = useState<ResourceState<T>>({ status: 'idle' });
  const [attempt, setAttempt] = useState(0);
  const effectiveId = id ?? null;
  const version = useCollectionVersion(...collections);

  useEffect(() => {
    if (!effectiveId) {
      setState({ status: 'idle' });
      return;
    }
    let cancelled = false;
    setState({ status: 'loading' });
    fetcher(effectiveId)
      .then((data) => {
        if (!cancelled) setState({ status: 'ready', data });
      })
      .catch((err) => {
        if (cancelled) return;
        const code = err instanceof ApiError ? err.status : undefined;
        setState({ status: 'error', code });
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [effectiveId, attempt, version]);

  const retry = useCallback(() => setAttempt((a) => a + 1), []);

  if (state.status === 'ready') return { status: 'ready', data: state.data, retry };
  if (state.status === 'loading') return { status: 'loading', retry };
  if (state.status === 'error') return { status: 'error', code: state.code, retry };
  return { status: 'idle', retry };
}
