/* Dashboard aggregates — computed by the backend (S12 §34: the frontend
   only transforms API data for display, never aggregates financial rules).
   Contracts verified live. */

import { apiGet } from './client';

/* GET /dashboard/cash-flow — snake_case, one point per active movement date. */
export interface CashFlowPoint {
  date: string;
  inflow: number;
  outflow: number;
  net: number;
}

export async function getCashFlow(from?: string, to?: string): Promise<CashFlowPoint[]> {
  const params = new URLSearchParams();
  if (from) params.set('from', from);
  if (to) params.set('to', to);
  const query = params.toString();
  return apiGet<CashFlowPoint[]>(`/dashboard/cash-flow${query ? `?${query}` : ''}`);
}

/* GET /dashboard/{income,expense}-distribution — [{category, amount_minor}].
   Empty distributions arrive as JSON null (Go nil slice); normalized here. */
export interface DistributionSlice {
  category: string;
  amount_minor: number;
}

export async function getIncomeDistribution(): Promise<DistributionSlice[]> {
  const data = await apiGet<DistributionSlice[] | null>('/dashboard/income-distribution');
  return data ?? [];
}

export async function getExpenseDistribution(): Promise<DistributionSlice[]> {
  const data = await apiGet<DistributionSlice[] | null>('/dashboard/expense-distribution');
  return data ?? [];
}
