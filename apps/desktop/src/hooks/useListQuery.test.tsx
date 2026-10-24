import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { invalidateCollections } from '../lib/dataBus';
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

  it('refetches when a subscribed collection is invalidated', async () => {
    const fetcher = vi.fn(async () => ['a']);
    const { result } = renderHook(() => useListQuery(fetcher, 0, ['movements']));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(fetcher).toHaveBeenCalledTimes(1);
    act(() => {
      invalidateCollections('movements');
    });
    await waitFor(() => expect(fetcher).toHaveBeenCalledTimes(2));
  });

  it('does not refetch when a different collection is invalidated', async () => {
    const fetcher = vi.fn(async () => ['a']);
    renderHook(() => useListQuery(fetcher, 0, ['movements']));
    await waitFor(() => expect(fetcher).toHaveBeenCalledTimes(1));
    act(() => {
      invalidateCollections('categories');
    });
    await new Promise((r) => setTimeout(r, 10));
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it('ignores the bus entirely when no collections are passed (mount only)', async () => {
    const fetcher = vi.fn(async () => ['a']);
    renderHook(() => useListQuery(fetcher));
    await waitFor(() => expect(fetcher).toHaveBeenCalledTimes(1));
    act(() => {
      invalidateCollections('movements', 'accounts');
    });
    await new Promise((r) => setTimeout(r, 10));
    expect(fetcher).toHaveBeenCalledTimes(1);
  });
});
