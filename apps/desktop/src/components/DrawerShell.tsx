import React, { useEffect } from 'react';

/* Shared right drawer shell — push-style panel (fixed right, below topbar).
   Esc closes. Headers and content belong to the caller. */

interface DrawerShellProps {
  ariaLabel: string;
  onClose: () => void;
  children: React.ReactNode;
}

/* The page-side counterpart of the panel width below: while a drawer is
   open the page reserves this exact column (PageWithDrawer uses it).
   One geometry, two consumers — they cannot drift apart. */
export const drawerReserveClass = 'pr-[calc(min(440px,40vw)_+_16px)]';

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

/* Key/value row inside drawers — the one definition for both drawers
   (DetailDrawer and DocumentDrawer used to carry a copy each). */
export function KV({ label, children }: { label: string; children?: React.ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-1">
      <dt className="shrink-0 text-meta text-ink-3">{label}</dt>
      <dd className="min-w-0 break-words text-right text-body text-ink">{children ?? '—'}</dd>
    </div>
  );
}

/* Section loading state — discreet bars, aria-hidden (the body wrapper
   announces politely when the content lands). */
export function SectionLoading() {
  return (
    <div className="space-y-2" aria-hidden="true">
      <div className="h-2.5 w-24 animate-pulse rounded-tk bg-interactive" />
      <div className="h-2.5 w-40 animate-pulse rounded-tk bg-interactive" />
    </div>
  );
}

/* Section failure with its retry: one section failing never blanks the
   others. */
export function SectionError({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex items-center gap-3">
      <p className="text-meta text-ink-2">No se pudo cargar esta sección.</p>
      <button onClick={onRetry} className="text-meta font-medium text-accent hover:underline">
        Reintentar
      </button>
    </div>
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
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="square"
        aria-hidden="true"
      >
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
