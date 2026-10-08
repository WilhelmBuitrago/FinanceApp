import React, { useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useAccounts } from '../hooks/useAccounts';
import {
  useCashFlow,
  useExpenseDistribution,
  useIncomeDistribution,
  useRecentDocuments,
  useRecentMovements,
} from '../hooks/useDashboard';
import { formatDate } from '../lib/formatDate';
import { formatPeriod } from '../lib/formatPeriod';
import { formatSignedMoney } from '../lib/formatMoney';
import { CashFlowChart } from '../components/CashFlowChart';
import { ChartCard } from '../components/ChartCard';
import { DonutChart } from '../components/DonutChart';
import { SectionHeader } from '../components/SectionHeader';
import { StatusBadge } from '../components/ui/StatusBadge';
import { SummaryStrip } from '../components/SummaryStrip';
import { EmptyState } from '../components/ui/EmptyState';
import { SkeletonRows } from '../components/ui/SkeletonRows';

/* Overview — the financial dashboard. Every figure comes from the backend's
   aggregates (S12 §34): cash flow points, distributions, recent lists.
   No invented metrics: empty charts get contextual states, never fake
   data to fill space. The money's flow is the protagonist. */

const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000;
const ATTENTION_STATUSES = new Set(['REVIEW_REQUIRED', 'FAILED']);

const tableHead = 'px-3 py-1.5 text-left font-medium uppercase tracking-[0.08em] text-meta text-ink-3';
const tableHeadRight = 'px-3 py-1.5 text-right font-medium uppercase tracking-[0.08em] text-meta text-ink-3';

const primaryButton =
  'inline-flex h-8 items-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';

function greeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return 'Buenos días';
  if (hour < 19) return 'Buenas tardes';
  return 'Buenas noches';
}

function todayIso(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

export default function DashboardPage() {
  const accounts = useAccounts();
  const cashFlow = useCashFlow();
  const incomeDist = useIncomeDistribution();
  const expenseDist = useExpenseDistribution();
  const recentMovements = useRecentMovements();
  const recentDocuments = useRecentDocuments();

  const accountNames = useMemo(() => {
    const map = new Map<string, string>();
    for (const account of accounts.data ?? []) map.set(account.ID, account.Name);
    return map;
  }, [accounts.data]);

  /* Backend points sorted for the chart. */
  const points = useMemo(
    () => [...(cashFlow.data ?? [])].sort((a, b) => a.date.localeCompare(b.date)),
    [cashFlow.data],
  );

  /* 30-day window over the backend's daily aggregates (visual transform). */
  const last30 = useMemo(() => {
    if (!cashFlow.data) return null;
    const cutoff = Date.now() - THIRTY_DAYS_MS;
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
  }, [cashFlow.data]);

  const attentionDocs = useMemo(
    () => (recentDocuments.data ?? []).filter((d) => ATTENTION_STATUSES.has(d.Status)),
    [recentDocuments.data],
  );

  const firstUse =
    !accounts.loading &&
    !cashFlow.loading &&
    !recentMovements.loading &&
    !recentDocuments.loading &&
    (accounts.data?.length ?? 0) === 0 &&
    (recentMovements.data?.length ?? 0) === 0 &&
    (recentDocuments.data?.length ?? 0) === 0;

  return (
    <div className="max-w-6xl space-y-8">
      {/* Context */}
      <p className="text-title font-medium text-ink">
        {greeting()}, Wilhelm
        <span className="ml-2 text-meta font-normal text-ink-3">{formatDate(todayIso())}</span>
      </p>

      {firstUse ? (
        <EmptyState
          title="Aún no hay datos financieros"
          description="Configura una cuenta e importa un documento para comenzar a ver tu situación."
          action={
            <Link to="/accounts" className={primaryButton}>
              Configurar cuenta
            </Link>
          }
        />
      ) : (
        <>
          {/* 1. Financial summary — real 30-day sums over backend aggregates */}
          <section aria-label="Situación financiera">
            {cashFlow.loading ? (
              <div className="h-14 w-full max-w-2xl animate-pulse rounded-tk bg-interactive" aria-hidden="true" />
            ) : cashFlow.error ? (
              <p className="text-body text-ink-2">
                No se pudo cargar la situación financiera desde el servidor.
              </p>
            ) : last30 && last30.count > 0 ? (
              <SummaryStrip income={last30.income} expense={-last30.expense} periodLabel="30 días" />
            ) : (
              <p className="text-body text-ink-2">
                Sin movimientos en los últimos 30 días.
              </p>
            )}
          </section>

          {/* 2. Cash flow — the protagonist */}
          <section aria-label="Flujo de dinero" className="space-y-2">
            <SectionHeader title="Flujo de dinero" />
            {cashFlow.loading ? (
              <div className="h-48 w-full animate-pulse rounded-tk border border-line bg-surface" aria-hidden="true" />
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

          {/* 3. Distributions — where money comes from / goes to */}
          <section aria-label="Distribución" className="space-y-2">
            <SectionHeader title="Distribución" />
            <div className="grid gap-6 md:grid-cols-2">
              {incomeDist.loading || expenseDist.loading ? (
                <div className="h-40 w-full animate-pulse rounded-tk border border-line bg-surface" aria-hidden="true" />
              ) : (
                <>
                  <ChartCard title="Ingresos">
                    {!incomeDist.loading && !incomeDist.error && (incomeDist.data?.length ?? 0) > 0 ? (
                      <DonutChart slices={incomeDist.data ?? []} />
                    ) : (
                      <EmptyState
                        title="Sin ingresos clasificados"
                        description={
                          incomeDist.error
                            ? 'No se pudo cargar la distribución.'
                            : 'Aún no hay ingresos clasificados para mostrar la distribución.'
                        }
                      />
                    )}
                  </ChartCard>
                  <ChartCard title="Egresos">
                    {!expenseDist.loading && !expenseDist.error && (expenseDist.data?.length ?? 0) > 0 ? (
                      <DonutChart slices={expenseDist.data ?? []} />
                    ) : (
                      <EmptyState
                        title="Sin egresos clasificados"
                        description={
                          expenseDist.error
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

          {/* 4. Actividad reciente — quick access to history (max 5, backend-limited) */}
          <section aria-label="Actividad reciente" className="space-y-2">
            <SectionHeader
              title="Actividad reciente"
              action={{ to: '/transactions', label: 'Ver todos los movimientos' }}
            />
            {recentMovements.loading ? (
              <SkeletonRows rows={5} />
            ) : (recentMovements.data?.length ?? 0) === 0 ? (
              <EmptyState
                title="Sin movimientos"
                description="Los movimientos aparecerán al importar documentos."
              />
            ) : (
              <div className="overflow-hidden rounded-tk border border-line bg-surface">
                <table className="w-full table-fixed text-body">
                  <thead className="bg-raised">
                    <tr>
                      <th className={`${tableHead} w-24`}>Fecha</th>
                      <th className={tableHead}>Descripción</th>
                      <th className={`${tableHead} w-40`}>Cuenta</th>
                      <th className={`${tableHeadRight} w-32`}>Monto</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(recentMovements.data ?? []).map((mv) => (
                      <tr key={mv.ID} className="border-t border-line/60">
                        <td className="whitespace-nowrap px-3 py-2 text-meta text-ink-3">
                          {formatDate(mv.MovementDate)}
                        </td>
                        <td className="w-full max-w-0 truncate px-3 py-2 text-ink" title={mv.Description ?? undefined}>
                          {mv.Description ?? '—'}
                        </td>
                        <td className="truncate px-3 py-2 text-ink-2" title={accountNames.get(mv.AccountID) ?? undefined}>
                          {accountNames.get(mv.AccountID) ?? '—'}
                        </td>
                        <td
                          className={`whitespace-nowrap px-3 py-2 text-right font-medium tabular-nums ${
                            mv.AmountMinor > 0 ? 'text-pos' : mv.AmountMinor < 0 ? 'text-neg' : 'text-ink'
                          }`}
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

          {/* 5. Requiere atención — only when something needs it */}
          {attentionDocs.length > 0 && (
            <section aria-label="Requiere atención" className="space-y-2">
              <SectionHeader title="Requiere atención" action={{ to: '/documents', label: 'Ver documentos' }} />
              <div className="overflow-hidden rounded-tk border border-line bg-surface">
                {attentionDocs.slice(0, 3).map((doc) => (
                  <div
                    key={doc.ID}
                    className="flex items-center justify-between gap-3 border-t border-line/60 px-3 py-2 first:border-t-0"
                  >
                    <div className="min-w-0">
                      <p className="truncate text-ink">{doc.DocumentName ?? 'Documento'}</p>
                      <p className="truncate text-meta text-ink-3">
                        {[accountNames.get(doc.AccountID ?? ''), formatPeriod(doc.PeriodStart, doc.PeriodEnd)]
                          .filter(Boolean)
                          .join(' · ')}
                      </p>
                    </div>
                    <StatusBadge status={doc.Status} />
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* 6. Documentos recientes (max 3) */}
          <section aria-label="Documentos recientes" className="space-y-2">
            <SectionHeader
              title="Documentos recientes"
              action={{ to: '/documents', label: 'Ver todos los documentos' }}
            />
            {recentDocuments.loading ? (
              <SkeletonRows rows={3} />
            ) : (recentDocuments.data?.length ?? 0) === 0 ? (
              <EmptyState
                title="Sin documentos"
                description="Los documentos bancarios importados aparecerán aquí."
              />
            ) : (
              <div className="overflow-hidden rounded-tk border border-line bg-surface">
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
                    {(recentDocuments.data ?? []).slice(0, 3).map((doc) => (
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
        </>
      )}
    </div>
  );
}
