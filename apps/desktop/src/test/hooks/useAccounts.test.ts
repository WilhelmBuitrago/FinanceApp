import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { invalidateCollections } from '../../lib/dataBus';
import { getAccounts } from '../../api/accounts';
import { useAccounts } from '../../hooks/useAccounts';

/* The accounts list hook: wiring over useListQuery. The primitive is
   tested on its own; what this pins is the endpoint it calls, the state
   it exposes, and the collection it subscribes to — a wrong collection
   silently stops refreshing every screen that shows accounts. */

vi.mock('../../api/accounts', () => ({ getAccounts: vi.fn() }));

describe('useAccounts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads the accounts and exposes the loaded state', async () => {
    vi.mocked(getAccounts).mockResolvedValue([
      {
        ID: 'acc-1',
        UserID: 'u',
        Name: 'Nu',
        AccountType: 'BANK_ACCOUNT',
        Currency: 'COP',
        Status: 'ACTIVE',
        Institution: 'NU',
        CreatedAt: '',
        UpdatedAt: '',
      },
    ]);

    const { result } = renderHook(() => useAccounts());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0].Name).toBe('Nu');
    expect(result.current.empty).toBe(false);
    expect(getAccounts).toHaveBeenCalledTimes(1);
  });

  it('reports an empty list as empty, not as an error', async () => {
    vi.mocked(getAccounts).mockResolvedValue([]);

    const { result } = renderHook(() => useAccounts());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.empty).toBe(true);
    expect(result.current.error).toBeUndefined();
  });

  it('surfaces a failed load as the error state', async () => {
    vi.mocked(getAccounts).mockRejectedValue(new Error('backend down'));

    const { result } = renderHook(() => useAccounts());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('backend down');
    expect(result.current.data).toBeUndefined();
  });

  it('refetches when the accounts collection is invalidated', async () => {
    vi.mocked(getAccounts).mockResolvedValue([]);

    const { result } = renderHook(() => useAccounts());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getAccounts).toHaveBeenCalledTimes(1);

    invalidateCollections('accounts');

    await waitFor(() => expect(getAccounts).toHaveBeenCalledTimes(2));
  });
});
