import type { UseQueryState } from '../../types/api';
import type { Movement } from '../../types/movement';
import { formatDate } from '../../lib/formatDate';
import { formatSignedMoney, signedAmountClass } from '../../lib/formatMoney';
import { EmptyState } from '../ui/EmptyState';
import { SkeletonRows } from '../ui/SkeletonRows';
import { SectionHeader } from '../SectionHeader';
import { tableHead, tableHeadRight } from '../ui/table';

/* Recent activity — the last movements (the backend already limits the
   list to five). Every documented record renders as its own row, never
   deduplicated. */

export function RecentMovementsTable({
  movements,
  accountNames,
}: {
  movements: UseQueryState<Movement[]>;
  accountNames: Map<string, string>;
}) {
  return (
    <section aria-label="Actividad reciente" className="space-y-2">
      <SectionHeader
        title="Actividad reciente"
        action={{ to: '/transactions', label: 'Ver todos los movimientos' }}
      />
      {movements.loading ? (
        <SkeletonRows rows={5} />
      ) : (movements.data?.length ?? 0) === 0 ? (
        <EmptyState
          title="Sin movimientos"
          description="Los movimientos aparecerán al importar documentos."
        />
      ) : (
        <div className="overflow-hidden rounded-card border border-line bg-surface">
          <table className="w-full table-fixed text-body">
            <thead className="bg-raised">
              <tr>
                <th className={`${tableHead} w-24`}>Fecha</th>
                <th className={tableHead}>Descripción</th>
                <th className={`${tableHead} w-32`}>Categoría</th>
                <th className={`${tableHead} w-40`}>Cuenta</th>
                <th className={`${tableHeadRight} w-32`}>Monto</th>
              </tr>
            </thead>
            <tbody>
              {(movements.data ?? []).map((mv) => (
                <tr key={mv.ID} className="border-t border-line/60">
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
                    className="truncate px-3 py-2 text-ink-2"
                    title={accountNames.get(mv.AccountID) ?? undefined}
                  >
                    {accountNames.get(mv.AccountID) ?? '—'}
                  </td>
                  <td
                    className={`whitespace-nowrap px-3 py-2 text-right font-medium tabular-nums ${signedAmountClass(mv.AmountMinor)}`}
                  >
                    {formatSignedMoney(mv.AmountMinor)}
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
