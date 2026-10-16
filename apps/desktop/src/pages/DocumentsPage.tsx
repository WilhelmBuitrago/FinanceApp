import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAccounts } from '../hooks/useAccounts';
import { useDocuments } from '../hooks/useDocuments';
import { formatDate } from '../lib/formatDate';
import { formatPeriod } from '../lib/formatPeriod';
import { DocumentDrawer } from '../components/DocumentDrawer';
import { ImportDocumentDrawer } from '../components/ImportDocumentDrawer';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { SkeletonRows } from '../components/ui/SkeletonRows';
import { StatusBadge } from '../components/ui/StatusBadge';
import type { Document } from '../types/document';

/* Documentos — provenance. An import ALWAYS requires a configured account
   (no orphan documents). Columns show the account, the FINANCIAL period
   (distinct from the upload date) and the state. No "Tipo" column: every
   supported document is a PDF, the format carries no information. */

const tableHead = 'px-3 py-1.5 text-left font-medium uppercase tracking-[0.08em] text-meta text-ink-3';

const primaryButton =
  'inline-flex h-8 items-center rounded-[8px] bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40 disabled:hover:bg-accent-strong';

export default function DocumentsPage() {
  const [refreshKey, setRefreshKey] = useState(0);
  const accounts = useAccounts(refreshKey);
  const documents = useDocuments(refreshKey);
  const [importOpen, setImportOpen] = useState(false);
  const [selected, setSelected] = useState<Document | null>(null);

  const accountNames = useMemo(() => {
    const map = new Map<string, string>();
    for (const account of accounts.data ?? []) map.set(account.ID, account.Name);
    return map;
  }, [accounts.data]);

  const noAccounts = !accounts.loading && !accounts.error && (accounts.data?.length ?? 0) === 0;
  const hasAccounts = !accounts.loading && !accounts.error && (accounts.data?.length ?? 0) > 0;

  return (
    <div
      className={`transition-[padding] duration-[220ms] ease-out-tk ${
        selected ? 'pr-[calc(min(440px,40vw)_+_16px)]' : ''
      }`}
    >
      {documents.loading ? (
        <SkeletonRows rows={5} />
      ) : documents.error ? (
        <ErrorState
          message="No se pudo cargar los documentos desde el servidor."
          onRetry={() => setRefreshKey((k) => k + 1)}
        />
      ) : noAccounts ? (
        /* The flow starts at configuration — no table, no upload button. */
        <EmptyState
          title="No tienes cuentas configuradas"
          description="Configura una cuenta antes de importar documentos financieros."
          action={
            <Link to="/accounts" className={primaryButton}>
              Configurar cuenta
            </Link>
          }
        />
      ) : (
        <>
          <div className="mb-6 flex items-baseline justify-between gap-4">
            <div>
              <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Documentos</h1>
              <p className="text-body text-ink-2">{documents.data?.length ?? 0} documentos registrados.</p>
            </div>
            <button onClick={() => setImportOpen(true)} className={primaryButton}>
              Importar documento
            </button>
          </div>

          {(documents.data?.length ?? 0) === 0 ? (
            <EmptyState
              title="Sin documentos"
              description="Importa el primer extracto de una de tus cuentas para generar movimientos."
            />
          ) : (
            <div className="overflow-x-auto rounded-[12px] border border-line bg-surface">
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
                  {(documents.data ?? []).map((doc) => {
                    const isSelected = selected?.ID === doc.ID;
                    return (
                      <tr
                        key={doc.ID}
                        tabIndex={0}
                        onClick={() => setSelected(doc)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') setSelected(doc);
                        }}
                        aria-label={`${doc.DocumentName ?? 'Documento'} — ${
                          accountNames.get(doc.AccountID ?? '') ?? 'sin cuenta'
                        }`}
                        className={`relative cursor-pointer border-t border-line/60 transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent ${
                          isSelected ? 'bg-accent-wash' : ''
                        }`}
                      >
                        {isSelected && (
                          <span className="absolute bottom-0 left-0 top-0 w-[2px] bg-accent" aria-hidden="true" />
                        )}
                        <td className="w-full max-w-0 truncate px-3 py-2 text-ink" title={doc.DocumentName ?? 'Documento'}>
                          {doc.DocumentName ?? 'Documento'}
                        </td>
                        <td className="w-full max-w-0 truncate px-3 py-2 text-ink-2" title={accountNames.get(doc.AccountID ?? '') ?? undefined}>
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
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {importOpen && hasAccounts && (
        <ImportDocumentDrawer
          accounts={accounts.data ?? []}
          onClose={() => setImportOpen(false)}
          onImported={() => {
            setImportOpen(false);
            setRefreshKey((k) => k + 1);
          }}
        />
      )}

      <DocumentDrawer
        document={selected}
        accountName={selected ? accountNames.get(selected.AccountID ?? '') ?? null : null}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
