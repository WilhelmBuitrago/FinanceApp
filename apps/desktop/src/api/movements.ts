import { apiGet, apiDelete, apiPut } from './client';
import type { Movement } from '../types/movement';

export async function getMovements(): Promise<Movement[]> {
  return apiGet<Movement[]>('/movements');
}

export async function getRecentMovements(): Promise<Movement[]> {
  return apiGet<Movement[]>('/movements/recent');
}

// Assigns a category to a movement, or clears it with null. The category
// lives on the financial event behind the movement; the backend moves it.
export async function setMovementCategory(id: string, categoryId: string | null): Promise<void> {
  await apiPut<void>(`/movements/${id}/category`, { categoryId });
}

export async function deleteMovement(id: string): Promise<void> {
  await apiDelete(`/movements/${id}`);
}
