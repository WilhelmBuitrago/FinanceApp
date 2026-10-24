import { describe, expect, it } from 'vitest';

import { progressLabel } from './updateProgress';

describe('progressLabel', () => {
  it('shows the percentage when the plugin reported a total', () => {
    expect(progressLabel(500, 1000)).toBe('50%');
    expect(progressLabel(1000, 1000)).toBe('100%');
    expect(progressLabel(1500, 1000)).toBe('150%');
  });

  it('falls back to kilobytes when the total is unknown', () => {
    expect(progressLabel(2048, 0)).toBe('2 KB');
    expect(progressLabel(0, 0)).toBe('0 KB');
  });
});
