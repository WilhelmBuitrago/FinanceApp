import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

import { AddMovementDrawer } from '../components/AddMovementDrawer';
import { AttentionSection } from '../components/dashboard/AttentionSection';
import { CashFlowSection } from '../components/dashboard/CashFlowSection';
import { DistributionSection } from '../components/dashboard/DistributionSection';
import { FinancialSummarySection } from '../components/dashboard/FinancialSummarySection';
import { NetTotalCard, type AccountNet } from '../components/dashboard/NetTotalCard';
import { NotificationBell } from '../components/dashboard/NotificationBell';
import { QuickActions } from '../components/dashboard/QuickActions';
import { RecentDocumentsTable } from '../components/dashboard/RecentDocumentsTable';
import { RecentMovementsTable } from '../components/dashboard/RecentMovementsTable';
import { EmptyState } from '../components/ui/EmptyState';
import { primaryButton } from '../components/ui/buttons';
import { useAccountNames } from '../hooks/useAccountNames';
import { useAccounts } from '../hooks/useAccounts';
import { useReviewNotifications } from '../hooks/useReviewNotifications';
import {
  useCashFlow,
  useExpenseDistribution,
  useIncomeDistribution,
  useRecentDocuments,
  useRecentMovements,
} from '../hooks/useDashboard';
import { useMovements } from '../hooks/useMovements';
import { formatDate } from '../lib/formatDate';

/* Overview — the financial dashboard. Every figure comes from the backend's
   aggregates (S12 §34): cash flow points, distributions, recent lists.
   No invented metrics: empty charts get contextual states, never fake
   data to fill space. The money's flow is the protagonist.

   This screen composes: the data hooks stay here, every visual section
   lives in components/dashboard/. */

const ATTENTION_STATUSES = new Set(['REVIEW_REQUIRED', 'FAILED']);

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
  const [addMovementOpen, setAddMovementOpen] = useState(false);
  const { notifications, notify } = useReviewNotifications();
  const accounts = useAccounts();
  const cashFlow = useCashFlow();
  const incomeDist = useIncomeDistribution();
  const expenseDist = useExpenseDistribution();
  const recentMovements = useRecentMovements();
  const recentDocuments = useRecentDocuments();
  const movements = useMovements();

  const accountNames = useAccountNames(accounts);

  /* Net total and its per-account breakdown: real sums over real data. */
  const { totalNet, perAccount } = useMemo(() => {
    const map = new Map<string, number>();
    let total = 0;
    const data = movements.data ?? [];
    for (const mv of data) {
      const amt = Number(mv.AmountMinor) || 0;
      total += amt;
      const acc = mv.AccountID as string;
      map.set(acc, (map.get(acc) ?? 0) + amt);
    }
    const rows: AccountNet[] = Array.from(map.entries()).map(([id, net]) => ({
      id,
      name: accountNames.get(id) ?? id,
      net,
    }));
    return { totalNet: total, perAccount: rows };
  }, [movements.data, accountNames]);

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
    <div className="space-y-8">
      <div className="flex items-baseline justify-between gap-4">
        <div>
          <h1 className="text-page font-semibold tracking-tight text-ink">{greeting()}</h1>
          <p className="text-meta text-ink-3">{formatDate(todayIso())}</p>
        </div>
        {firstUse ? null : (
          <div className="flex items-center gap-2">
            <NotificationBell notifications={notifications} />
            <QuickActions onAddMovement={() => setAddMovementOpen(true)} />
          </div>
        )}
      </div>

      {!firstUse && <NetTotalCard totalNet={totalNet} perAccount={perAccount} />}

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
          <FinancialSummarySection cashFlow={cashFlow} />
          <CashFlowSection cashFlow={cashFlow} />
          <DistributionSection income={incomeDist} expense={expenseDist} />
          <RecentMovementsTable movements={recentMovements} accountNames={accountNames} />
          {attentionDocs.length > 0 && (
            <AttentionSection documents={attentionDocs} accountNames={accountNames} />
          )}
          <RecentDocumentsTable documents={recentDocuments} accountNames={accountNames} />
        </>
      )}
      {/* Add Movement Drawer */}
      {addMovementOpen && (
        <AddMovementDrawer
          accounts={accounts.data ?? []}
          onClose={() => setAddMovementOpen(false)}
          onDuplicate={notify}
        />
      )}
    </div>
  );
}
