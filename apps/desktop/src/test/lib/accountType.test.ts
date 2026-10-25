import { describe, expect, it } from 'vitest';

import { accountTypeLabel } from '../../lib/accountType';

describe('accountTypeLabel', () => {
  it('labels the closed set of account types', () => {
    expect(accountTypeLabel('BANK_ACCOUNT')).toBe('Cuenta bancaria');
    expect(accountTypeLabel('CREDIT_CARD')).toBe('Tarjeta de crédito');
    expect(accountTypeLabel('CASH')).toBe('Efectivo');
  });

  it('keeps the legacy value legible', () => {
    expect(accountTypeLabel('BANK')).toBe('Cuenta bancaria');
  });

  it('falls back to the raw value for an unknown type', () => {
    expect(accountTypeLabel('MYSTERY_TYPE')).toBe('MYSTERY_TYPE');
  });
});
