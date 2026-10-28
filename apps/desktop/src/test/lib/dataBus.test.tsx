import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { invalidateCollections, useCollectionVersion } from '../../lib/dataBus';

/* The bus is shared module state: every assertion compares before/after,
   never absolute values, because versions accumulate across tests. */

describe('useCollectionVersion', () => {
  it('changes only for the invalidated collection', () => {
    const { result } = renderHook(() => ({
      movements: useCollectionVersion('movements'),
      accounts: useCollectionVersion('accounts'),
    }));
    const before = result.current;
    act(() => {
      invalidateCollections('movements');
    });
    expect(result.current.movements).not.toBe(before.movements);
    expect(result.current.accounts).toBe(before.accounts);
  });

  it('changes when any of the subscribed collections is invalidated', () => {
    const { result } = renderHook(() => ({
      both: useCollectionVersion('movements', 'dashboard'),
    }));
    const before = result.current.both;
    act(() => {
      invalidateCollections('dashboard');
    });
    expect(result.current.both).not.toBe(before);
  });

  it('gives the same stamp to two hooks subscribed to the same collections', () => {
    const one = renderHook(() => useCollectionVersion('review'));
    const two = renderHook(() => useCollectionVersion('review'));
    expect(one.result.current).toBe(two.result.current);
    act(() => {
      invalidateCollections('review');
    });
    expect(one.result.current).toBe(two.result.current);
    expect(one.result.current).not.toBe(0);
  });
});

describe('invalidateCollections', () => {
  it('ignores an empty call: no listener runs for nothing', () => {
    const { result } = renderHook(() => useCollectionVersion('categories'));
    const before = result.current;
    act(() => {
      invalidateCollections();
    });
    expect(result.current).toBe(before);
  });

  it('reaches every mounted subscriber at once (two hooks, one publish)', () => {
    const first = renderHook(() => useCollectionVersion('documents'));
    const second = renderHook(() => useCollectionVersion('documents'));
    const beforeFirst = first.result.current;
    const beforeSecond = second.result.current;
    act(() => {
      invalidateCollections('documents');
    });
    expect(first.result.current).not.toBe(beforeFirst);
    expect(second.result.current).not.toBe(beforeSecond);
  });
});
