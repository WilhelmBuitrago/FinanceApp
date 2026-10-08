import { useEffect, useState } from 'react';
import { getDocuments } from '../api/documents';
import type { Document } from '../types/document';
import type { UseQueryState } from '../types/api';

export function useDocuments(refreshKey = 0): UseQueryState<Document[]> {
  const [state, setState] = useState<UseQueryState<Document[]>>({ loading: true });

  useEffect(() => {
    let cancelled = false;
    setState({ loading: true });
    getDocuments()
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
