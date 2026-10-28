import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { FinancialSummarySection } from '../../../components/dashboard/FinancialSummarySection';
import type { CashFlowPoint } from '../../../api/dashboard';
import type { UseQueryState } from '../../../types/api';

/* The 30-day window: a visual transform over the backend's daily
   aggregates, anchored at mount. Points inside the window reach the
   strip; older ones are honestly reported as no activity. */

const isoDaysAgo = (days: number) =>
  new Date(Date.now() - days * 86400000).toISOString().slice(0, 10);

const point = (days: number, inflow: number, outflow: number): CashFlowPoint => ({
  date: isoDaysAgo(days),
  inflow,
  outflow,
  net: inflow - outflow,
});

const ready = (data: CashFlowPoint[]): UseQueryState<CashFlowPoint[]> => ({
  loading: false,
  data,
  empty: data.length === 0,
});

describe('FinancialSummarySection', () => {
  it('shows the loading placeholder while the aggregates load', () => {
    const { container } = render(<FinancialSummarySection cashFlow={{ loading: true }} />);

    expect(container.querySelector('.animate-pulse')).not.toBeNull();
    expect(screen.queryByText('Sin movimientos en los últimos 30 días.')).toBeNull();
  });

  it('reports a failed load without inventing figures', () => {
    render(
      <FinancialSummarySection cashFlow={{ loading: false, error: new Error('backend down') }} />,
    );

    expect(
      screen.getByText('No se pudo cargar la situación financiera desde el servidor.'),
    ).toBeTruthy();
  });

  it('sums the points inside the 30-day window', () => {
    render(
      <FinancialSummarySection cashFlow={ready([point(5, 900000, 150000), point(2, 0, 50000)])} />,
    );

    // income 900000, expense -(150000+50000) = -200000, net 700000
    expect(screen.getByText('Ingresos · 30 días')).toBeTruthy();
    expect(screen.getByText('+$9.000,00')).toBeTruthy();
    expect(screen.getByText('-$2.000,00')).toBeTruthy();
    expect(screen.getByText('+$7.000,00')).toBeTruthy();
  });

  it('excludes points older than the window', () => {
    render(<FinancialSummarySection cashFlow={ready([point(60, 900000, 0)])} />);

    expect(screen.getByText('Sin movimientos en los últimos 30 días.')).toBeTruthy();
  });

  it('says so when there is no activity at all', () => {
    render(<FinancialSummarySection cashFlow={ready([])} />);

    expect(screen.getByText('Sin movimientos en los últimos 30 días.')).toBeTruthy();
  });
});
