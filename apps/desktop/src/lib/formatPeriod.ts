/* Document period formatting — the FINANCIAL period of the statement
   (PeriodStart–PeriodEnd), distinct from the upload date (CreatedAt). */

import { formatDate } from './formatDate';

export function formatPeriod(start?: string | null, end?: string | null): string {
  if (start && end) return `${formatDate(start)} — ${formatDate(end)}`;
  if (start) return formatDate(start);
  if (end) return formatDate(end);
  return '—';
}
