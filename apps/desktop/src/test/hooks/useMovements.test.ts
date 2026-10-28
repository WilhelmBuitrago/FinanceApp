import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { invalidateCollections } from '../../lib/dataBus';
import { getMovements } from '../../api/movements';
import { useMovements } from '../../hooks/useMovements';

/* The movements list hook: wiring over useListQuery — the endpoint it
   calls, the state it exposes and the collection that refreshes it. */

vi.mock('../../api/movements', () => ({ getMovements: vi.fn() }));

const movement = {
  ID: 'mv-1',
  FinancialEventID: 'fe-1',
  AccountID: 'acc-1',
  MovementDate: '2026-10-20',
  AmountMinor: -1500,
  Description: 'Mercado',
  Status: 'ACTIVE',
  categoryName: 'Comida',
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
};

describe('useMovements', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads the movements and exposes the loaded state', async () => {
    vi.mocked(getMovements).mockResolvedValue([movement]);

    const { result } = renderHook(() => useMovements());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data?.[0].AmountMinor).toBe(-1500);
    expect(result.current.data?.[0].categoryName).toBe('Comida');
    expect(getMovements).toHaveBeenCalledTimes(1);
  });

  it('surfaces a failed load as the error state', async () => {
    vi.mocked(getMovements).mockRejectedValue(new Error('backend down'));

    const { result } = renderHook(() => useMovements());

    await waitFor(() => expect(result.current.error).toBeInstanceOf(Error));
    expect(result.current.data).toBeUndefined();
  });

  it('refetches when the movements collection is invalidated', async () => {
    vi.mocked(getMovements).mockResolvedValue([movement]);

    const { result } = renderHook(() => useMovements());
    await waitFor(() => expect(result.current.loading).toBe(false));

    invalidateCollections('movements');

    await waitFor(() => expect(getMovements).toHaveBeenCalledTimes(2));
  });

  it('does not refetch when an unrelated collection is invalidated', async () => {
    vi.mocked(getMovements).mockResolvedValue([movement]);

    const { result } = renderHook(() => useMovements());
    await waitFor(() => expect(result.current.loading).toBe(false));

    invalidateCollections('documents');

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getMovements).toHaveBeenCalledTimes(1);
  });
});
