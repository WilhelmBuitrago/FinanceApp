import { apiGet, apiPostForm } from './client';
import type { Document } from '../types/document';

export async function getDocuments(): Promise<Document[]> {
  return apiGet<Document[]>('/documents');
}

export async function getDocument(id: string): Promise<Document> {
  return apiGet<Document>(`/documents/${id}`);
}

export async function getRecentDocuments(): Promise<Document[]> {
  return apiGet<Document[]>('/documents/recent');
}

/* POST /documents/upload (multipart): account_id + file + optional entity.
   Entity is verified by the backend against the target account's
   institution — the frontend sends the selected entity-account's
   institution so the verification can pass honestly.
   Response contract (UploadResponse json tags): snake_case. */
export interface UploadDocumentResponse {
  document_id: string;
  job_id: string;
  processing_id?: string;
}

export async function uploadDocument(
  accountId: string,
  file: File,
  entity?: string,
): Promise<UploadDocumentResponse> {
  const form = new FormData();
  form.append('account_id', accountId);
  if (entity) form.append('entity', entity);
  form.append('file', file);
  return apiPostForm<UploadDocumentResponse>('/documents/upload', form);
}
