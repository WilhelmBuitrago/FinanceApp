import { apiGet } from './client';
import type { Institution } from '../types/institution';

export async function getInstitutions(): Promise<Institution[]> {
  return apiGet<Institution[]>('/institutions');
}
