import { createBrowserRouter } from 'react-router-dom';
import type { ReactElement } from 'react';

import AppShell from './layout/AppShell';
import { NAV_ITEMS } from './nav';
import DashboardPage from '../pages/DashboardPage';
import AccountsPage from '../pages/AccountsPage';
import AccountDetailPage from '../pages/AccountDetailPage';
import TransactionsPage from '../pages/TransactionsPage';
import DocumentsPage from '../pages/DocumentsPage';
import ReviewPage from '../pages/ReviewPage';
import ReportsPage from '../pages/ReportsPage';
import SettingsPage from '../pages/SettingsPage';

/* Routes come from the nav registry; composition stays here. The detail
   route has no nav entry (it carries no chrome label) and is listed
   once, next to the registry-derived ones. */

const elementByPath: Record<string, ReactElement> = {
  '/': <DashboardPage />,
  '/accounts': <AccountsPage />,
  '/transactions': <TransactionsPage />,
  '/documents': <DocumentsPage />,
  '/review': <ReviewPage />,
  '/reports': <ReportsPage />,
  '/settings': <SettingsPage />,
};

export const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      ...NAV_ITEMS.map((item) => ({ path: item.to, element: elementByPath[item.to] })),
      { path: '/accounts/:id', element: <AccountDetailPage /> },
    ],
  },
]);
