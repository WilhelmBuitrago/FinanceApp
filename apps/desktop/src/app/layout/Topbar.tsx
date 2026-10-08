import React from 'react';
import { useLocation } from 'react-router-dom';
import { IconMenu } from '../../components/icons';

interface TopbarProps {
  onToggleSidebar: () => void;
}

const routeTitles: Record<string, string> = {
  '/': 'Overview',
  '/accounts': 'Cuentas',
  '/transactions': 'Movimientos',
  '/documents': 'Documentos',
  '/review': 'Revisión',
  '/reports': 'Reportes',
  '/settings': 'Configuración',
};

function titleFor(pathname: string): string {
  if (routeTitles[pathname]) return routeTitles[pathname];
  if (pathname.startsWith('/accounts/')) return 'Cuentas';
  return 'FinanceApp';
}

export default function Topbar({ onToggleSidebar }: TopbarProps) {
  const location = useLocation();
  const title = titleFor(location.pathname);

  return (
    <header className="flex h-12 shrink-0 items-center border-b border-line bg-bg px-4">
      <div className="flex min-w-0 items-center gap-3">
        <button
          onClick={onToggleSidebar}
          aria-label="Contraer panel lateral"
          className="flex h-8 w-8 items-center justify-center rounded-tk text-ink-3 transition-colors duration-[120ms] hover:bg-interactive hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
        >
          <IconMenu />
        </button>
        <div className="h-5 w-px bg-line" aria-hidden="true" />
        <h1 className="truncate text-title font-semibold text-ink">{title}</h1>
      </div>
    </header>
  );
}
