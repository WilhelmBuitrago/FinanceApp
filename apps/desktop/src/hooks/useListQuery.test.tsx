import { renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { useListQuery } from './useListQuery';

describe('useListQuery', () => {
  it('moves from loading to ready with data and the empty flag', async () => {
    const fetcher = vi.fn(async () => [1, 2, 3]);
    const { result } = renderHook(() => useListQuery(fetcher));
    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data).toEqual([1, 2, 3]);
    expect(result.current.empty).toBe(false);
  });

  it('marks an empty list as empty', async () => {
    const { result } = renderHook(() => useListQuery(async () => []));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.empty).toBe(true);
  });

  it('exposes the failure as an Error instance', async () => {
    const { result } = renderHook(() =>
      useListQuery(async () => {
        throw 'network down'; // a non-Error rejection still becomes an Error
      }),
    );
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it('refetches when the refreshKey changes', async () => {
    const fetcher = vi.fn(async () => ['a']);
    const { result, rerender } = renderHook(({ key }) => useListQuery(fetcher, key), {
      initialProps: { key: 0 },
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    rerender({ key: 1 });
    await waitFor(() => expect(fetcher).toHaveBeenCalledTimes(2));
  });
});
