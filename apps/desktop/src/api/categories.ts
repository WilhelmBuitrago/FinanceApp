import { apiGet } from './client';
import type { Category } from '../types/category';

export async function getCategories(): Promise<Category[]> {
  return apiGet<Category[]>('/categories');
}
