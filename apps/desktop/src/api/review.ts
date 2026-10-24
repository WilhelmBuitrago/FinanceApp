import { apiPost, apiGet } from './client';
import type { Movement } from '../types/movement';

/* Human review queue: where untrusted or ambiguous data lands until a
   human decides. One function per backend handler (handlers_review.go),
   with the exact request bodies the handlers expect. */

export interface ReviewItem {
  ID: string;
  Kind: 'DUPLICATE' | 'CATEGORY' | 'EXTRACTION' | 'DOCUMENT_PERIOD';
  PayloadJson?: string;
  SourceEntryID?: string;
  Status: string;
  CreatedAt: string;
}

export async function getReviewEntries(): Promise<ReviewItem[]> {
  return apiGet<ReviewItem[]>('/review/entries');
}

// Resolving an EXTRACTION/DUPLICATE item writes the corrected movement:
// the handler requires all of these fields (400 otherwise).
export async function resolveReviewMovement(
  id: string,
  fields: {
    description: string;
    amount_minor: number;
    date: string;
    type: string;
    account_id: string;
  },
): Promise<Movement> {
  return apiPost<Movement>(`/review/entries/${id}/resolve`, fields);
}

// Resolving a CATEGORY item only moves the category on the financial
// event. null leaves the movement uncategorized: a valid state, never a
// forced review item.
export async function resolveReviewCategory(id: string, categoryId: string | null): Promise<Movement> {
  return apiPost<Movement>(`/review/entries/${id}/resolve`, { category_id: categoryId });
}

// Document-level resolution: the human supplies the statement month and
// year; the backend derives the period and the parked movement dates.
export async function resolveReviewPeriod(id: string, month: number, year: number): Promise<void> {
  await apiPost<void>(`/review/entries/${id}/resolve-period`, { month, year });
}

// Discarding parks the item as discarded: nothing reaches the ledger.
export async function discardReviewEntry(id: string): Promise<void> {
  await apiPost<void>(`/review/entries/${id}/discard`, {});
}
