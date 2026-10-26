import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { getCategories } from '../../api/categories';
import { deleteMovement, setMovementCategory } from '../../api/movements';
import { invalidateCollections } from '../../lib/dataBus';
import { useMovementActions } from '../../hooks/useMovementActions';
import type { Movement } from '../../types/movement';

/* The traceability drawer's action machine: category assignment (it lives
   on the financial event behind the movement) and confirmed deletion.
   Failure text and bus invalidations are the contract; a missing
   invalidation leaves every other screen stale. */

vi.mock('../../api/categories', () => ({ getCategories: vi.fn() }));
vi.mock('../../api/movements', () => ({ setMovementCategory: vi.fn(), deleteMovement: vi.fn() }));
vi.mock('../../lib/dataBus', () => ({ invalidateCollections: vi.fn() }));

const movement: Movement = {
  ID: 'mv-1',
  FinancialEventID: 'fe-1',
  AccountID: 'acc-1',
  MovementDate: '2026-10-20',
  AmountMinor: -1500,
  Description: 'Mercado',
  Status: 'ACTIVE',
  categoryName: '',
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
};

const categories = [
  {
    ID: 'cat-food',
    Name: 'Comida',
    CategoryType: 'EXPENSE',
    Status: 'ACTIVE',
    CreatedAt: '',
    UpdatedAt: '',
  },
  {
    ID: 'cat-pay',
    Name: 'Sueldo',
    CategoryType: 'INCOME',
    Status: 'ACTIVE',
    CreatedAt: '',
    UpdatedAt: '',
  },
];

function setup(overrides: { movement?: Movement | null; eventType?: string } = {}) {
  const onClose = vi.fn();
  const view = renderHook(() =>
    useMovementActions(
      overrides.movement === undefined ? movement : overrides.movement,
      overrides.eventType ?? 'EXPENSE',
      onClose,
    ),
  );
  return { ...view, onClose };
}

describe('useMovementActions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getCategories).mockResolvedValue(categories);
    vi.mocked(setMovementCategory).mockResolvedValue(undefined);
    vi.mocked(deleteMovement).mockResolvedValue(undefined);
  });

  it('loads the categories and keeps only the event type ones', async () => {
    const { result } = setup();

    await act(async () => {});
    expect(result.current.categories.map((c) => c.ID)).toEqual(['cat-food']);
  });

  it('keeps every category while the event type is unknown', async () => {
    const { result } = setup({ eventType: '' });

    await act(async () => {});
    expect(result.current.categories).toHaveLength(2);
  });

  it('assigns a category, invalidates the movements collection and leaves no error', async () => {
    const { result } = setup();
    await act(async () => {});

    await act(async () => {
      await result.current.assignCategory('cat-food');
    });

    expect(setMovementCategory).toHaveBeenCalledWith('mv-1', 'cat-food');
    expect(invalidateCollections).toHaveBeenCalledWith('movements');
    expect(result.current.categoryError).toBeNull();
    expect(result.current.savingCategory).toBe(false);
  });

  it('assigns null (uncategorized) without inventing a category', async () => {
    const { result } = setup();
    await act(async () => {});

    await act(async () => {
      await result.current.assignCategory(null);
    });

    expect(setMovementCategory).toHaveBeenCalledWith('mv-1', null);
  });

  it('reports a failed assignment and does not invalidate', async () => {
    vi.mocked(setMovementCategory).mockRejectedValue(new Error('nope'));
    const { result } = setup();
    await act(async () => {});

    await act(async () => {
      await result.current.assignCategory('cat-food');
    });

    expect(result.current.categoryError).toBe('No se pudo guardar la categoría.');
    expect(invalidateCollections).not.toHaveBeenCalled();
  });

  it('confirms deletion through the modal, closes the drawer and invalidates both collections', async () => {
    const { result, onClose } = setup();
    await act(async () => {});

    act(() => result.current.requestDelete());
    expect(result.current.confirmingDelete).toBe(true);

    act(() => result.current.cancelDelete());
    expect(result.current.confirmingDelete).toBe(false);

    act(() => result.current.requestDelete());
    await act(async () => {
      await result.current.confirmDelete();
    });

    expect(deleteMovement).toHaveBeenCalledWith('mv-1');
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(invalidateCollections).toHaveBeenCalledWith('movements', 'dashboard');
    expect(result.current.confirmingDelete).toBe(false);
  });

  it('keeps the drawer open and reports a failed deletion', async () => {
    vi.mocked(deleteMovement).mockRejectedValue(new Error('nope'));
    const { result, onClose } = setup();
    await act(async () => {});

    act(() => result.current.requestDelete());
    await act(async () => {
      await result.current.confirmDelete();
    });

    expect(result.current.deleteError).toBe('No se pudo eliminar el movimiento');
    expect(result.current.confirmingDelete).toBe(false);
    expect(onClose).not.toHaveBeenCalled();
    expect(invalidateCollections).not.toHaveBeenCalled();
  });

  it('does nothing when there is no movement to act on', async () => {
    const { result } = setup({ movement: null });
    await act(async () => {});

    await act(async () => {
      await result.current.assignCategory('cat-food');
      await result.current.confirmDelete();
    });

    expect(setMovementCategory).not.toHaveBeenCalled();
    expect(deleteMovement).not.toHaveBeenCalled();
  });
});
