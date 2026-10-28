import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { RecentMovementsTable } from '../../../components/dashboard/RecentMovementsTable';
import type { UseQueryState } from '../../../types/api';
import type { Movement } from '../../../types/movement';

/* Recent activity: the backend already limits the list to five; this
   renders every documented record as its own row with its real values,
   and the honest states around it. */

const movement = (over: Partial<Movement>): Movement => ({
  ID: 'mv-1',
  FinancialEventID: 'fe-1',
  AccountID: 'acc-1',
  MovementDate: '2026-10-20',
  AmountMinor: -1500,
  Description: 'Mercado',
  Status: 'ACTIVE',
  categoryName: 'Comida',
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
  ...over,
});

const ready = (data: Movement[]): UseQueryState<Movement[]> => ({
  loading: false,
  data,
  empty: data.length === 0,
});

const renderTable = (movements: UseQueryState<Movement[]>) =>
  render(
    <MemoryRouter>
      <RecentMovementsTable
        movements={movements}
        accountNames={
          new Map([
            ['acc-1', 'Nu'],
            ['acc-2', 'Bancolombia'],
          ])
        }
      />
    </MemoryRouter>,
  );

describe('RecentMovementsTable', () => {
  it('shows the skeleton while loading and no rows', () => {
    renderTable({ loading: true });

    expect(screen.queryAllByRole('row')).toHaveLength(0);
  });

  it('shows the empty state when there is nothing to show', () => {
    renderTable(ready([]));

    expect(screen.getByText('Sin movimientos')).toBeTruthy();
  });

  it('renders every documented record with its real values', () => {
    renderTable(
      ready([
        movement({ ID: 'mv-1' }),
        movement({
          ID: 'mv-2',
          AmountMinor: 9000,
          Description: null,
          categoryName: 'Transporte',
          AccountID: 'acc-2',
          MovementDate: '2026-10-21',
        }),
      ]),
    );

    expect(screen.getByText('20 oct 2026')).toBeTruthy();
    expect(screen.getByText('21 oct 2026')).toBeTruthy();
    expect(screen.getByText('Mercado')).toBeTruthy();
    expect(screen.getByText('Comida')).toBeTruthy();
    expect(screen.getByText('Transporte')).toBeTruthy();
    expect(screen.getByText('Nu')).toBeTruthy();
    expect(screen.getByText('Bancolombia')).toBeTruthy();
    expect(screen.getByText('-$15,00')).toBeTruthy();
    expect(screen.getByText('+$90,00')).toBeTruthy();
  });

  it('falls back to a dash for missing description, category and account', () => {
    renderTable(
      ready([movement({ Description: null, categoryName: '', AccountID: 'acc-unknown' })]),
    );

    expect(screen.getAllByText('—').length).toBeGreaterThanOrEqual(3);
  });

  it('never deduplicates: two records render as two rows', () => {
    renderTable(ready([movement({}), movement({})]));

    expect(screen.getAllByRole('row')).toHaveLength(3); // header + 2 records
  });

  it('formats a row with the currency of its movement, not an assumed one', () => {
    renderTable(ready([movement({ AmountMinor: 150000, currency: 'USD' })]));

    expect(screen.getByText(/US\$\s?1\.500,00/)).toBeTruthy();
  });
});
