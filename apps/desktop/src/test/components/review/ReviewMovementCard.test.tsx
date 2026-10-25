import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { ReviewItem } from '../../../api/review';
import type { Account } from '../../../types/account';
import { ReviewMovementCard } from '../../../components/review/ReviewMovementCard';

const payload = {
  description: 'EXTRACCION SIN MONTO',
  description_original: 'EXTRACCION SIN MONTO',
  amount_minor: -2500,
  account_id: 'acc_1',
  date: '2026-10-02',
  reason: 'Movimiento no normalizado',
};

function setup(overrides: Partial<React.ComponentProps<typeof ReviewMovementCard>> = {}) {
  const item: ReviewItem = {
    ID: 'rv_1',
    Kind: 'EXTRACTION',
    Status: 'PENDING',
    CreatedAt: '2026-10-24T00:00:00Z',
  };
  const account: Account = {
    ID: 'acc_1',
    UserID: 'local',
    Name: 'Nu',
    AccountType: 'BANK_ACCOUNT',
    Currency: 'COP',
    Status: 'ACTIVE',
    CreatedAt: '2026-10-01T00:00:00Z',
    UpdatedAt: '2026-10-01T00:00:00Z',
  };
  const props = {
    item,
    payload,
    form: {},
    accounts: [account],
    onChange: vi.fn(),
    onSubmit: vi.fn(),
    onDiscard: vi.fn(),
    ...overrides,
  };
  render(<ReviewMovementCard {...props} />);
  return props;
}

describe('ReviewMovementCard', () => {
  it('explains the parked reason for an EXTRACTION item', () => {
    setup();
    expect(screen.getByText('Movimiento no normalizado')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Confirmar corrección' })).toBeInTheDocument();
  });

  it('warns about the collision for a DUPLICATE item', () => {
    setup({
      item: {
        ID: 'rv_2',
        Kind: 'DUPLICATE',
        Status: 'PENDING',
        CreatedAt: '2026-10-24T00:00:00Z',
      },
    });
    expect(screen.getByText(/Ya existe un movimiento/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Confirmar y agregar' })).toBeInTheDocument();
  });

  it('reports every edited field as a patch', () => {
    const props = setup();
    fireEvent.change(screen.getByPlaceholderText('Descripción corregida'), {
      target: { value: 'Compra corregida' },
    });
    expect(props.onChange).toHaveBeenCalledWith({ description: 'Compra corregida' });
    fireEvent.change(screen.getByPlaceholderText('Monto'), { target: { value: '30.5' } });
    expect(props.onChange).toHaveBeenCalledWith({ amount: '30.5' });
    // The type select is the first combobox (the second is the account).
    fireEvent.change(screen.getAllByRole('combobox')[0], { target: { value: 'INCOME' } });
    expect(props.onChange).toHaveBeenCalledWith({ type: 'INCOME' });
  });

  it('submits and discards', () => {
    const props = setup();
    fireEvent.click(screen.getByRole('button', { name: 'Confirmar corrección' }));
    expect(props.onSubmit).toHaveBeenCalledTimes(1);
    fireEvent.click(screen.getByRole('button', { name: 'Descartar' }));
    expect(props.onDiscard).toHaveBeenCalledTimes(1);
  });
});
