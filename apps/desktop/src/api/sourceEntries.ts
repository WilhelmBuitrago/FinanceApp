import { apiGet } from './client';
import type { SourceEntry } from '../types/source-entry';

export async function getSourceEntry(id: string): Promise<SourceEntry> {
  return apiGet<SourceEntry>(`/source-entries/${id}`);
}
