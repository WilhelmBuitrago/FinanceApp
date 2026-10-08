import { useEffect, useState } from 'react';
import { getMovements } from '../api/movements';
import type { Movement } from '../types/movement';
import type { UseQueryState } from '../types/api';

export function useMovements(refreshKey = 0): UseQueryState<Movement[]> {
  const [state, setState] = useState<UseQueryState<Movement[]>>({ loading: true });

  useEffect(() => {
    let cancelled = false;
    setState({ loading: true });
    getMovements()
      .then((data) => {
        if (cancelled) return;
        setState({ loading: false, data, empty: data.length === 0 });
      })
      .catch((err) => {
        if (cancelled) return;
        setState({ loading: false, error: err instanceof Error ? err : new Error(String(err)) });
      });
    return () => {
      cancelled = true;
    };
  }, [refreshKey]);

  return state;
}
