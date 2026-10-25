import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { usePopoverDismiss } from '../../hooks/usePopover';
import { IconBell } from '../icons';

/* Dashboard notification bell — the review queue as notifications: one
   line per parked item, clicking one opens the queue. Esc and outside
   click close it (usePopoverDismiss owns both). */

export function NotificationBell({ notifications }: { notifications: string[] }) {
  const [open, setOpen] = useState(false);
  const ref = usePopoverDismiss(() => setOpen(false));
  const navigate = useNavigate();

  return (
    <div className="relative" ref={ref}>
      <button
        className="relative p-2 rounded-tk border border-line bg-surface"
        onClick={() => setOpen((v) => !v)}
        title="Notificaciones"
        aria-label="Notificaciones"
      >
        <IconBell />
        {notifications.length > 0 && (
          <span className="absolute -top-1 -right-1 bg-neg text-white text-xs rounded-full px-1">
            {notifications.length}
          </span>
        )}
      </button>
      {open && (
        <div className="absolute right-0 top-full mt-2 w-80 rounded-tk border border-line bg-surface shadow-xl p-2 z-30">
          <p className="text-sm font-medium px-2 py-1">Notificaciones</p>
          {notifications.length === 0 ? (
            <p className="text-sm text-ink-3 px-2 py-1">Sin notificaciones</p>
          ) : (
            notifications.map((n, i) => (
              <div
                key={i}
                className="px-2 py-1 text-sm border-b border-line last:border-0 cursor-pointer hover:bg-raised"
                onClick={() => {
                  setOpen(false);
                  navigate('/review');
                }}
              >
                {n}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
