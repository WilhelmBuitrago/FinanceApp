import { apiGet } from './client';
import type { ImportJob } from '../types/import-job';

export async function getJob(id: string): Promise<ImportJob> {
  return apiGet<ImportJob>(`/jobs/${id}`);
}
