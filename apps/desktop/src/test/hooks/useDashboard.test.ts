import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { invalidateCollections } from '../../lib/dataBus';
import { getCashFlow, getExpenseDistribution, getIncomeDistribution } from '../../api/dashboard';
import { getInstitutions } from '../../api/institutions';
import { getRecentDocuments } from '../../api/documents';
import { getRecentMovements } from '../../api/movements';
import {
  useCashFlow,
  useExpenseDistribution,
  useIncomeDistribution,
  useInstitutions,
  useRecentDocuments,
  useRecentMovements,
} from '../../hooks/useDashboard';

/* The dashboard data hooks: the endpoint each one calls and the
   collection each one subscribes to. Institutions are a fixed catalog:
   they fetch on mount and never subscribe, so an invalidation must not
   refetch them. */

vi.mock('../../api/dashboard', () => ({
  getCashFlow: vi.fn(),
  getIncomeDistribution: vi.fn(),
  getExpenseDistribution: vi.fn(),
}));
vi.mock('../../api/institutions', () => ({ getInstitutions: vi.fn() }));
vi.mock('../../api/documents', () => ({ getRecentDocuments: vi.fn() }));
vi.mock('../../api/movements', () => ({ getRecentMovements: vi.fn() }));

describe('useDashboard hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getCashFlow).mockResolvedValue([]);
    vi.mocked(getIncomeDistribution).mockResolvedValue([]);
    vi.mocked(getExpenseDistribution).mockResolvedValue([]);
    vi.mocked(getInstitutions).mockResolvedValue([
      { code: 'NU', name: 'Nu', supported_account_types: ['BANK_ACCOUNT', 'CREDIT_CARD'] },
    ]);
    vi.mocked(getRecentDocuments).mockResolvedValue([]);
    vi.mocked(getRecentMovements).mockResolvedValue([]);
  });

  it('each hook fetches its own endpoint', async () => {
    renderHook(() => ({
      cashFlow: useCashFlow(),
      income: useIncomeDistribution(),
      expense: useExpenseDistribution(),
      institutions: useInstitutions(),
      recentDocuments: useRecentDocuments(),
      recentMovements: useRecentMovements(),
    }));

    await waitFor(() => {
      expect(getCashFlow).toHaveBeenCalledTimes(1);
      expect(getIncomeDistribution).toHaveBeenCalledTimes(1);
      expect(getExpenseDistribution).toHaveBeenCalledTimes(1);
      expect(getInstitutions).toHaveBeenCalledTimes(1);
      expect(getRecentDocuments).toHaveBeenCalledTimes(1);
      expect(getRecentMovements).toHaveBeenCalledTimes(1);
    });
  });

  it('the dashboard hooks refetch when the dashboard collection is invalidated', async () => {
    const { result } = renderHook(() => ({
      cashFlow: useCashFlow(),
      income: useIncomeDistribution(),
      expense: useExpenseDistribution(),
    }));

    await waitFor(() => expect(result.current.cashFlow.loading).toBe(false));

    invalidateCollections('dashboard');

    await waitFor(() => {
      expect(getCashFlow).toHaveBeenCalledTimes(2);
      expect(getIncomeDistribution).toHaveBeenCalledTimes(2);
      expect(getExpenseDistribution).toHaveBeenCalledTimes(2);
    });
  });

  it('recent movements refetch with the movements collection, recent documents with documents', async () => {
    const { result } = renderHook(() => ({
      recentMovements: useRecentMovements(),
      recentDocuments: useRecentDocuments(),
    }));

    await waitFor(() => expect(result.current.recentMovements.loading).toBe(false));

    invalidateCollections('movements', 'documents');

    await waitFor(() => {
      expect(getRecentMovements).toHaveBeenCalledTimes(2);
      expect(getRecentDocuments).toHaveBeenCalledTimes(2);
    });
  });

  it('institutions never subscribe: an invalidation must not refetch them', async () => {
    const { result } = renderHook(() => ({ institutions: useInstitutions() }));

    await waitFor(() => expect(result.current.institutions.loading).toBe(false));
    expect(getInstitutions).toHaveBeenCalledTimes(1);

    invalidateCollections(
      'dashboard',
      'movements',
      'documents',
      'accounts',
      'categories',
      'review',
    );

    await waitFor(() => expect(result.current.institutions.loading).toBe(false));
    expect(getInstitutions).toHaveBeenCalledTimes(1);
  });
});
