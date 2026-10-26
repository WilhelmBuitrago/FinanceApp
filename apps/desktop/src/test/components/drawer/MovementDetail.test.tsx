import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import { MovementDetail } from '../../../components/drawer/MovementDetail';
import type { Category } from '../../../types/category';
import type { Movement } from '../../../types/movement';

/* The movement section of the traceability drawer: local data always
   present, category assignment through the machine's props, and
   deletion behind a confirmation. */

const movement: Movement = {
  ID: 'mv_abcdefghijklm',
  FinancialEventID: 'fe-1',
  AccountID: 'acc-1',
  MovementDate: '2026-10-20',
  AmountMinor: -1500,
  Description: 'Mercado',
  Status: 'ACTIVE',
  categoryName: 'Comida',
  categoryId: 'cat-food',
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
};

const categories: Category[] = [
  {
    ID: 'cat-food',
    Name: 'Comida',
    CategoryType: 'EXPENSE',
    Status: 'ACTIVE',
    CreatedAt: '',
    UpdatedAt: '',
  },
  {
    ID: 'cat-bus',
    Name: 'Bus',
    CategoryType: 'EXPENSE',
    Status: 'ACTIVE',
    CreatedAt: '',
    UpdatedAt: '',
  },
];

function setup(overrides: Partial<Parameters<typeof MovementDetail>[0]> = {}) {
  const handlers = {
    onAssignCategory: vi.fn(),
    onRequestDelete: vi.fn(),
    onCancelDelete: vi.fn(),
    onConfirmDelete: vi.fn(),
  };
  const props = {
    movement,
    accountName: 'Nu',
    categories,
    savingCategory: false,
    categoryError: null,
    onAssignCategory: handlers.onAssignCategory,
    confirmingDelete: false,
    deleting: false,
    deleteError: null,
    onRequestDelete: handlers.onRequestDelete,
    onCancelDelete: handlers.onCancelDelete,
    onConfirmDelete: handlers.onConfirmDelete,
    ...overrides,
  };
  render(
    <MemoryRouter>
      <MovementDetail {...props} />
    </MemoryRouter>,
  );
  return { ...props, ...handlers };
}

describe('MovementDetail', () => {
  it('renders the local data: date, account, status and a shortened id', () => {
    setup();

    expect(screen.getByText('20 oct 2026')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Nu' }).getAttribute('href')).toBe('/accounts/acc-1');
    expect(screen.getByText('Activo')).toBeTruthy();
    expect(screen.getByTitle('Copiar ID').textContent).toBe('mv_abcdefghij…');
  });

  it('falls back to the account id when the name is unknown', () => {
    setup({ accountName: null });

    expect(screen.getByRole('link', { name: 'acc-1' })).toBeTruthy();
  });

  it('offers the categories in the select and assigns the picked one', () => {
    const { onAssignCategory } = setup();

    fireEvent.change(screen.getByLabelText('Categoría del movimiento'), {
      target: { value: 'cat-bus' },
    });

    expect(onAssignCategory).toHaveBeenCalledWith('cat-bus');
  });

  it('assigns null when the "no category" option is picked', () => {
    const { onAssignCategory } = setup();

    fireEvent.change(screen.getByLabelText('Categoría del movimiento'), {
      target: { value: '' },
    });

    expect(onAssignCategory).toHaveBeenCalledWith(null);
  });

  it('disables the select while the assignment is being saved', () => {
    setup({ savingCategory: true });

    expect(screen.getByLabelText('Categoría del movimiento')).toHaveProperty('disabled', true);
  });

  it('replaces the select with the error when the assignment failed', () => {
    setup({ categoryError: 'No se pudo guardar la categoría.' });

    expect(screen.queryByLabelText('Categoría del movimiento')).toBeNull();
    expect(screen.getByText('No se pudo guardar la categoría.')).toBeTruthy();
  });

  it('asks for confirmation before deleting: the button requests, the modal does not exist yet', () => {
    const { onRequestDelete } = setup();

    fireEvent.click(screen.getByRole('button', { name: 'Eliminar movimiento' }));

    expect(onRequestDelete).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });

  it('cancels the confirmation without deleting', () => {
    const { onCancelDelete, onConfirmDelete } = setup({ confirmingDelete: true });

    fireEvent.click(screen.getByRole('button', { name: 'Cancelar' }));

    expect(onCancelDelete).toHaveBeenCalledTimes(1);
    expect(onConfirmDelete).not.toHaveBeenCalled();
  });

  it('deletes only when the modal confirms, and shows the busy state', () => {
    setup({ confirmingDelete: true, deleting: true });

    const confirm = screen.getByRole('button', { name: 'Borrando…' });
    expect(confirm).toHaveProperty('disabled', true);
  });

  it('calls the confirm handler when the modal confirms', () => {
    const { onConfirmDelete } = setup({ confirmingDelete: true });

    fireEvent.click(screen.getByRole('button', { name: 'Eliminar' }));

    expect(onConfirmDelete).toHaveBeenCalledTimes(1);
  });

  it('surfaces a failed deletion without losing the drawer', () => {
    setup({ deleteError: 'No se pudo eliminar el movimiento' });

    expect(screen.getByRole('alert').textContent).toBe('No se pudo eliminar el movimiento');
  });
});
