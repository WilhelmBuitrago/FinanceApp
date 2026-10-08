import { apiGet } from './client';
import type { FinancialEvent } from '../types/financial-event';

export async function getFinancialEvent(id: string): Promise<FinancialEvent> {
  return apiGet<FinancialEvent>(`/financial-events/${id}`);
}
