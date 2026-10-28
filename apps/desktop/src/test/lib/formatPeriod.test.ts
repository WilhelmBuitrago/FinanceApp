import { describe, expect, it } from 'vitest';

import { formatPeriod } from '../../lib/formatPeriod';

describe('formatPeriod', () => {
  it('renders both ends of the statement period', () => {
    expect(formatPeriod('2026-09-01', '2026-09-30')).toBe('01 sept 2026 — 30 sept 2026');
  });

  it('renders the single end it has', () => {
    expect(formatPeriod('2026-09-01', null)).toBe('01 sept 2026');
    expect(formatPeriod(null, '2026-09-30')).toBe('30 sept 2026');
  });

  it('renders an em dash when the period is unknown', () => {
    expect(formatPeriod(null, null)).toBe('—');
    expect(formatPeriod(undefined, undefined)).toBe('—');
  });
});
