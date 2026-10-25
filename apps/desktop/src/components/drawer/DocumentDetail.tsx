import { Link } from 'react-router-dom';

import type { ResourceHandle } from '../../hooks/useResource';
import { formatDate } from '../../lib/formatDate';
import type { Document } from '../../types/document';
import { CopyId } from '../CopyId';
import { DrawerSectionLabel, KV, SectionError, SectionLoading } from '../DrawerShell';
import { StatusBadge } from '../ui/StatusBadge';

/* DOCUMENTO — appears progressively once the source entry resolves. The
   period is the FINANCIAL period, not the upload date. */

export function DocumentDetail({ document: doc }: { document: ResourceHandle<Document> }) {
  return (
    <section className="space-y-1.5">
      <DrawerSectionLabel>Documento</DrawerSectionLabel>
      {doc.status === 'loading' && <SectionLoading />}
      {doc.status === 'error' && <SectionError onRetry={doc.retry} />}
      {doc.status === 'ready' && (
        <div className="space-y-2">
          <dl>
            <KV label="Nombre">{doc.data.DocumentName ?? 'Documento'}</KV>
            <KV label="Tipo">{doc.data.DocumentType}</KV>
            <KV label="Institución">{doc.data.Institution}</KV>
            <KV label="Período">
              {doc.data.PeriodStart && doc.data.PeriodEnd
                ? `${formatDate(doc.data.PeriodStart)} — ${formatDate(doc.data.PeriodEnd)}`
                : '—'}
            </KV>
            <KV label="Estado">
              <StatusBadge status={doc.data.Status} />
            </KV>
            <KV label="ID">
              <CopyId id={doc.data.ID} />
            </KV>
          </dl>
          <Link
            to="/documents"
            className="inline-flex text-meta font-medium text-accent hover:underline"
          >
            Ver en Documentos
          </Link>
        </div>
      )}
    </section>
  );
}
