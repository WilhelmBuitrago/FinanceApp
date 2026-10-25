import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createMovement } from '../../api/movements';
import { getCategories } from '../../api/categories';
import { ApiError } from '../../api/client';
import { invalidateCollections } from '../../lib/dataBus';
import { AddMovementDrawer } from '../../components/AddMovementDrawer';
import type { Account } from '../../types/account';

vi.mock('../../api/movements', () => ({ createMovement: vi.fn() }));
vi.mock('../../api/categories', () => ({ getCategories: vi.fn() }));
vi.mock('../../lib/dataBus', () => ({
  invalidateCollections: vi.fn(),
  useCollectionVersion: vi.fn(() => '0'),
}));

const accounts: Account[] = [
  {
    ID: 'acc_1',
    UserID: 'local',
    Name: 'Nu',
    AccountType: 'BANK_ACCOUNT',
    Currency: 'COP',
    Status: 'ACTIVE',
    CreatedAt: '2026-10-01T00:00:00Z',
    UpdatedAt: '2026-10-01T00:00:00Z',
  },
];

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getCategories).mockResolvedValue([]);
});

function setup(onDuplicate = vi.fn(), onClose = vi.fn()) {
  render(<AddMovementDrawer accounts={accounts} onClose={onClose} onDuplicate={onDuplicate} />);
  return { onDuplicate, onClose };
}

describe('AddMovementDrawer', () => {
  it('refuses a zero amount inline, with the old copy', () => {
    setup();
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);
    expect(screen.getByRole('alert')).toHaveTextContent('Cantidad debe ser positiva');
    expect(createMovement).not.toHaveBeenCalled();
  });

  it('sends the corrected body and republishes the moved data', async () => {
    vi.mocked(createMovement).mockResolvedValue({ id: 'mv_1' });
    setup();
    fireEvent.change(screen.getByLabelText(/Cantidad/), { target: { value: '25.5' } });
    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Egreso' } });
    fireEvent.change(screen.getByLabelText(/Descripción/), {
      target: { value: 'Mercado' },
    });
    fireEvent.change(screen.getByLabelText(/Fecha/), { target: { value: '2026-10-01' } });
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);
    await vi.waitFor(() => expect(createMovement).toHaveBeenCalledTimes(1));
    expect(createMovement).toHaveBeenCalledWith({
      amount_minor: -2550, // Egreso is the negative side
      type: 'EXPENSE',
      account_id: 'acc_1',
      description: 'Mercado',
      date: '2026-10-01',
      category_id: '',
    });
    // The bus carries the invalidation: lists and aggregates refetch, no
    // document reload.
    expect(invalidateCollections).toHaveBeenCalledWith('movements', 'dashboard');
  });

  it('reports a duplicate through onDuplicate and closes, without inventing data', async () => {
    vi.mocked(createMovement).mockRejectedValue(new ApiError('duplicate', 409, {}));
    const { onDuplicate, onClose } = setup();
    fireEvent.change(screen.getByLabelText(/Cantidad/), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText(/Fecha/), { target: { value: '2026-10-01' } });
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);
    await vi.waitFor(() => expect(onDuplicate).toHaveBeenCalledTimes(1));
    expect(onDuplicate.mock.calls[0][0]).toContain('Quedó en Revisión');
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('shows the server body when the create fails', async () => {
    vi.mocked(createMovement).mockRejectedValue(
      new ApiError('bad request', 400, '{"error":"date is required"}'),
    );
    setup();
    fireEvent.change(screen.getByLabelText(/Cantidad/), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText(/Fecha/), { target: { value: '2026-10-01' } });
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);
    await vi.waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    expect(screen.getByRole('alert')).toHaveTextContent('date is required');
  });
});
