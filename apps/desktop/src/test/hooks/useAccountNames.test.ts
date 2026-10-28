import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useAccountNames } from '../../hooks/useAccountNames';
import type { UseQueryState } from '../../types/api';
import type { Account } from '../../types/account';

/* The id -> name map the movement and document screens share. Pure
   derivation over the accounts query: no fetches, no subscriptions. */

function accountsState(data: Account[] | undefined): UseQueryState<Account[]> {
  if (data === undefined) return { loading: true };
  return { loading: false, data, empty: data.length === 0 };
}

const account = (id: string, name: string): Account => ({
  ID: id,
  UserID: 'u',
  Name: name,
  AccountType: 'BANK_ACCOUNT',
  Currency: 'COP',
  Status: 'ACTIVE',
  Institution: 'NU',
  CreatedAt: '2026-10-01T00:00:00Z',
  UpdatedAt: '2026-10-01T00:00:00Z',
});

describe('useAccountNames', () => {
  it('maps every account id to its name', () => {
    const { result } = renderHook(() =>
      useAccountNames(accountsState([account('acc-1', 'Nu'), account('acc-2', 'Bancolombia')])),
    );

    expect(result.current.get('acc-1')).toBe('Nu');
    expect(result.current.get('acc-2')).toBe('Bancolombia');
    expect(result.current.size).toBe(2);
  });

  it('is empty while the accounts query has no data', () => {
    const { result } = renderHook(() => useAccountNames(accountsState(undefined)));
    expect(result.current.size).toBe(0);
  });

  it('recomputes when the accounts data changes', () => {
    const { result, rerender } = renderHook(
      ({ data }: { data: Account[] | undefined }) => useAccountNames(accountsState(data)),
      { initialProps: { data: [account('acc-1', 'Nu')] as Account[] | undefined } },
    );
    expect(result.current.get('acc-1')).toBe('Nu');

    rerender({ data: [account('acc-1', 'Nu renombrada')] });
    expect(result.current.get('acc-1')).toBe('Nu renombrada');

    rerender({ data: undefined });
    expect(result.current.size).toBe(0);
  });
});
