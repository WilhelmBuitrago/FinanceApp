/* Spanish labels for AccountType (closed set: DB CHECK BANK_ACCOUNT/CASH/CREDIT_CARD). */

export const ACCOUNT_TYPE_LABELS: Record<string, string> = {
  BANK_ACCOUNT: 'Cuenta bancaria',
  CASH: 'Efectivo',
  CREDIT_CARD: 'Tarjeta de crédito',
  BANK: 'Cuenta bancaria', // legacy value from pre-S12 data
};

export function accountTypeLabel(accountType: string): string {
  return ACCOUNT_TYPE_LABELS[accountType] ?? accountType;
}
