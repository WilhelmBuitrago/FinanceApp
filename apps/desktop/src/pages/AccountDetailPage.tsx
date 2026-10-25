import { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getAccountDocuments, getAccountMovements } from '../api/accounts';
import { DetailDrawer } from '../components/DetailDrawer';
import { SectionHeader } from '../components/SectionHeader';
import { SummaryStrip } from '../components/SummaryStrip';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { SkeletonRows } from '../components/ui/SkeletonRows';
import { StatusBadge } from '../components/ui/StatusBadge';
import { tableHead, tableHeadRight } from '../components/ui/table';
import { useAccounts } from '../hooks/useAccounts';
import { useResource } from '../hooks/useResource';
import { accountTypeLabel } from '../lib/accountType';
import { invalidateCollections } from '../lib/dataBus';
import { formatDate } from '../lib/formatDate';
import { formatSignedMoney, signedAmountClass } from '../lib/formatMoney';
import type { Movement } from '../types/movement';

/* Cuenta individual — the account as a financial entity: identity, its real
   activity, its documents. Identity resolves against the complete account
   list (single-user desktop app): an account referenced by movements but
   not yet registered shows the honest "sin ficha" state — no invented
   entity, no 404 round-trip. Its movements and documents are real. */

export default function AccountDetailPage() {
  const { id = '' } = useParams();

  const accounts = useAccounts();
  const movements = useResource(id || null, getAccountMovements);
  const documents = useResource(id || null, getAccountDocuments);

  const account = useMemo(
    () => (accounts.data ?? []).find((a) => a.ID === id) ?? null,
    [accounts.data, id],
  );

  const [selected, setSelected] = useState<Movement | null>(null);

  /* Real all-time sums over the account's movements. */
  const sums = useMemo(() => {
    if (movements.status !== 'ready') return null;
    let income = 0;
    let expense = 0;
    for (const mv of movements.data) {
      if (mv.AmountMinor > 0) income += mv.AmountMinor;
      else if (mv.AmountMinor < 0) expense += mv.AmountMinor;
    }
    return { income, expense };
  }, [movements]);

  const movementList = movements.status === 'ready' ? movements.data : [];
  const documentList = documents.status === 'ready' ? documents.data : [];
  const accountName = account?.Name ?? null;

  return (
    <div
      className={`space-y-8 transition-[padding] duration-[220ms] ease-out-tk ${
        selected ? 'pr-[calc(min(440px,40vw)_+_16px)]' : ''
      }`}
    >
      {/* Identidad de la cuenta */}
      {accounts.loading ? (
        <div className="h-14 w-72 animate-pulse rounded-tk bg-interactive" aria-hidden="true" />
      ) : accounts.error ? (
        <ErrorState
          message="No se pudo cargar la cuenta desde el servidor."
          onRetry={() => invalidateCollections('accounts', 'movements', 'documents')}
        />
      ) : account ? (
        <header className="space-y-3">
          <h1 className="text-[1.75rem] font-semibold tracking-tight text-ink">{account.Name}</h1>
          <div className="flex flex-wrap items-center gap-2 text-body text-ink-2">
            <span>{accountTypeLabel(account.AccountType)}</span>
            <span aria-hidden="true">·</span>
            <span>{account.Institution ?? 'Sin institución'}</span>
            <span aria-hidden="true">·</span>
            <span className="uppercase">{account.Currency}</span>
            <span aria-hidden="true">·</span>
            <StatusBadge status={account.Status} />
            <span aria-hidden="true">·</span>
            <span className="text-meta text-ink-3">Creada {formatDate(account.CreatedAt)}</span>
          </div>
        </header>
      ) : (
        <header className="space-y-1.5">
          <p className="text-display font-semibold text-ink">Cuenta sin ficha</p>
          <p className="flex flex-wrap items-center gap-2 text-body text-ink-2">
            <span className="font-mono text-meta text-ink-3">{id}</span>
            <span aria-hidden="true">·</span>
            <span>
              Referenciada por la actividad importada, pero aún no existe como entidad registrada.
            </span>
          </p>
        </header>
      )}

      <div className="mt-6 space-y-8">
        {/* Resumen real del histórico de la cuenta */}
        {sums && movementList.length > 0 && (
          <SummaryStrip income={sums.income} expense={sums.expense} periodLabel="histórico" />
        )}

        {/* Movimientos de la cuenta */}
        <section aria-label="Movimientos de la cuenta" className="space-y-2">
          <SectionHeader
            title={`Movimientos${movements.status === 'ready' ? ` (${movementList.length})` : ''}`}
            action={{
              to: `/transactions?cuenta=${encodeURIComponent(id)}`,
              label: 'Ver en Movimientos',
            }}
          />
          {movements.status === 'loading' ? (
            <SkeletonRows rows={5} />
          ) : movements.status === 'error' ? (
            <ErrorState
              message="No se pudo cargar los movimientos de la cuenta."
              onRetry={movements.retry}
            />
          ) : movementList.length === 0 ? (
            <EmptyState
              title="Sin movimientos"
              description="Esta cuenta todavía no tiene movimientos registrados."
            />
          ) : (
            <div className="overflow-x-auto rounded-card border border-line bg-surface">
              <table
                className="w-full text-body"
                onKeyDown={(e) => {
                  if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
                  const rows = Array.from(e.currentTarget.querySelectorAll('tbody tr'));
                  const index = rows.indexOf(document.activeElement as HTMLElement);
                  const next = e.key === 'ArrowDown' ? index + 1 : index - 1;
                  if (next >= 0 && next < rows.length) {
                    e.preventDefault();
                    (rows[next] as HTMLElement).focus();
                  }
                }}
              >
                <thead className="bg-raised">
                  <tr>
                    <th className={`${tableHead} w-24`}>Fecha</th>
                    <th className={tableHead}>Descripción</th>
                    <th className={`${tableHead} w-32`}>Categoría</th>
                    <th className={`${tableHeadRight} w-32`}>Monto</th>
                  </tr>
                </thead>
                <tbody>
                  {movementList.map((mv) => {
                    const isSelected = selected?.ID === mv.ID;
                    return (
                      <tr
                        key={mv.ID}
                        tabIndex={0}
                        onClick={() => setSelected(mv)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') setSelected(mv);
                        }}
                        aria-label={`${formatDate(mv.MovementDate)} — ${mv.Description ?? 'Sin descripción'} — ${formatSignedMoney(mv.AmountMinor)}`}
                        className={`relative cursor-pointer border-t border-line/60 transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent ${
                          isSelected ? 'bg-accent-wash' : ''
                        }`}
                      >
                        {isSelected && (
                          <span
                            className="absolute bottom-0 left-0 top-0 w-[2px] bg-accent"
                            aria-hidden="true"
                          />
                        )}
                        <td className="whitespace-nowrap px-3 py-2 text-meta text-ink-3">
                          {formatDate(mv.MovementDate)}
                        </td>
                        <td
                          className="w-full max-w-0 truncate px-3 py-2 text-ink"
                          title={mv.Description ?? undefined}
                        >
                          {mv.Description ?? '—'}
                        </td>
                        <td className="truncate px-3 py-2 text-ink-2">{mv.categoryName || '—'}</td>
                        <td
                          className={`whitespace-nowrap px-3 py-2 text-right font-medium tabular-nums ${signedAmountClass(mv.AmountMinor)}`}
                        >
                          {formatSignedMoney(mv.AmountMinor)}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {/* Documentos de la cuenta */}
        <section aria-label="Documentos de la cuenta" className="space-y-2">
          <SectionHeader
            title={`Documentos${documents.status === 'ready' ? ` (${documentList.length})` : ''}`}
            action={{ to: '/documents', label: 'Ver todos los documentos' }}
          />
          {documents.status === 'loading' ? (
            <SkeletonRows rows={3} />
          ) : documents.status === 'error' ? (
            <ErrorState
              message="No se pudo cargar los documentos de la cuenta."
              onRetry={documents.retry}
            />
          ) : documentList.length === 0 ? (
            <EmptyState
              title="Sin documentos"
              description="Esta cuenta todavía no tiene documentos asociados."
            />
          ) : (
            <div className="overflow-x-auto rounded-card border border-line bg-surface">
              <table className="w-full text-body">
                <thead className="bg-raised">
                  <tr>
                    <th className={tableHead}>Nombre</th>
                    <th className={tableHead}>Institución</th>
                    <th className={tableHead}>Estado</th>
                    <th className={`${tableHead} w-24`}>Fecha</th>
                  </tr>
                </thead>
                <tbody>
                  {documentList.map((doc) => (
                    <tr key={doc.ID} className="border-t border-line/60">
                      <td className="w-full max-w-0 truncate px-3 py-2 text-ink">
                        {doc.DocumentName ?? 'Documento'}
                      </td>
                      <td className="px-3 py-2 text-ink-2">{doc.Institution ?? '—'}</td>
                      <td className="px-3 py-2">
                        <StatusBadge status={doc.Status} />
                      </td>
                      <td className="whitespace-nowrap px-3 py-2 text-meta text-ink-3">
                        {formatDate(doc.DocumentDate ?? doc.CreatedAt)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>

      <DetailDrawer
        movement={selected}
        accountName={selected ? (accountName ?? selected.AccountID) : null}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
