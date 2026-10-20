import { apiGet, apiDelete } from './client';
import type { Movement } from '../types/movement';

export async function getMovements(): Promise<Movement[]> {
  return apiGet<Movement[]>('/movements');
}

export async function getRecentMovements(): Promise<Movement[]> {
  return apiGet<Movement[]>('/movements/recent');
}

export async function deleteMovement(id: string): Promise<void> {
  await apiDelete(`/movements/${id}`);
}
