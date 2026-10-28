import { apiGet, apiDelete, apiPut, apiPost } from './client';
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

export interface CreateMovementInput {
  amount_minor: number;
  type: string; // INCOME | EXPENSE | TRANSFER
  account_id: string;
  description: string;
  date: string;
  category_id?: string;
}

// A colliding movement answers 409 with the parked review item: the error
// carries it (ApiError.response.review_item_id) so the UI can say where
// the duplicate went. A transfer answers 201 with both legs: the created
// movement and its counterpart (the other side of the same decision).
export async function createMovement(
  input: CreateMovementInput,
): Promise<{ id: string; counterpart_id?: string }> {
  return apiPost<{ id: string; counterpart_id?: string }>('/movements', input);
}

export async function deleteMovement(id: string): Promise<void> {
  await apiDelete(`/movements/${id}`);
}
