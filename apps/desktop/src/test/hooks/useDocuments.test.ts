import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { invalidateCollections } from '../../lib/dataBus';
import { getDocuments } from '../../api/documents';
import { useDocuments } from '../../hooks/useDocuments';

/* The documents list hook: wiring over useListQuery — the endpoint it
   calls, the state it exposes and the collection that refreshes it. */

vi.mock('../../api/documents', () => ({ getDocuments: vi.fn() }));

const document = {
  ID: 'doc-1',
  AccountID: 'acc-1',
  DocumentType: 'PDF',
  Institution: 'NU',
  DocumentName: 'extracto.pdf',
  Status: 'IMPORTED',
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
};

describe('useDocuments', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads the documents and exposes the loaded state', async () => {
    vi.mocked(getDocuments).mockResolvedValue([document]);

    const { result } = renderHook(() => useDocuments());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data?.[0].DocumentName).toBe('extracto.pdf');
    expect(getDocuments).toHaveBeenCalledTimes(1);
  });

  it('surfaces a failed load as the error state', async () => {
    vi.mocked(getDocuments).mockRejectedValue(new Error('backend down'));

    const { result } = renderHook(() => useDocuments());

    await waitFor(() => expect(result.current.error).toBeInstanceOf(Error));
    expect(result.current.data).toBeUndefined();
  });

  it('refetches when the documents collection is invalidated', async () => {
    vi.mocked(getDocuments).mockResolvedValue([document]);

    const { result } = renderHook(() => useDocuments());
    await waitFor(() => expect(result.current.loading).toBe(false));

    invalidateCollections('documents');

    await waitFor(() => expect(getDocuments).toHaveBeenCalledTimes(2));
  });
});
