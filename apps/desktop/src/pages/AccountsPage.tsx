import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAccounts } from '../hooks/useAccounts';
import { useMovements } from '../hooks/useMovements';
import { accountTypeLabel } from '../lib/accountType';
import { formatDate } from '../lib/formatDate';
import { NewAccountDrawer } from '../components/NewAccountDrawer';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { SkeletonRows } from '../components/ui/SkeletonRows';
import { StatusBadge } from '../components/ui/StatusBadge';

/* Cuentas — the configuration center for financial sources: registered
   entities on the list, creation via the wizard drawer, honest empty state
   that offers configuration (never a dead-end table). */

const tableHead =
  'px-3 py-1.5 text-left font-medium uppercase tracking-[0.08em] text-meta text-ink-3';

const primaryButton =
  'inline-flex h-8 items-center rounded-[8px] bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40 disabled:hover:bg-accent-strong';

export default function AccountsPage() {
  const [refreshKey, setRefreshKey] = useState(0);
  const accounts = useAccounts(refreshKey);
  const movements = useMovements(refreshKey);
  const navigate = useNavigate();
  const [wizardOpen, setWizardOpen] = useState(false);

  /* Real movement count per account, derived client-side from real data. */
  const movementCounts = useMemo(() => {
    const counts = new Map<string, number>();
    for (const mv of movements.data ?? []) {
      counts.set(mv.AccountID, (counts.get(mv.AccountID) ?? 0) + 1);
    }
    return counts;
  }, [movements.data]);

  /* Accounts referenced by movements but never registered — surfaced in the
     empty state because it is a real integrity observation. */
  const unregisteredReferences = useMemo(() => {
    if (!movements.data || movements.loading) return 0;
    const registered = new Set((accounts.data ?? []).map((a) => a.ID));
    let count = 0;
    for (const id of movementCounts.keys()) if (!registered.has(id)) count += 1;
    return count;
  }, [movementCounts, accounts.data, movements.data, movements.loading]);

  const noAccounts = !accounts.loading && !accounts.error && (accounts.data?.length ?? 0) === 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Cuentas</h1>
          <p className="text-body text-ink-2">
            Tus entidades financieras: qué puedes importar y desde dónde.
          </p>
        </div>
        <button onClick={() => setWizardOpen(true)} className={primaryButton}>
          Nueva cuenta
        </button>
      </div>

      {accounts.loading ? (
        <SkeletonRows rows={5} />
      ) : accounts.error ? (
        <ErrorState
          message="No se pudo cargar las cuentas desde el servidor."
          onRetry={() => setRefreshKey((k) => k + 1)}
        />
      ) : noAccounts ? (
        <EmptyState
          title="No tienes cuentas configuradas"
          description={
            unregisteredReferences > 0
              ? `Configura una cuenta antes de importar documentos financieros. Hay ${unregisteredReferences} ${
                  unregisteredReferences === 1 ? 'cuenta referenciada' : 'cuentas referenciadas'
                } por movimientos importados que aún no tienen ficha.`
              : 'Configura una cuenta antes de importar documentos financieros.'
          }
          action={
            <button onClick={() => setWizardOpen(true)} className={primaryButton}>
              Configurar cuenta
            </button>
          }
        />
      ) : (
        <div className="overflow-x-auto rounded-[12px] border border-line bg-surface">
          <table className="w-full text-body">
            <thead className="bg-raised">
              <tr>
                <th className={tableHead}>Nombre</th>
                <th className={tableHead}>Producto</th>
                <th className={tableHead}>Institución</th>
                <th className={tableHead}>Moneda</th>
                <th className={`${tableHead} w-28`}>Estado</th>
                <th className={`${tableHead.replace('text-left', 'text-right')} w-32`}>
                  Movimientos
                </th>
                <th className={`${tableHead} w-28`}>Creada</th>
              </tr>
            </thead>
            <tbody>
              {(accounts.data ?? []).map((acc) => (
                <tr
                  key={acc.ID}
                  tabIndex={0}
                  onClick={() => navigate(`/accounts/${acc.ID}`)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') navigate(`/accounts/${acc.ID}`);
                  }}
                  aria-label={`${acc.Name} — ${accountTypeLabel(acc.AccountType)}`}
                  className="cursor-pointer border-t border-line/60 transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent"
                >
                  <td
                    className="w-full max-w-0 truncate px-3 py-2 font-medium text-ink"
                    title={acc.Name}
                  >
                    {acc.Name}
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 text-ink-2">
                    {accountTypeLabel(acc.AccountType)}
                  </td>
                  <td
                    className="w-full max-w-0 truncate px-3 py-2 text-ink-2"
                    title={acc.Institution ?? undefined}
                  >
                    {acc.Institution ?? '—'}
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 text-ink-2">{acc.Currency}</td>
                  <td className="px-3 py-2">
                    <StatusBadge status={acc.Status} />
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 text-right tabular-nums text-ink-2">
                    {movementCounts.get(acc.ID) ?? 0}
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 text-meta text-ink-3">
                    {formatDate(acc.CreatedAt)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {wizardOpen && (
        <NewAccountDrawer
          onClose={() => setWizardOpen(false)}
          onCreated={() => {
            setWizardOpen(false);
            setRefreshKey((k) => k + 1);
          }}
        />
      )}
    </div>
  );
}
