import { afterEach, describe, expect, it, vi } from 'vitest';

import { ApiError, apiDelete, apiGet } from './client';

afterEach(() => {
  vi.unstubAllGlobals();
});

function jsonResponse(body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json' },
  });
}

describe('fetchJson (through the api helpers)', () => {
  it('returns the parsed body on success', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => jsonResponse({ ok: true }, 200)),
    );
    await expect(apiGet('/x')).resolves.toEqual({ ok: true });
  });

  it('treats 204 as no body even when the server sets content-type', async () => {
    // The trap fetchJson was written for: parsing "" as JSON throws.
    vi.stubGlobal(
      'fetch',
      vi.fn(
        async () =>
          new Response(null, { status: 204, headers: { 'content-type': 'application/json' } }),
      ),
    );
    await expect(apiDelete('/x')).resolves.toBeUndefined();
  });

  it('throws ApiError carrying status and body on failure', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => jsonResponse({ error: 'duplicate' }, 409)),
    );
    await expect(apiGet('/x')).rejects.toMatchObject({
      name: 'ApiError',
      status: 409,
      response: { error: 'duplicate' },
    });
    const err = await apiGet('/x').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
  });

  it('propagates a network failure as-is', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => {
        throw new TypeError('Failed to fetch');
      }),
    );
    await expect(apiGet('/x')).rejects.toThrow('Failed to fetch');
  });
});
