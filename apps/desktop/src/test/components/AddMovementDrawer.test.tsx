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

/* Transfers: the destination field only exists for them, the amount
   stays positive (the backend derives the signs), and a transfer
   carries no category. */

const twoAccounts: Account[] = [
  {
    ID: 'acc_from',
    UserID: 'local',
    Name: 'Nu',
    AccountType: 'BANK_ACCOUNT',
    Currency: 'COP',
    Status: 'ACTIVE',
    CreatedAt: '2026-10-01T00:00:00Z',
    UpdatedAt: '2026-10-01T00:00:00Z',
  },
  {
    ID: 'acc_to',
    UserID: 'local',
    Name: 'Ahorros',
    AccountType: 'BANK_ACCOUNT',
    Currency: 'COP',
    Status: 'ACTIVE',
    CreatedAt: '2026-10-01T00:00:00Z',
    UpdatedAt: '2026-10-01T00:00:00Z',
  },
];

function setupTransfer(onDuplicate = vi.fn(), onClose = vi.fn()) {
  render(<AddMovementDrawer accounts={twoAccounts} onClose={onClose} onDuplicate={onDuplicate} />);
  return { onDuplicate, onClose };
}

describe('AddMovementDrawer transfers', () => {
  it('offers the destination field only for a transfer', () => {
    setupTransfer();

    expect(screen.queryByLabelText(/Cuenta destino/)).toBeNull();
    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Transferencia' } });
    expect(screen.getByLabelText(/Cuenta destino/)).toBeInTheDocument();
  });

  it('hides the category for a transfer: a transfer has none', () => {
    setupTransfer();

    expect(screen.getByLabelText(/Categoría/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Transferencia' } });
    expect(screen.queryByLabelText(/Categoría/)).toBeNull();
  });

  it('never offers the origin account as its own destination', () => {
    setupTransfer();

    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Transferencia' } });
    const options = screen.getByLabelText(/Cuenta destino/).querySelectorAll('option');
    const values = Array.from(options).map((o) => (o as HTMLOptionElement).value);
    expect(values).toContain('acc_to');
    expect(values).not.toContain('acc_from');
  });

  it('sends the positive amount, the destination and no category', async () => {
    vi.mocked(createMovement).mockResolvedValue({ id: 'mv_out', counterpart_id: 'mv_in' });
    setupTransfer();

    fireEvent.change(screen.getByLabelText(/Cantidad/), { target: { value: '25.5' } });
    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Transferencia' } });
    fireEvent.change(screen.getByLabelText(/Cuenta destino/), { target: { value: 'acc_to' } });
    fireEvent.change(screen.getByLabelText(/Descripción/), { target: { value: 'Alquiler' } });
    fireEvent.change(screen.getByLabelText(/Fecha/), { target: { value: '2026-10-01' } });
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);

    await vi.waitFor(() => expect(createMovement).toHaveBeenCalledTimes(1));
    expect(createMovement).toHaveBeenCalledWith({
      amount_minor: 2550, // positive: the signs are derived server-side
      type: 'TRANSFER',
      account_id: 'acc_from',
      to_account_id: 'acc_to',
      description: 'Alquiler',
      date: '2026-10-01',
      category_id: '',
    });
    expect(invalidateCollections).toHaveBeenCalledWith('movements', 'dashboard');
  });

  it('refuses a transfer without destination before calling the API', () => {
    setupTransfer();

    fireEvent.change(screen.getByLabelText(/Cantidad/), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Transferencia' } });
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);

    expect(screen.getByRole('alert')).toHaveTextContent('Elegí la cuenta destino');
    expect(createMovement).not.toHaveBeenCalled();
  });

  it('reports a server-rejected transfer without inventing data', async () => {
    vi.mocked(createMovement).mockRejectedValue(
      new ApiError('bad request', 400, '{"error":"a transfer needs two different accounts"}'),
    );
    setupTransfer();

    fireEvent.change(screen.getByLabelText(/Cantidad/), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText(/Tipo/), { target: { value: 'Transferencia' } });
    fireEvent.change(screen.getByLabelText(/Cuenta destino/), { target: { value: 'acc_to' } });
    fireEvent.change(screen.getByLabelText(/Fecha/), { target: { value: '2026-10-01' } });
    fireEvent.submit(screen.getByRole('button', { name: 'Aceptar' }).closest('form')!);

    await vi.waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    expect(screen.getByRole('alert')).toHaveTextContent('a transfer needs two different accounts');
  });
});
