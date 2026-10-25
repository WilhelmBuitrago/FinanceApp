import type { DistributionSlice } from '../../api/dashboard';
import type { UseQueryState } from '../../types/api';
import { ChartCard } from '../ChartCard';
import { DonutChart } from '../DonutChart';
import { EmptyState } from '../ui/EmptyState';
import { SectionHeader } from '../SectionHeader';

/* Distributions — where the money comes from and where it goes, as the
   backend classifies it. Two cards, each resolving on its own. */

export function DistributionSection({
  income,
  expense,
}: {
  income: UseQueryState<DistributionSlice[]>;
  expense: UseQueryState<DistributionSlice[]>;
}) {
  return (
    <section aria-label="Distribución" className="space-y-2">
      <SectionHeader title="Distribución" />
      <div className="grid gap-6 md:grid-cols-2">
        {income.loading || expense.loading ? (
          <div
            className="h-40 w-full animate-pulse rounded-card border border-line bg-surface"
            aria-hidden="true"
          />
        ) : (
          <>
            <ChartCard title="Ingresos">
              {!income.loading && !income.error && (income.data?.length ?? 0) > 0 ? (
                <DonutChart slices={income.data ?? []} />
              ) : (
                <EmptyState
                  title="Sin ingresos clasificados"
                  description={
                    income.error
                      ? 'No se pudo cargar la distribución.'
                      : 'Aún no hay ingresos clasificados para mostrar la distribución.'
                  }
                />
              )}
            </ChartCard>
            <ChartCard title="Egresos">
              {!expense.loading && !expense.error && (expense.data?.length ?? 0) > 0 ? (
                <DonutChart slices={expense.data ?? []} />
              ) : (
                <EmptyState
                  title="Sin egresos clasificados"
                  description={
                    expense.error
                      ? 'No se pudo cargar la distribución.'
                      : 'Aún no hay egresos clasificados para mostrar la distribución.'
                  }
                />
              )}
            </ChartCard>
          </>
        )}
      </div>
    </section>
  );
}
