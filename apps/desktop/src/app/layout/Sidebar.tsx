import React, { useEffect, useState } from 'react';
import { NavLink } from 'react-router-dom';
import {
  IconChart,
  IconDocument,
  IconGear,
  IconGrid,
  IconShieldCheck,
  IconTransfer,
  IconWallet,
} from '../../components/icons';
import { useCollectionVersion } from '../../lib/dataBus';
import { getReviewEntries } from '../../api/review';
import { APP_VERSION } from '../../lib/version';

interface SidebarProps {
  collapsed: boolean;
}

type NavItemDef = {
  to: string;
  label: string;
  Icon: (p: { className?: string }) => React.ReactElement;
};

const systemNav: NavItemDef[] = [
  { to: '/', label: 'Overview', Icon: IconGrid },
  { to: '/accounts', label: 'Cuentas', Icon: IconWallet },
  { to: '/transactions', label: 'Movimientos', Icon: IconTransfer },
  { to: '/documents', label: 'Documentos', Icon: IconDocument },
  { to: '/review', label: 'Revisión', Icon: IconShieldCheck },
];

const toolsNav: NavItemDef[] = [
  { to: '/reports', label: 'Reportes', Icon: IconChart },
  { to: '/settings', label: 'Configuración', Icon: IconGear },
];

// Pending review count: the honest "you have items waiting" signal. It
// reads the real queue and refetches whenever anything invalidates the
// 'review' collection (a resolution, an import, a category delete).
function useReviewCount(): number {
  const [count, setCount] = useState(0);
  const reviewVersion = useCollectionVersion('review');
  useEffect(() => {
    getReviewEntries()
      .then((d) => setCount((d || []).length))
      .catch(() => {});
  }, [reviewVersion]);
  return count;
}

function NavItem({
  to,
  label,
  Icon,
  collapsed,
  badge,
}: NavItemDef & { collapsed: boolean; badge?: number }) {
  const showBadge = !!badge && badge > 0;
  return (
    <NavLink
      to={to}
      end={to === '/'}
      title={collapsed ? label : undefined}
      className={({ isActive }) =>
        `relative flex h-9 items-center mx-2 rounded-[8px] text-body font-medium transition-colors duration-[120ms] ${
          collapsed ? 'justify-center' : 'gap-3 px-3'
        } ${isActive ? 'bg-raised text-ink' : 'text-ink-2 hover:bg-interactive hover:text-ink'}`
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span
              className="absolute left-0 top-2 bottom-2 w-[3px] rounded-r-full bg-accent"
              aria-hidden="true"
            />
          )}
          <Icon className={isActive ? 'shrink-0 text-accent' : 'shrink-0 text-ink-3'} />
          <span className={`truncate ${collapsed ? 'sr-only' : ''}`}>{label}</span>
          {showBadge && (
            <span
              className={
                collapsed
                  ? 'absolute top-1 right-1 h-2 w-2 rounded-full bg-accent-strong'
                  : 'ml-auto min-w-5 rounded-full bg-accent-strong px-1.5 py-0.5 text-center text-meta font-mono text-white'
              }
              aria-label={`${badge} ítems pendientes de revisión`}
            >
              {!collapsed && badge}
            </span>
          )}
        </>
      )}
    </NavLink>
  );
}

function GroupLabel({ label, collapsed }: { label: string; collapsed: boolean }) {
  if (collapsed) return <div className="mx-4 my-2 h-px bg-line" aria-hidden="true" />;
  return (
    <p className="px-4 pt-4 pb-1.5 text-meta font-mono font-medium uppercase tracking-[0.08em] text-ink-3">
      {label}
    </p>
  );
}

export default function Sidebar({ collapsed }: SidebarProps) {
  const reviewCount = useReviewCount();
  return (
    <aside
      className={`flex shrink-0 flex-col border-r border-line bg-surface transition-[width] duration-[220ms] ease-out-tk ${
        collapsed ? 'w-16' : 'w-64'
      }`}
    >
      {/* Brand */}
      <div
        className={`flex h-12 shrink-0 items-center border-b border-line ${
          collapsed ? 'justify-center' : 'gap-2.5 px-4'
        }`}
      >
        <span className="h-2.5 w-2.5 shrink-0 bg-accent-strong" aria-hidden="true" />
        {!collapsed && (
          <span className="text-body font-semibold tracking-wide text-ink">FinanceApp</span>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex flex-col pb-2">
        <GroupLabel label="Sistema" collapsed={collapsed} />
        <div className="flex flex-col gap-0.5">
          {systemNav.map((item) => (
            <NavItem
              key={item.to}
              {...item}
              collapsed={collapsed}
              badge={item.to === '/review' ? reviewCount : undefined}
            />
          ))}
        </div>
        <GroupLabel label="Herramientas" collapsed={collapsed} />
        <div className="flex flex-col gap-0.5">
          {toolsNav.map((item) => (
            <NavItem key={item.to} {...item} collapsed={collapsed} />
          ))}
        </div>
      </nav>

      {/* Version — real metadata, not decoration */}
      <div className="mt-auto px-4 pb-4">
        {!collapsed && <span className="text-meta font-mono text-ink-3">v{APP_VERSION}</span>}
      </div>
    </aside>
  );
}
