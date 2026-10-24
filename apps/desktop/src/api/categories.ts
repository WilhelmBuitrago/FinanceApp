import { apiGet, apiPost, apiDelete } from './client';
import type { Category } from '../types/category';

export async function getCategories(): Promise<Category[]> {
  return apiGet<Category[]>('/categories');
}

// The id is generated on the client (same scheme as the accounts).
export async function createCategory(
  id: string,
  name: string,
  categoryType: string,
): Promise<Category> {
  return apiPost<Category>('/categories', { id, name, category_type: categoryType });
}

export async function deleteCategory(id: string): Promise<void> {
  await apiDelete(`/categories/${id}`);
}
