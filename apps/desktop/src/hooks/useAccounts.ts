import { getAccounts } from '../api/accounts';
import type { Account } from '../types/account';
import { useListQuery } from './useListQuery';

export function useAccounts(refreshKey = 0) {
  return useListQuery<Account[]>(getAccounts, refreshKey, ['accounts']);
}
