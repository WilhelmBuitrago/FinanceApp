import { afterEach, describe, expect, it, vi } from 'vitest';

import { createMovement, deleteMovement, getMovements, setMovementCategory } from './movements';

afterEach(() => {
  vi.unstubAllGlobals();
});

function stubFetch() {
  const calls: Array<{ url: string; method: string; body: unknown }> = [];
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo, init?: RequestInit) => {
      calls.push({
        url: String(input),
        method: init?.method ?? 'GET',
        body: init?.body ? JSON.parse(String(init.body)) : undefined,
      });
      return new Response(null, { status: 204 });
    }),
  );
  return calls;
}

describe('api/movements', () => {
  it('creates with the exact body the handler validates', async () => {
    const calls = stubFetch();
    await createMovement({
      amount_minor: -2550,
      type: 'EXPENSE',
      account_id: 'acc_1',
      description: 'Mercado',
      date: '2026-10-01',
      category_id: '',
    });
    expect(calls[0].url).toMatch(/\/movements$/);
    expect(calls[0].method).toBe('POST');
    expect(calls[0].body).toEqual({
      amount_minor: -2550,
      type: 'EXPENSE',
      account_id: 'acc_1',
      description: 'Mercado',
      date: '2026-10-01',
      category_id: '',
    });
  });

  it('moves the category with PUT /movements/{id}/category (null clears)', async () => {
    const calls = stubFetch();
    await setMovementCategory('mv_1', 'cat_1');
    expect(calls[0].url).toMatch(/\/movements\/mv_1\/category$/);
    expect(calls[0].method).toBe('PUT');
    expect(calls[0].body).toEqual({ categoryId: 'cat_1' });
    await setMovementCategory('mv_1', null);
    expect(calls[1].body).toEqual({ categoryId: null });
  });

  it('lists and deletes', async () => {
    const calls = stubFetch();
    await getMovements();
    expect(calls[0].url).toMatch(/\/movements$/);
    expect(calls[0].method).toBe('GET');
    await deleteMovement('mv_1');
    expect(calls[1].url).toMatch(/\/movements\/mv_1$/);
    expect(calls[1].method).toBe('DELETE');
  });
});
