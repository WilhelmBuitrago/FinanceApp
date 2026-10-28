import { describe, expect, it } from 'vitest';

import { parseAmountMinor } from '../../lib/amount';

/* The drawer's amount field is free text: a Colombian writes "25,5"
   and parseFloat used to read 25, saving 2500 centavos for a 2550
   movement with no complaint anywhere. The parser must read both
   conventions and refuse what it cannot read honestly. */

describe('parseAmountMinor', () => {
  it('reads the comma as the decimal separator', () => {
    expect(parseAmountMinor('25,5')).toBe(2550);
    expect(parseAmountMinor('0,05')).toBe(5);
  });

  it('reads the dot as decimals when one or two digits follow', () => {
    expect(parseAmountMinor('25.5')).toBe(2550);
    expect(parseAmountMinor('25.55')).toBe(2555);
  });

  it('reads a three-digit group after a separator as thousands', () => {
    expect(parseAmountMinor('25.500')).toBe(2550000);
    expect(parseAmountMinor('1.250.000')).toBe(125000000);
    expect(parseAmountMinor('25,567')).toBe(2556700);
  });

  it('reads a mixed amount with the last separator as decimals', () => {
    expect(parseAmountMinor('1.250.000,5')).toBe(125000050);
    expect(parseAmountMinor('1.250.000,50')).toBe(125000050);
  });

  it('reads a bare number as whole pesos', () => {
    expect(parseAmountMinor('2500')).toBe(250000);
    expect(parseAmountMinor('0')).toBe(0);
  });

  it('returns null for what it cannot read', () => {
    expect(parseAmountMinor('')).toBeNull();
    expect(parseAmountMinor('   ')).toBeNull();
    expect(parseAmountMinor('abc')).toBeNull();
    expect(parseAmountMinor('-5')).toBeNull();
    expect(parseAmountMinor('12,34,56')).toBeNull();
    expect(parseAmountMinor('1 500')).toBeNull();
    expect(parseAmountMinor('1.2.3.4.5')).toBeNull();
  });
});
