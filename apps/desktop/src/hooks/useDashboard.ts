import { useEffect, useState } from 'react';

import { getCashFlow, getExpenseDistribution, getIncomeDistribution } from '../api/dashboard';
import { getInstitutions } from '../api/institutions';
import { getRecentDocuments } from '../api/documents';
import { getRecentMovements } from '../api/movements';
import type { Institution } from '../types/institution';
import type { Movement } from '../types/movement';
import type { Document } from '../types/document';
import type { UseQueryState } from '../types/api';

/* Dashboard/institutions data hooks — same contract as the list hooks. */

function useListQuery<T>(fetcher: () => Promise<T>): UseQueryState<T> {
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
  }, []);

  return state;
}

export function useInstitutions(): UseQueryState<Institution[]> {
  return useListQuery<Institution[]>(getInstitutions);
}

export function useCashFlow() {
  return useListQuery(getCashFlow);
}

export function useIncomeDistribution() {
  return useListQuery(getIncomeDistribution);
}

export function useExpenseDistribution() {
  return useListQuery(getExpenseDistribution);
}

export function useRecentMovements(): UseQueryState<Movement[]> {
  return useListQuery<Movement[]>(getRecentMovements);
}

export function useRecentDocuments(): UseQueryState<Document[]> {
  return useListQuery<Document[]>(getRecentDocuments);
}
