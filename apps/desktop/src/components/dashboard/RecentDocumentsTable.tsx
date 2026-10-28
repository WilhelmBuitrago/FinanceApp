import type { UseQueryState } from '../../types/api';
import type { Document } from '../../types/document';
import { formatDate } from '../../lib/formatDate';
import { formatPeriod } from '../../lib/formatPeriod';
import { EmptyState } from '../ui/EmptyState';
import { SkeletonRows } from '../ui/SkeletonRows';
import { SectionHeader } from '../SectionHeader';
import { StatusBadge } from '../ui/StatusBadge';
import { tableHead } from '../ui/table';

/* Recent documents (three, backend-limited): name, account, FINANCIAL
   period, state and upload date. */

export function RecentDocumentsTable({
  documents,
  accountNames,
}: {
  documents: UseQueryState<Document[]>;
  accountNames: Map<string, string>;
}) {
  return (
    <section aria-label="Documentos recientes" className="space-y-2">
      <SectionHeader
        title="Documentos recientes"
        action={{ to: '/documents', label: 'Ver todos los documentos' }}
      />
      {documents.loading ? (
        <SkeletonRows rows={3} />
      ) : (documents.data?.length ?? 0) === 0 ? (
        <EmptyState
          title="Sin documentos"
          description="Los documentos bancarios importados aparecerán aquí."
        />
      ) : (
        <div className="overflow-hidden rounded-card border border-line bg-surface">
          <table className="w-full text-body">
            <thead className="bg-raised">
              <tr>
                <th className={tableHead}>Nombre</th>
                <th className={tableHead}>Cuenta</th>
                <th className={tableHead}>Período</th>
                <th className={`${tableHead} w-28`}>Estado</th>
                <th className={`${tableHead} w-28`}>Subido</th>
              </tr>
            </thead>
            <tbody>
              {(documents.data ?? []).slice(0, 3).map((doc) => (
                <tr key={doc.ID} className="border-t border-line/60">
                  <td className="w-full max-w-0 truncate px-3 py-2 text-ink">
                    {doc.DocumentName ?? 'Documento'}
                  </td>
                  <td className="w-full max-w-0 truncate px-3 py-2 text-ink-2">
                    {accountNames.get(doc.AccountID ?? '') ?? '—'}
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 text-ink-2">
                    {formatPeriod(doc.PeriodStart, doc.PeriodEnd)}
                  </td>
                  <td className="px-3 py-2">
                    <StatusBadge status={doc.Status} />
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 text-meta text-ink-3">
                    {formatDate(doc.CreatedAt)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
