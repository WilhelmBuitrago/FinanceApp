import { getDocuments } from '../api/documents';
import type { Document } from '../types/document';
import { useListQuery } from './useListQuery';

export function useDocuments(refreshKey = 0) {
  return useListQuery<Document[]>(getDocuments, refreshKey);
}
