import { getCashFlow, getExpenseDistribution, getIncomeDistribution } from '../api/dashboard';
import { getInstitutions } from '../api/institutions';
import { getRecentDocuments } from '../api/documents';
import { getRecentMovements } from '../api/movements';
import type { Institution } from '../types/institution';
import type { Movement } from '../types/movement';
import type { Document } from '../types/document';
import type { UseQueryState } from '../types/api';
import { useListQuery } from './useListQuery';

/* Dashboard/institutions data hooks — same contract as the list hooks,
   each subscribed to the collections that make it stale. Institutions
   are a fixed catalog: they never subscribe (mount only). */

export function useInstitutions(): UseQueryState<Institution[]> {
  return useListQuery<Institution[]>(getInstitutions);
}

export function useCashFlow() {
  return useListQuery(getCashFlow, 0, ['dashboard']);
}

export function useIncomeDistribution() {
  return useListQuery(getIncomeDistribution, 0, ['dashboard']);
}

export function useExpenseDistribution() {
  return useListQuery(getExpenseDistribution, 0, ['dashboard']);
}

export function useRecentMovements(): UseQueryState<Movement[]> {
  return useListQuery<Movement[]>(getRecentMovements, 0, ['movements']);
}

export function useRecentDocuments(): UseQueryState<Document[]> {
  return useListQuery<Document[]>(getRecentDocuments, 0, ['documents']);
}
