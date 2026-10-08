import { apiDelete } from './client';

/* Admin operations — destructive, guarded by confirmation UI. */

export async function purgeAllData(): Promise<void> {
  return apiDelete('/admin/purge');
}
