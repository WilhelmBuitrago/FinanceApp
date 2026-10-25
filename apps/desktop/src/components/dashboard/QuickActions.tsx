import { useState } from 'react';
import { Link } from 'react-router-dom';

import { usePopoverDismiss } from '../../hooks/usePopover';
import { primaryButton } from '../ui/buttons';

/* The "+" quick actions menu on the dashboard header: import a document
   or register a movement by hand. */

export function QuickActions({ onAddMovement }: { onAddMovement: () => void }) {
  const [open, setOpen] = useState(false);
  const ref = usePopoverDismiss(() => setOpen(false));

  return (
    <div className="relative" ref={ref}>
      <button className={primaryButton} onClick={() => setOpen((v) => !v)}>
        +
      </button>
      {open && (
        <div className="absolute right-0 top-full mt-2 w-56 rounded-tk border border-line bg-surface shadow-float p-1 z-30">
          <Link
            to="/documents"
            onClick={() => setOpen(false)}
            className="block px-3 py-2 rounded-sm hover:bg-raised text-body"
          >
            Importar documento
          </Link>
          <button
            onClick={() => {
              setOpen(false);
              onAddMovement();
            }}
            className="w-full text-left px-3 py-2 rounded-sm hover:bg-raised text-body"
          >
            Agregar movimiento
          </button>
        </div>
      )}
    </div>
  );
}
