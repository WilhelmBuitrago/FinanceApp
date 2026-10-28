import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { getAccounts } from '../../api/accounts';
import { getMovements } from '../../api/movements';
import TransactionsPage from '../../pages/TransactionsPage';
import type { Account } from '../../types/account';
import type { Movement } from '../../types/movement';

vi.mock('../../api/accounts', () => ({ getAccounts: vi.fn() }));
vi.mock('../../api/movements', () => ({ getMovements: vi.fn() }));
// The drawer is another component's contract (already tested there); here
// it only needs to announce that a row opened it.
vi.mock('../../components/DetailDrawer', () => ({
  DetailDrawer: ({ movement }: { movement: unknown }) =>
    movement ? <div data-testid="detail-drawer" /> : null,
}));

const accounts: Account[] = [
  {
    ID: 'acc_1',
    UserID: 'usr_1',
    Name: 'Cuenta Nu',
    AccountType: 'BANK_ACCOUNT',
    Currency: 'COP',
    Status: 'ACTIVE',
    CreatedAt: '2026-09-01T00:00:00Z',
    UpdatedAt: '2026-09-01T00:00:00Z',
  },
  {
    ID: 'acc_2',
    UserID: 'usr_1',
    Name: 'Tarjeta Visa',
    AccountType: 'CREDIT_CARD',
    Currency: 'COP',
    Status: 'ACTIVE',
    CreatedAt: '2026-09-02T00:00:00Z',
    UpdatedAt: '2026-09-02T00:00:00Z',
  },
];

function movement(overrides: Partial<Movement> & { ID: string }): Movement {
  return {
    FinancialEventID: `fe_${overrides.ID}`,
    AccountID: 'acc_1',
    MovementDate: '2026-10-01',
    AmountMinor: -2500,
    Description: 'Mercado',
    Status: 'ACTIVE',
    categoryName: 'Comida',
    CreatedAt: '2026-10-01T00:00:00Z',
    UpdatedAt: '2026-10-01T00:00:00Z',
    ...overrides,
  };
}

const movements: Movement[] = [
  movement({ ID: 'mv_1' }),
  movement({ ID: 'mv_2', AccountID: 'acc_2', Description: 'Gasolina', AmountMinor: 90000 }),
  // A non-ACTIVE status is the only reason the Estado column exists.
  movement({ ID: 'mv_3', Status: 'REVERSED', Description: 'Reembolso' }),
];

function setup(initialEntry = '/transactions') {
  render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <TransactionsPage />
    </MemoryRouter>,
  );
}

function rows() {
  // query, not get: the empty state legitimately renders no rows.
  return screen.queryAllByRole('row').filter((r) => !r.closest('thead'));
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getAccounts).mockResolvedValue(accounts);
  vi.mocked(getMovements).mockResolvedValue(movements);
});

describe('TransactionsPage', () => {
  it('renders every documented movement as its own row', async () => {
    setup();
    await waitFor(() => expect(rows()).toHaveLength(3));
    expect(screen.getByText('Mercado')).toBeTruthy();
    expect(screen.getByText('Gasolina')).toBeTruthy();
    // The Estado column only exists when something non-ACTIVE is present.
    expect(screen.getByRole('columnheader', { name: 'Estado' })).toBeTruthy();
    expect(screen.getByText('Revertidos')).toBeTruthy();
  });

  it('filters by free text through the URL (q) and offers the chip back', async () => {
    setup();
    await waitFor(() => expect(rows()).toHaveLength(3));

    fireEvent.change(screen.getByLabelText('Buscar movimientos por descripción'), {
      target: { value: 'gas' },
    });

    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(screen.getByText('1 de 3 movimientos')).toBeTruthy();

    // Removing the chip restores the full list.
    fireEvent.click(screen.getByLabelText('Quitar filtro “gas”'));
    await waitFor(() => expect(rows()).toHaveLength(3));
  });

  it('filters by account and reports the honest empty state', async () => {
    setup();
    await waitFor(() => expect(rows()).toHaveLength(3));

    fireEvent.change(screen.getByLabelText('Filtrar por cuenta'), {
      target: { value: 'acc_2' },
    });
    await waitFor(() => expect(rows()).toHaveLength(1));
    // The name appears twice (row + filter chip); the row is inside the table.
    expect(within(screen.getByRole('table')).getByText('Tarjeta Visa')).toBeTruthy();
    expect(screen.getByText('Gasolina')).toBeTruthy();

    // A filter that matches nothing shows the no-results state, never an error.
    fireEvent.change(screen.getByLabelText('Buscar movimientos por descripción'), {
      target: { value: 'zzz' },
    });
    await waitFor(() => expect(screen.getByText('Sin resultados')).toBeTruthy());
    expect(rows()).toHaveLength(0);
  });

  it('honors the URL as the initial filter state', async () => {
    setup('/transactions?q=reembolso');
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(screen.getByText('Reembolso')).toBeTruthy();
  });

  it('opens the traceability drawer when a row is activated', async () => {
    setup();
    await waitFor(() => expect(rows()).toHaveLength(3));
    expect(screen.queryByTestId('detail-drawer')).toBeNull();

    fireEvent.click(screen.getByLabelText(/Mercado/));
    expect(screen.getByTestId('detail-drawer')).toBeTruthy();
  });
});
