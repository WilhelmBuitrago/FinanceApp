import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { ReviewItem } from '../../api/review';
import { ReviewPeriodCard } from './ReviewPeriodCard';

const item: ReviewItem = {
  ID: 'rv_1',
  Kind: 'DOCUMENT_PERIOD',
  Status: 'PENDING',
  CreatedAt: '2026-10-24T00:00:00Z',
};

function setup(overrides: Partial<React.ComponentProps<typeof ReviewPeriodCard>> = {}) {
  const props = {
    item,
    payload: { raw_period: 'PERIODO ILEGIBLE' },
    monthValue: '',
    yearValue: '',
    onMonth: vi.fn(),
    onYear: vi.fn(),
    onConfirm: vi.fn(),
    onDiscard: vi.fn(),
    ...overrides,
  };
  render(<ReviewPeriodCard {...props} />);
  return props;
}

describe('ReviewPeriodCard', () => {
  it('shows the raw period it could not read', () => {
    setup();
    expect(screen.getByText(/PERIODO ILEGIBLE/)).toBeInTheDocument();
  });

  it('keeps confirm disabled until both month and year are set', () => {
    const { rerender } = render(
      <ReviewPeriodCard
        item={item}
        payload={{ raw_period: 'PERIODO ILEGIBLE' }}
        monthValue=""
        yearValue=""
        onMonth={vi.fn()}
        onYear={vi.fn()}
        onConfirm={vi.fn()}
        onDiscard={vi.fn()}
      />,
    );
    expect(screen.getByRole('button', { name: 'Confirmar período' })).toBeDisabled();
    rerender(
      <ReviewPeriodCard
        item={item}
        payload={{ raw_period: 'PERIODO ILEGIBLE' }}
        monthValue="9"
        yearValue=""
        onMonth={vi.fn()}
        onYear={vi.fn()}
        onConfirm={vi.fn()}
        onDiscard={vi.fn()}
      />,
    );
    expect(screen.getByRole('button', { name: 'Confirmar período' })).toBeDisabled();
    rerender(
      <ReviewPeriodCard
        item={item}
        payload={{ raw_period: 'PERIODO ILEGIBLE' }}
        monthValue="9"
        yearValue="2026"
        onMonth={vi.fn()}
        onYear={vi.fn()}
        onConfirm={vi.fn()}
        onDiscard={vi.fn()}
      />,
    );
    expect(screen.getByRole('button', { name: 'Confirmar período' })).toBeEnabled();
  });

  it('reports month and year picks', () => {
    const props = setup({ monthValue: '9', yearValue: '2026' });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: '10' } });
    expect(props.onMonth).toHaveBeenCalledWith('10');
    fireEvent.change(screen.getByPlaceholderText('Año'), { target: { value: '2025' } });
    expect(props.onYear).toHaveBeenCalledWith('2025');
  });

  it('labels the preloaded values as a suggestion, never as fact', () => {
    setup({
      payload: { raw_period: 'X', suggested_month: 9, suggested_year: 2026 },
      monthValue: '9',
      yearValue: '2026',
    });
    expect(screen.getByText(/es una sugerencia, no un dato confirmado/)).toBeInTheDocument();
  });

  it('can discard the item', () => {
    const props = setup();
    fireEvent.click(screen.getByRole('button', { name: 'Descartar' }));
    expect(props.onDiscard).toHaveBeenCalledTimes(1);
  });
});
