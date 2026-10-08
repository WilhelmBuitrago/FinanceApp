import { apiGet, apiPost } from './client';
import type { Account } from '../types/account';
import type { Movement } from '../types/movement';
import type { Document } from '../types/document';

export async function getAccounts(): Promise<Account[]> {
  return apiGet<Account[]>('/accounts');
}

/* POST /accounts contract (verified against backend handler):
   client provides the id (UUID), user_id and body fields in snake_case;
   the response is the created domain account (PascalCase). */
export interface CreateAccountPayload {
  id: string;
  user_id: string;
  name: string;
  account_type: 'BANK_ACCOUNT' | 'CREDIT_CARD';
  currency: string;
  institution: string;
  validation_data?: string;
}

export async function createAccount(payload: CreateAccountPayload): Promise<Account> {
  return apiPost<Account>('/accounts', payload);
}

export async function getAccount(id: string): Promise<Account> {
  return apiGet<Account>(`/accounts/${id}`);
}

export async function getAccountMovements(id: string): Promise<Movement[]> {
  return apiGet<Movement[]>(`/accounts/${id}/movements`);
}

export async function getAccountDocuments(id: string): Promise<Document[]> {
  return apiGet<Document[]>(`/accounts/${id}/documents`);
}
