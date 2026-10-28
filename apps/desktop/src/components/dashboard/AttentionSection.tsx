import type { Document } from '../../types/document';
import { formatPeriod } from '../../lib/formatPeriod';
import { SectionHeader } from '../SectionHeader';
import { StatusBadge } from '../ui/StatusBadge';

/* Requires attention — documents whose import parked something
   (review required or failed). The section only exists when the list
   is not empty; three rows, the rest live in Documents. */

export function AttentionSection({
  documents,
  accountNames,
}: {
  documents: Document[];
  accountNames: Map<string, string>;
}) {
  return (
    <section aria-label="Requiere atención" className="space-y-2">
      <SectionHeader
        title="Requiere atención"
        action={{ to: '/documents', label: 'Ver documentos' }}
      />
      <div className="overflow-hidden rounded-card border border-line bg-surface">
        {documents.slice(0, 3).map((doc) => (
          <div
            key={doc.ID}
            className="flex items-center justify-between gap-3 border-t border-line/60 px-3 py-2 first:border-t-0"
          >
            <div className="min-w-0">
              <p className="truncate text-ink">{doc.DocumentName ?? 'Documento'}</p>
              <p className="truncate text-meta text-ink-3">
                {[
                  accountNames.get(doc.AccountID ?? ''),
                  formatPeriod(doc.PeriodStart, doc.PeriodEnd),
                ]
                  .filter(Boolean)
                  .join(' · ')}
              </p>
            </div>
            <StatusBadge status={doc.Status} />
          </div>
        ))}
      </div>
    </section>
  );
}
