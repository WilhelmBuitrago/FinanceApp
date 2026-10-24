import { defineConfig } from 'vitest/config';

// Component and unit tests. The API layer is stubbed per test; nothing
// here talks to the real backend.
export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
  },
});
