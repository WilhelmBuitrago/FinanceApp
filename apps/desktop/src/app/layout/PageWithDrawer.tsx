import type { ReactNode } from 'react';

import { drawerReserveClass } from '../../components/DrawerShell';

/* Pages that open a drawer reserve its column while it is open. Three
   pages repeated this transition — and its width had to match the
   shell's own panel width. The rule lives here once, in the layout
   layer, and the geometry comes from DrawerShell itself. */

export function PageWithDrawer({
  open,
  className = '',
  children,
}: {
  open: boolean;
  className?: string;
  children: ReactNode;
}) {
  return (
    <div
      className={`transition-[padding] duration-[220ms] ease-out-tk ${className} ${
        open ? drawerReserveClass : ''
      }`}
    >
      {children}
    </div>
  );
}
