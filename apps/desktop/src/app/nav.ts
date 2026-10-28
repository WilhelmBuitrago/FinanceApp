import type { ReactElement } from 'react';

import {
  IconChart,
  IconDocument,
  IconGear,
  IconGrid,
  IconShieldCheck,
  IconTransfer,
  IconWallet,
} from '../components/icons';

/* The one route registry. The router, the Sidebar and the Topbar each
   carried their own list of routes before the audit: three lists that
   drift the moment a page is added. Labels here are the topbar titles;
   groups feed the sidebar sections. */

export interface NavItem {
  to: string;
  label: string;
  Icon: (p: { className?: string }) => ReactElement;
  group: 'Sistema' | 'Herramientas';
}

export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Overview', Icon: IconGrid, group: 'Sistema' },
  { to: '/accounts', label: 'Cuentas', Icon: IconWallet, group: 'Sistema' },
  { to: '/transactions', label: 'Movimientos', Icon: IconTransfer, group: 'Sistema' },
  { to: '/documents', label: 'Documentos', Icon: IconDocument, group: 'Sistema' },
  { to: '/review', label: 'Revisión', Icon: IconShieldCheck, group: 'Sistema' },
  { to: '/reports', label: 'Reportes', Icon: IconChart, group: 'Herramientas' },
  { to: '/settings', label: 'Configuración', Icon: IconGear, group: 'Herramientas' },
];

/* Topbar title: the nav label for a known route, the parent route's
   label for detail paths (/accounts/:id shows "Cuentas"). */

export function titleFor(pathname: string): string {
  const exact = NAV_ITEMS.find((item) => item.to === pathname);
  if (exact) return exact.label;
  const parent = NAV_ITEMS.find((item) => item.to !== '/' && pathname.startsWith(`${item.to}/`));
  return parent?.label ?? 'FinanceApp';
}
