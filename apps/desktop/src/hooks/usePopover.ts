import { useEffect, useRef } from 'react';

/* Popover dismissal: Escape and an outside pointer close it. The
   dashboard bell and its action menu stayed open until the next click
   on the trigger; every other dismissable surface in the app
   (ConfirmModal) already honors both. The ref wraps trigger AND
   popover, so clicking the trigger toggles instead of closing. */

export function usePopoverDismiss(onDismiss: () => void) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onDismiss();
    };
    const onPointerDown = (e: PointerEvent) => {
      const target = e.target as Node;
      if (ref.current && !ref.current.contains(target)) onDismiss();
    };
    window.addEventListener('keydown', onKey);
    window.addEventListener('pointerdown', onPointerDown);
    return () => {
      window.removeEventListener('keydown', onKey);
      window.removeEventListener('pointerdown', onPointerDown);
    };
  }, [onDismiss]);

  return ref;
}
