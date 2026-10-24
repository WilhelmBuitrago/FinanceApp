import { describe, expect, it } from 'vitest';

import { APP_VERSION } from './version';

describe('APP_VERSION', () => {
  it('is the package.json version (0.1.1 for this component)', () => {
    expect(APP_VERSION).toBe('0.1.1');
  });
});
