import type { ReviewItem } from '../api/review';

/* Parked payloads are free-form JSON written by the backend reviewers;
   each card reads the keys its kind needs. Parsing is defensive on
   purpose: a payload we cannot read must not break the queue. */

export function payloadOf(item: ReviewItem): any {
  try {
    return JSON.parse(item.PayloadJson || '{}');
  } catch {
    return {};
  }
}
