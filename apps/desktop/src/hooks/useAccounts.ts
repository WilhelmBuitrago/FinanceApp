import { useEffect, useState } from 'react';
import { getAccounts } from '../api/accounts';
import type { Account } from '../types/account';
import type { UseQueryState } from '../types/api';

export function useAccounts(refreshKey = 0): UseQueryState<Account[]> {
  const [state, setState] = useState<UseQueryState<Account[]>>({ loading: true });

  useEffect(() => {
    let cancelled = false;
    setState({ loading: true });
    getAccounts()
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
