import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ReviewCategoryCard } from '../../../components/review/ReviewCategoryCard';

const payload = {
  category_name: 'Supermercado',
  category_type: 'EXPENSE',
  movement_date: '2026-10-01',
  amount_minor: -2500,
};
const categories = [
  { id: 'cat_1', name: 'Cine', type: 'EXPENSE' },
  { id: 'cat_2', name: 'Sueldo', type: 'INCOME' },
];

function setup(overrides: Partial<React.ComponentProps<typeof ReviewCategoryCard>> = {}) {
  const props = {
    payload,
    categories,
    selectedId: '',
    onSelectChange: vi.fn(),
    onAssign: vi.fn(),
    onDiscard: vi.fn(),
    ...overrides,
  };
  render(<ReviewCategoryCard {...props} />);
  return props;
}

describe('ReviewCategoryCard', () => {
  it('names the deleted category and offers only categories of the same type', () => {
    setup();
    expect(screen.getByText(/Supermercado/)).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Cine' })).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'Sueldo' })).not.toBeInTheDocument();
  });

  it('lets the human leave the movement uncategorized', () => {
    const props = setup();
    fireEvent.click(screen.getByRole('button', { name: 'Guardar sin categoría' }));
    expect(props.onAssign).toHaveBeenCalledTimes(1);
  });

  it('reports the picked category and switches the button copy', () => {
    const props = setup({ selectedId: 'cat_1' });
    expect(screen.getByRole('button', { name: 'Asignar categoría' })).toBeInTheDocument();
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'cat_1' } });
    expect(props.onSelectChange).toHaveBeenCalledWith('cat_1');
    fireEvent.click(screen.getByRole('button', { name: 'Asignar categoría' }));
    expect(props.onAssign).toHaveBeenCalledTimes(1);
  });

  it('can discard the item', () => {
    const props = setup();
    fireEvent.click(screen.getByRole('button', { name: 'Descartar' }));
    expect(props.onDiscard).toHaveBeenCalledTimes(1);
  });
});
