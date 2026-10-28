import { useMemo, useState } from 'react';

import type { CashFlowPoint } from '../../api/dashboard';
import type { UseQueryState } from '../../types/api';
import { SummaryStrip } from '../SummaryStrip';

/* Financial summary — real 30-day sums over the backend's daily
   aggregates (a visual transform, never a new financial rule).

   The window anchors at mount through a lazy state initializer: calling
   Date.now() inside the memo made the "last 30 days" cutoff drift with
   every recompute, which the React purity rules reject. */

const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000;

export function FinancialSummarySection({
  cashFlow,
}: {
  cashFlow: UseQueryState<CashFlowPoint[]>;
}) {
  const [mountedAt] = useState(() => Date.now());

  const last30 = useMemo(() => {
    if (!cashFlow.data) return null;
    const cutoff = mountedAt - THIRTY_DAYS_MS;
    let income = 0;
    let expense = 0;
    let count = 0;
    for (const p of cashFlow.data) {
      const t = new Date(p.date).getTime();
      if (Number.isNaN(t) || t < cutoff) continue;
      count += 1;
      income += p.inflow;
      expense += p.outflow;
    }
    return { income, expense, count };
  }, [cashFlow.data, mountedAt]);

  return (
    <section aria-label="Situación financiera">
      {cashFlow.loading ? (
        <div
          className="h-14 w-full max-w-2xl animate-pulse rounded-tk bg-interactive"
          aria-hidden="true"
        />
      ) : cashFlow.error ? (
        <p className="text-body text-ink-2">
          No se pudo cargar la situación financiera desde el servidor.
        </p>
      ) : last30 && last30.count > 0 ? (
        <SummaryStrip income={last30.income} expense={-last30.expense} periodLabel="30 días" />
      ) : (
        <p className="text-body text-ink-2">Sin movimientos en los últimos 30 días.</p>
      )}
    </section>
  );
}
