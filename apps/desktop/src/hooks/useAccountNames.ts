import { useMemo } from 'react';

import type { UseQueryState } from '../types/api';
import type { Account } from '../types/account';

/* Account id -> name map for the screens that render movements and
   documents. Three pages carried the same memo before the audit. */

export function useAccountNames(accounts: UseQueryState<Account[]>): Map<string, string> {
  return useMemo(() => {
    const map = new Map<string, string>();
    for (const account of accounts.data ?? []) map.set(account.ID, account.Name);
    return map;
  }, [accounts.data]);
}
