import { describe, expect, it } from 'vitest';

import { formatMoney, formatSignedMoney, signedAmountClass } from '../../lib/formatMoney';

describe('formatMoney', () => {
  it('renders minor units as pesos with two decimals', () => {
    // 150042 minor -> $1.500,42 (es-CO).
    expect(formatMoney(150042)).toBe('$1.500,42');
    expect(formatMoney(15044)).toBe('$150,44');
  });

  it('keeps the sign and never drops the fraction for small amounts', () => {
    expect(formatMoney(-2500)).toBe('-$25,00');
    expect(formatMoney(5)).toBe('$0,05');
  });

  it('formats a known non-COP currency with its own symbol', () => {
    // Until multi-currency arrives, COP keeps the hand-rolled "$"; anything
    // the catalog admits (USD and friends) formats through Intl so the
    // symbol and separators stop lying about the money.
    expect(formatMoney(150000, 'USD')).not.toBe(formatMoney(150000, 'COP'));
    expect(formatMoney(150000, 'USD')).toMatch(/^\+?-?US\$\s?1\.500,00$/);
  });
});

describe('formatSignedMoney', () => {
  it('prefixes only positives, zero stays unsigned', () => {
    expect(formatSignedMoney(4100000)).toBe('+$41.000,00');
    expect(formatSignedMoney(-1250000)).toBe('-$12.500,00');
    expect(formatSignedMoney(0)).toBe('$0,00');
  });
});

describe('signedAmountClass', () => {
  it('maps the sign to the financial semantic colors', () => {
    expect(signedAmountClass(4100000)).toBe('text-pos');
    expect(signedAmountClass(-1250000)).toBe('text-neg');
    expect(signedAmountClass(0)).toBe('text-ink');
  });
});
