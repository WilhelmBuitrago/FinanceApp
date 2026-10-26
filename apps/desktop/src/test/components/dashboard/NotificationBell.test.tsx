import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { NotificationBell } from '../../../components/dashboard/NotificationBell';

/* The review queue as a bell: the count, the popover, and the click that
   opens the queue. Dismissal is the hook's contract (pinned in
   usePopover.test); this pins the composition and the navigation. */

function LocationProbe() {
  const location = useLocation();
  return <p data-testid="location">{location.pathname}</p>;
}

const renderBell = (notifications: string[]) =>
  render(
    <MemoryRouter>
      <NotificationBell notifications={notifications} />
      <LocationProbe />
    </MemoryRouter>,
  );

describe('NotificationBell', () => {
  it('starts closed and shows the count on the trigger', () => {
    renderBell(['Movimiento sin categoría', 'Documento con período no leído']);

    expect(screen.getByTitle('Notificaciones')).toBeTruthy();
    expect(screen.getByText('2')).toBeTruthy();
    expect(screen.queryByText('Notificaciones', { selector: 'p' })).toBeNull();
  });

  it('opens the popover with one line per parked item', () => {
    renderBell(['Movimiento sin categoría', 'Documento con período no leído']);

    fireEvent.click(screen.getByTitle('Notificaciones'));

    expect(screen.getByText('Movimiento sin categoría')).toBeTruthy();
    expect(screen.getByText('Documento con período no leído')).toBeTruthy();
  });

  it('says so when there is nothing parked', () => {
    renderBell([]);

    fireEvent.click(screen.getByTitle('Notificaciones'));

    expect(screen.getByText('Sin notificaciones')).toBeTruthy();
  });

  it('opens the review queue when a notification is clicked', () => {
    renderBell(['Movimiento sin categoría']);

    fireEvent.click(screen.getByTitle('Notificaciones'));
    fireEvent.click(screen.getByText('Movimiento sin categoría'));

    expect(screen.getByTestId('location').textContent).toBe('/review');
  });

  it('toggles closed on a second click on the trigger', () => {
    renderBell(['Movimiento sin categoría']);

    fireEvent.click(screen.getByTitle('Notificaciones'));
    expect(screen.getByText('Movimiento sin categoría')).toBeTruthy();

    fireEvent.click(screen.getByTitle('Notificaciones'));
    expect(screen.queryByText('Movimiento sin categoría')).toBeNull();
  });

  it('closes on Escape', () => {
    renderBell(['Movimiento sin categoría']);

    fireEvent.click(screen.getByTitle('Notificaciones'));
    fireEvent.keyDown(window, { key: 'Escape' });

    expect(screen.queryByText('Movimiento sin categoría')).toBeNull();
  });
});
