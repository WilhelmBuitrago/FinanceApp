import { renderHook, waitFor } from '@testing-library/react';
import { act } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { getReviewEntries } from '../api/review';
import { invalidateCollections } from '../lib/dataBus';
import { useReviewNotifications } from './useReviewNotifications';

vi.mock('../api/review', () => ({ getReviewEntries: vi.fn() }));

const item = (kind: string, payload: object) => ({
  ID: `rv_${kind}`,
  Kind: kind,
  Status: 'PENDING',
  CreatedAt: '2026-10-24T00:00:00Z',
  PayloadJson: JSON.stringify(payload),
});

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getReviewEntries).mockResolvedValue([]);
});

describe('useReviewNotifications', () => {
  it('turns every queue trigger into its honest line', async () => {
    vi.mocked(getReviewEntries).mockResolvedValue([
      item('CATEGORY', {}) as never,
      item('DUPLICATE', {}) as never,
      item('EXTRACTION', { reason: 'sin fecha' }) as never,
    ]);
    const { result } = renderHook(() => useReviewNotifications());
    await waitFor(() => expect(result.current.notifications).toHaveLength(3));
    expect(result.current.notifications[0]).toMatch(/Movimiento sin categoría/);
    expect(result.current.notifications[1]).toMatch(/Movimiento duplicado/);
    expect(result.current.notifications[2]).toBe('Extracción requiere verificación (sin fecha)');
  });

  it('pushes a mutation line on top of the queue', async () => {
    const { result } = renderHook(() => useReviewNotifications());
    await waitFor(() => expect(getReviewEntries).toHaveBeenCalled());
    act(() => {
      result.current.notify('Movimiento duplicado detectado el 2026-10-01. Quedó en Revisión.');
    });
    expect(result.current.notifications).toHaveLength(1);
  });

  it('reloads the queue when anything invalidates review', async () => {
    const { result } = renderHook(() => useReviewNotifications());
    await waitFor(() => expect(getReviewEntries).toHaveBeenCalledTimes(1));
    act(() => {
      invalidateCollections('review');
    });
    await waitFor(() => expect(getReviewEntries).toHaveBeenCalledTimes(2));
    expect(result.current.notifications).toEqual([]);
  });
});
