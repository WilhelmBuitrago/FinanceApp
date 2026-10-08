import React, { useEffect } from 'react';

/* Shared right drawer shell — push-style panel (fixed right, below topbar).
   Esc closes. Headers and content belong to the caller. */

interface DrawerShellProps {
  ariaLabel: string;
  onClose: () => void;
  children: React.ReactNode;
}

export function DrawerShell({ ariaLabel, onClose, children }: DrawerShellProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <aside
      aria-label={ariaLabel}
      className="fixed bottom-0 right-0 top-12 z-30 flex w-[min(440px,40vw)] flex-col border-l border-line bg-raised"
    >
      {children}
    </aside>
  );
}

/* Drawer header row: title (or amount) on the left, close on the right. */
export function DrawerHeader({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-14 shrink-0 items-center justify-between border-b border-line px-4">
      {children}
    </div>
  );
}

export function DrawerCloseButton({ onClose }: { onClose: () => void }) {
  return (
    <button
      onClick={onClose}
      aria-label="Cerrar panel"
      className="flex h-8 w-8 items-center justify-center rounded-tk text-ink-3 transition-colors duration-[120ms] hover:bg-interactive hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="square" aria-hidden="true">
        <path d="M6 6l12 12M18 6 6 18" />
      </svg>
    </button>
  );
}

/* Body of a drawer: scrollable, sectioned. Async sections announce politely. */
export function DrawerBody({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex-1 space-y-6 overflow-y-auto p-4" role="status" aria-live="polite">
      {children}
    </div>
  );
}

/* Section eyebrow inside drawers. */
export function DrawerSectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="font-mono text-meta font-medium uppercase tracking-[0.08em] text-ink-3">
      {children}
    </h2>
  );
}
