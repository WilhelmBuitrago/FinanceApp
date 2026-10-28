import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { CashFlowSection } from '../../../components/dashboard/CashFlowSection';
import type { CashFlowPoint } from '../../../api/dashboard';
import type { UseQueryState } from '../../../types/api';

/* Cash flow: the backend's daily points, sorted, drawn from two points
   up. Below that the honest state says so instead of drawing a line out
   of a single day. */

const point = (date: string, net: number): CashFlowPoint => ({
  date,
  inflow: Math.max(net, 0),
  outflow: Math.max(-net, 0),
  net,
});

const ready = (data: CashFlowPoint[]): UseQueryState<CashFlowPoint[]> => ({
  loading: false,
  data,
  empty: data.length === 0,
});

describe('CashFlowSection', () => {
  it('shows the loading placeholder while the points load', () => {
    const { container } = render(<CashFlowSection cashFlow={{ loading: true }} />);

    expect(container.querySelector('.animate-pulse')).not.toBeNull();
    expect(screen.queryByRole('img')).toBeNull();
  });

  it('reports a failed load as an empty state, never as a broken chart', () => {
    render(<CashFlowSection cashFlow={{ loading: false, error: new Error('backend down') }} />);

    expect(screen.getByText('No se pudo cargar el flujo')).toBeTruthy();
  });

  it('draws the chart from two points up, with the day count as context', () => {
    render(
      <CashFlowSection cashFlow={ready([point('2026-10-01', 500), point('2026-10-02', -250)])} />,
    );

    expect(screen.getByText('2 días con actividad')).toBeTruthy();
    expect(screen.getByRole('img', { name: /Flujo de dinero por día: 2 días/ })).toBeTruthy();
  });

  it('refuses to draw a line out of a single day', () => {
    render(<CashFlowSection cashFlow={ready([point('2026-10-01', 500)])} />);

    expect(screen.getByText('Aún no hay suficiente información')).toBeTruthy();
    expect(screen.queryByRole('img')).toBeNull();
  });

  it('sorts the points by date before drawing', () => {
    const { container } = render(
      <CashFlowSection
        cashFlow={ready([
          point('2026-10-03', -100),
          point('2026-10-01', 300),
          point('2026-10-02', 200),
        ])}
      />,
    );

    // The x labels are first / middle / last of the sorted series: the
    // spans that follow the chart container.
    const labels = container.querySelectorAll('[role="img"] ~ div span');
    expect(labels).toHaveLength(3);
    const [first, , last] = Array.from(labels).map((el) => el.textContent);
    expect(first).toBe('01 oct 2026');
    expect(last).toBe('03 oct 2026');
  });
});
