import { getCashFlow, getExpenseDistribution, getIncomeDistribution } from '../api/dashboard';
import { getInstitutions } from '../api/institutions';
import { getRecentDocuments } from '../api/documents';
import { getRecentMovements } from '../api/movements';
import type { Institution } from '../types/institution';
import type { Movement } from '../types/movement';
import type { Document } from '../types/document';
import type { UseQueryState } from '../types/api';
import { useListQuery } from './useListQuery';

/* Dashboard/institutions data hooks — all mount-only lists over the shared
   useListQuery. */

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
