import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import { QuickActions } from '../../../components/dashboard/QuickActions';

/* The "+" menu: import a document (link) or register a movement by hand
   (callback), each closing the menu. */

describe('QuickActions', () => {
  const setup = () => {
    const onAddMovement = vi.fn();
    render(
      <MemoryRouter>
        <QuickActions onAddMovement={onAddMovement} />
      </MemoryRouter>,
    );
    return { onAddMovement };
  };

  it('starts closed', () => {
    setup();

    expect(screen.queryByText('Importar documento')).toBeNull();
    expect(screen.queryByText('Agregar movimiento')).toBeNull();
  });

  it('opens the menu with both actions', () => {
    setup();

    fireEvent.click(screen.getByText('+'));

    expect(screen.getByRole('link', { name: 'Importar documento' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Agregar movimiento' })).toBeTruthy();
  });

  it('links the import action to the documents page', () => {
    setup();

    fireEvent.click(screen.getByText('+'));

    expect(screen.getByRole('link', { name: 'Importar documento' }).getAttribute('href')).toBe(
      '/documents',
    );
  });

  it('invokes the callback and closes the menu on the manual movement', () => {
    const { onAddMovement } = setup();

    fireEvent.click(screen.getByText('+'));
    fireEvent.click(screen.getByRole('button', { name: 'Agregar movimiento' }));

    expect(onAddMovement).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('button', { name: 'Agregar movimiento' })).toBeNull();
  });

  it('toggles closed on a second click on the trigger', () => {
    setup();

    fireEvent.click(screen.getByText('+'));
    fireEvent.click(screen.getByText('+'));

    expect(screen.queryByRole('link', { name: 'Importar documento' })).toBeNull();
  });
});
