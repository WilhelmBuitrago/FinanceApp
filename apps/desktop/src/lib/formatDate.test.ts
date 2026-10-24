import { describe, expect, it } from 'vitest';

import { formatDate, formatDateTime } from './formatDate';

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

describe('formatDateTime', () => {
  it('includes the time for full timestamps', () => {
    // 15:04 UTC, rendered without a timezone suffix.
    expect(formatDateTime('2026-10-01T15:04:00Z')).toMatch(/^01 oct 2026, \d{2}:\d{2}$/);
  });

  it('renders an em dash for absent timestamps', () => {
    expect(formatDateTime(null)).toBe('—');
  });
});
