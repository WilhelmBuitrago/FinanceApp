import { afterEach, describe, expect, it, vi } from 'vitest';

import { createCategory, deleteCategory, getCategories } from './categories';

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

describe('api/categories', () => {
  it('lists the active categories', async () => {
    const calls = stubFetch();
    await getCategories();
    expect(calls[0].url).toMatch(/\/categories$/);
    expect(calls[0].method).toBe('GET');
  });

  it('creates with the body the handler expects', async () => {
    const calls = stubFetch();
    await createCategory('cat_1', 'Supermercado', 'EXPENSE');
    expect(calls[0].body).toEqual({ id: 'cat_1', name: 'Supermercado', category_type: 'EXPENSE' });
  });

  it('deletes by id', async () => {
    const calls = stubFetch();
    await deleteCategory('cat_1');
    expect(calls[0].url).toMatch(/\/categories\/cat_1$/);
    expect(calls[0].method).toBe('DELETE');
  });
});
