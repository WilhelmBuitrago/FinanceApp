import { describe, expect, it } from 'vitest';

import { formatDate } from '../../lib/formatDate';

describe('formatDate', () => {
  it('renders an em dash for absent dates', () => {
    expect(formatDate(null)).toBe('—');
    expect(formatDate(undefined)).toBe('—');
    expect(formatDate('')).toBe('—');
  });

  it('formats date-only values in UTC so the written day never shifts', () => {
    // 2026-01-31 parsed as local midnight would show 30 Jan in negative
    // offsets; the UTC read keeps the day.
    expect(formatDate('2026-01-31')).toBe('31 ene 2026');
    expect(formatDate('2026-10-01')).toBe('01 oct 2026');
  });

  it('returns the raw value when it cannot parse', () => {
    expect(formatDate('not-a-date')).toBe('not-a-date');
  });
});
