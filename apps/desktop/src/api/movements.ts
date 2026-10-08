import { apiGet } from './client';
import type { Movement } from '../types/movement';

export async function getMovements(): Promise<Movement[]> {
  return apiGet<Movement[]>('/movements');
}

export async function getRecentMovements(): Promise<Movement[]> {
  return apiGet<Movement[]>('/movements/recent');
}
