import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { getAccounts } from '../api/accounts';
import { getCategories } from '../api/categories';
import { ApiError } from '../api/client';
import {
  discardReviewEntry,
  getReviewEntries,
  resolveReviewCategory,
  resolveReviewMovement,
  resolveReviewPeriod,
} from '../api/review';
import type { ReviewItem } from '../api/review';
import { useReviewQueue } from './useReviewQueue';

vi.mock('../api/review', () => ({
  getReviewEntries: vi.fn(),
  resolveReviewMovement: vi.fn(),
  resolveReviewCategory: vi.fn(),
  resolveReviewPeriod: vi.fn(),
  discardReviewEntry: vi.fn(),
}));
vi.mock('../api/categories', () => ({ getCategories: vi.fn() }));
vi.mock('../api/accounts', () => ({ getAccounts: vi.fn() }));

function parkedItem(id: string, kind: string, payload: object): ReviewItem {
  return {
    ID: id,
    Kind: kind as ReviewItem['Kind'],
    Status: 'PENDING',
    CreatedAt: '2026-10-24T00:00:00Z',
    PayloadJson: JSON.stringify(payload),
  };
}

const extraction = () =>
  parkedItem('rv_1', 'EXTRACTION', {
    description: 'Compra',
    amount_minor: -1500,
    account_id: 'acc_1',
    date: '2026-10-01',
  });

const flush = async () => {
  await act(async () => {});
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getReviewEntries).mockResolvedValue([extraction()]);
  vi.mocked(getCategories).mockResolvedValue([]);
  vi.mocked(getAccounts).mockResolvedValue([]);
  vi.mocked(resolveReviewMovement).mockResolvedValue({} as never);
  vi.mocked(resolveReviewCategory).mockResolvedValue({} as never);
  vi.mocked(resolveReviewPeriod).mockResolvedValue(undefined);
  vi.mocked(discardReviewEntry).mockResolvedValue(undefined);
});

describe('useReviewQueue', () => {
  it('loads the queue with the categories and accounts its cards need', async () => {
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    expect(result.current.items).toHaveLength(1);
    expect(getReviewEntries).toHaveBeenCalledTimes(1);
    expect(getCategories).toHaveBeenCalledTimes(1);
    expect(getAccounts).toHaveBeenCalledTimes(1);
  });

  it('resolves a correction with the payload values and removes the item', async () => {
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    await act(async () => {
      await result.current.resolveMovement(result.current.items[0]);
    });
    expect(resolveReviewMovement).toHaveBeenCalledWith('rv_1', {
      description: 'Compra',
      amount_minor: -1500,
      date: '2026-10-01',
      type: 'EXPENSE',
      account_id: 'acc_1',
    });
    expect(result.current.items).toHaveLength(0);
    expect(result.current.notice).toBeNull();
  });

  it('refuses to resolve without date and account, with the same copy as the old alert', async () => {
    vi.mocked(getReviewEntries).mockResolvedValue([parkedItem('rv_2', 'EXTRACTION', {})]);
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    await act(async () => {
      await result.current.resolveMovement(result.current.items[0]);
    });
    expect(resolveReviewMovement).not.toHaveBeenCalled();
    expect(result.current.notice).toBe('Completá fecha, monto y cuenta para confirmar');
    expect(result.current.items).toHaveLength(1);
  });

  it('treats a 409 as "already resolved": tells the human and removes the item', async () => {
    vi.mocked(resolveReviewMovement).mockRejectedValue(new ApiError('gone', 409, {}));
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    await act(async () => {
      await result.current.resolveMovement(result.current.items[0]);
    });
    expect(result.current.notice).toBe('Este ítem ya fue resuelto');
    expect(result.current.items).toHaveLength(0);
  });

  it('reports any other failure inline and keeps the item', async () => {
    vi.mocked(resolveReviewMovement).mockRejectedValue(new ApiError('boom', 500, {}));
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    await act(async () => {
      await result.current.resolveMovement(result.current.items[0]);
    });
    expect(result.current.notice).toBe('Error al confirmar la corrección');
    expect(result.current.items).toHaveLength(1);
  });

  it('sends only the category for a CATEGORY item (null = leave it uncategorized)', async () => {
    vi.mocked(getReviewEntries).mockResolvedValue([
      parkedItem('rv_3', 'CATEGORY', { movement_id: 'mv_1', category_name: 'Vieja' }),
    ]);
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    await act(async () => {
      await result.current.resolveCategory(result.current.items[0], null);
    });
    expect(resolveReviewCategory).toHaveBeenCalledWith('rv_3', null);
    expect(result.current.items).toHaveLength(0);
  });

  it('resolves the document period with the chosen month and year', async () => {
    vi.mocked(getReviewEntries).mockResolvedValue([
      parkedItem('rv_4', 'DOCUMENT_PERIOD', { document_id: 'doc_1' }),
    ]);
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    act(() => {
      result.current.setField('rv_4', { month: '9', year: '2026' });
    });
    await act(async () => {
      await result.current.resolvePeriod(result.current.items[0]);
    });
    expect(resolveReviewPeriod).toHaveBeenCalledWith('rv_4', 9, 2026);
    expect(result.current.items).toHaveLength(0);
  });

  it('discards an item without writing anything', async () => {
    const { result } = renderHook(() => useReviewQueue());
    await flush();
    await act(async () => {
      await result.current.discard(result.current.items[0]);
    });
    expect(discardReviewEntry).toHaveBeenCalledWith('rv_1');
    expect(result.current.items).toHaveLength(0);
  });
});
