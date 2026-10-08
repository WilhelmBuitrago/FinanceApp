import { createBrowserRouter } from 'react-router-dom';
import AppShell from './layout/AppShell';
import DashboardPage from '../pages/DashboardPage';
import AccountsPage from '../pages/AccountsPage';
import AccountDetailPage from '../pages/AccountDetailPage';
import TransactionsPage from '../pages/TransactionsPage';
import DocumentsPage from '../pages/DocumentsPage';
import ReviewPage from '../pages/ReviewPage';
import ReportsPage from '../pages/ReportsPage';
import SettingsPage from '../pages/SettingsPage';

export const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      { path: '/', element: <DashboardPage /> },
      { path: '/accounts', element: <AccountsPage /> },
      { path: '/accounts/:id', element: <AccountDetailPage /> },
      { path: '/transactions', element: <TransactionsPage /> },
      { path: '/documents', element: <DocumentsPage /> },
      { path: '/review', element: <ReviewPage /> },
      { path: '/reports', element: <ReportsPage /> },
      { path: '/settings', element: <SettingsPage /> },
    ],
  },
]);
