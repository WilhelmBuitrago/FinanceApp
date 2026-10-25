import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../api/client';
import { getDocument } from '../../api/documents';
import { invalidateCollections } from '../../lib/dataBus';
import { useResource } from '../../hooks/useResource';

vi.mock('../../api/documents', () => ({ getDocument: vi.fn() }));

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useResource', () => {
  it('stays idle without an id', () => {
    const { result } = renderHook(() => useResource(null, getDocument));
    expect(result.current.status).toBe('idle');
  });

  it('moves to ready with the fetched data', async () => {
    vi.mocked(getDocument).mockResolvedValue({ ID: 'doc_1' } as never);
    const { result } = renderHook(() => useResource('doc_1', getDocument));
    await waitFor(() => expect(result.current.status).toBe('ready'));
    expect(result.current.status === 'ready' && result.current.data).toEqual({ ID: 'doc_1' });
  });

  it('exposes the HTTP status code on failure (404 is honest absence)', async () => {
    vi.mocked(getDocument).mockRejectedValue(new ApiError('gone', 404));
    const { result } = renderHook(() => useResource('doc_x', getDocument));
    await waitFor(() => expect(result.current.status).toBe('error'));
    expect(result.current.status === 'error' && result.current.code).toBe(404);
  });

  it('retries on demand', async () => {
    vi.mocked(getDocument).mockRejectedValueOnce(new ApiError('boom', 500));
    vi.mocked(getDocument).mockResolvedValueOnce({ ID: 'doc_2' } as never);
    const { result } = renderHook(() => useResource('doc_2', getDocument));
    await waitFor(() => expect(result.current.status).toBe('error'));
    result.current.retry();
    await waitFor(() => expect(result.current.status).toBe('ready'));
    expect(getDocument).toHaveBeenCalledTimes(2);
  });

  it('refetches when a subscribed collection is invalidated', async () => {
    vi.mocked(getDocument).mockResolvedValue({ ID: 'doc_1' } as never);
    const { result } = renderHook(() => useResource('doc_1', getDocument, ['movements']));
    await waitFor(() => expect(result.current.status).toBe('ready'));
    invalidateCollections('movements');
    await waitFor(() => expect(getDocument).toHaveBeenCalledTimes(2));
  });

  it('does not refetch for a collection it does not subscribe to', async () => {
    vi.mocked(getDocument).mockResolvedValue({ ID: 'doc_1' } as never);
    renderHook(() => useResource('doc_1', getDocument, ['movements']));
    await waitFor(() => expect(getDocument).toHaveBeenCalledTimes(1));
    invalidateCollections('categories');
    await new Promise((r) => setTimeout(r, 10));
    expect(getDocument).toHaveBeenCalledTimes(1);
  });
});
