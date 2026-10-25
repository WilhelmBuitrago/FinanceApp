import { afterEach, describe, expect, it, vi } from 'vitest';

import {
  discardReviewEntry,
  getReviewEntries,
  resolveReviewCategory,
  resolveReviewMovement,
  resolveReviewPeriod,
} from '../../api/review';

afterEach(() => {
  vi.unstubAllGlobals();
});

/* Each review endpoint pins its URL and its exact request body: these are
   the contracts the backend validates per kind, and a drifted body is a
   silent failure (the handler answers 400 and the UI shows a generic
   error). */

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
      return new Response('{}', { status: 200, headers: { 'content-type': 'application/json' } });
    }),
  );
  return calls;
}

describe('api/review', () => {
  it('reads the queue from GET /review/entries', async () => {
    const calls = stubFetch();
    await getReviewEntries();
    expect(calls[0].url).toMatch(/\/review\/entries$/);
    expect(calls[0].method).toBe('GET');
  });

  it('sends the full correction when resolving a movement', async () => {
    const calls = stubFetch();
    await resolveReviewMovement('rv_1', {
      description: 'Compra',
      amount_minor: -1500,
      date: '2026-10-01',
      type: 'EXPENSE',
      account_id: 'acc_1',
    });
    expect(calls[0].url).toMatch(/\/review\/entries\/rv_1\/resolve$/);
    expect(calls[0].method).toBe('POST');
    expect(calls[0].body).toEqual({
      description: 'Compra',
      amount_minor: -1500,
      date: '2026-10-01',
      type: 'EXPENSE',
      account_id: 'acc_1',
    });
  });

  it('sends only the category when resolving a CATEGORY item (null clears it)', async () => {
    const calls = stubFetch();
    await resolveReviewCategory('rv_2', 'cat_1');
    expect(calls[0].body).toEqual({ category_id: 'cat_1' });
    await resolveReviewCategory('rv_2', null);
    expect(calls[1].body).toEqual({ category_id: null });
  });

  it('sends the confirmed month and year for a document period', async () => {
    const calls = stubFetch();
    await resolveReviewPeriod('rv_3', 9, 2026);
    expect(calls[0].url).toMatch(/\/review\/entries\/rv_3\/resolve-period$/);
    expect(calls[0].body).toEqual({ month: 9, year: 2026 });
  });

  it('discards with an empty body', async () => {
    const calls = stubFetch();
    await discardReviewEntry('rv_4');
    expect(calls[0].url).toMatch(/\/review\/entries\/rv_4\/discard$/);
    expect(calls[0].method).toBe('POST');
  });
});
