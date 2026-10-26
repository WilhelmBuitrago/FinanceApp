import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { DistributionSection } from '../../../components/dashboard/DistributionSection';
import type { DistributionSlice } from '../../../api/dashboard';
import type { UseQueryState } from '../../../types/api';

/* Distributions: the backend's classification, drawn as two donuts with
   their legends. Each side resolves on its own; an empty side is an
   honest empty state, a failed side says it failed. */

const slice = (category: string, amount_minor: number): DistributionSlice => ({
  category,
  amount_minor,
});

const ready = (data: DistributionSlice[]): UseQueryState<DistributionSlice[]> => ({
  loading: false,
  data,
  empty: data.length === 0,
});

const errorState: UseQueryState<DistributionSlice[]> = {
  loading: false,
  error: new Error('backend down'),
};

describe('DistributionSection', () => {
  it('shows the loading placeholder while either side loads', () => {
    const { container } = render(
      <DistributionSection income={{ loading: true }} expense={ready([])} />,
    );

    expect(container.querySelector('.animate-pulse')).not.toBeNull();
  });

  it('draws both legends with real values and percentages', () => {
    render(
      <DistributionSection
        income={ready([slice('Sueldo', 900000)])}
        expense={ready([slice('Comida', 600000), slice('Transporte', 400000)])}
      />,
    );

    expect(screen.getByText('$9.000,00')).toBeTruthy();
    expect(screen.getByText('100%')).toBeTruthy();
    expect(screen.getByText('Comida')).toBeTruthy();
    expect(screen.getByText('$6.000,00')).toBeTruthy();
    expect(screen.getByText('60%')).toBeTruthy();
    expect(screen.getByText('Transporte')).toBeTruthy();
    expect(screen.getByText('40%')).toBeTruthy();
  });

  it('labels the backend uncategorized key in Spanish', () => {
    render(
      <DistributionSection income={ready([slice('uncategorized', 500000)])} expense={ready([])} />,
    );

    expect(screen.getByText('Sin categorizar')).toBeTruthy();
  });

  it('groups the smallest slices beyond five as Otros', () => {
    render(
      <DistributionSection
        income={ready([
          slice('a', 60000),
          slice('b', 50000),
          slice('c', 40000),
          slice('d', 30000),
          slice('e', 20000),
          slice('f', 10000),
        ])}
        expense={ready([])}
      />,
    );

    expect(screen.getByText('Otros')).toBeTruthy();
    expect(screen.getByText('$100,00')).toBeTruthy();
  });

  it('keeps each side independent: an empty income with data in expense', () => {
    render(<DistributionSection income={ready([])} expense={ready([slice('Comida', 500000)])} />);

    expect(screen.getByText('Sin ingresos clasificados')).toBeTruthy();
    expect(screen.queryByText('Sin egresos clasificados')).toBeNull();
    expect(screen.getByText('Comida')).toBeTruthy();
  });

  it('says a failed side failed, without blaming the other one', () => {
    render(<DistributionSection income={errorState} expense={ready([slice('Comida', 500000)])} />);

    expect(screen.getByText('No se pudo cargar la distribución.')).toBeTruthy();
    expect(screen.getByText('Comida')).toBeTruthy();
  });
});
