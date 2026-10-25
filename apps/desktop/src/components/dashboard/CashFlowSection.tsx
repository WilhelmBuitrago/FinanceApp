import { useMemo } from 'react';

import type { CashFlowPoint } from '../../api/dashboard';
import type { UseQueryState } from '../../types/api';
import { CashFlowChart } from '../CashFlowChart';
import { ChartCard } from '../ChartCard';
import { EmptyState } from '../ui/EmptyState';
import { SectionHeader } from '../SectionHeader';

/* Cash flow — the protagonist: the backend's daily points, sorted, as a
   line with its area. Two points minimum; below that the honest state
   says so instead of drawing a line out of one day. */

export function CashFlowSection({ cashFlow }: { cashFlow: UseQueryState<CashFlowPoint[]> }) {
  const points = useMemo(
    () => [...(cashFlow.data ?? [])].sort((a, b) => a.date.localeCompare(b.date)),
    [cashFlow.data],
  );

  return (
    <section aria-label="Flujo de dinero" className="space-y-2">
      <SectionHeader title="Flujo de dinero" />
      {cashFlow.loading ? (
        <div
          className="h-48 w-full animate-pulse rounded-card border border-line bg-surface"
          aria-hidden="true"
        />
      ) : cashFlow.error ? (
        <EmptyState
          title="No se pudo cargar el flujo"
          description="No se pudo cargar la información desde el servidor."
        />
      ) : points.length >= 2 ? (
        <ChartCard title="Flujo de dinero" context={`${points.length} días con actividad`}>
          <CashFlowChart points={points} />
        </ChartCard>
      ) : (
        <ChartCard title="Flujo de dinero">
          <EmptyState
            title="Aún no hay suficiente información"
            description="Importa un documento para comenzar a ver el flujo financiero."
          />
        </ChartCard>
      )}
    </section>
  );
}
