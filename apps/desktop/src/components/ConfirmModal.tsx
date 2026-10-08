import React, { useEffect, useRef } from 'react';

/* Confirmation modal — for destructive, irreversible actions only.
   Names the specific consequence, action-labeled buttons (never Sí/No),
   focus lands on the SAFE option, Esc and backdrop cancel. */

interface ConfirmModalProps {
  title: string;
  message: string;
  confirmLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
  busy?: boolean;
}

export function ConfirmModal({
  title,
  message,
  confirmLabel,
  onConfirm,
  onCancel,
  busy = false,
}: ConfirmModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    cancelRef.current?.focus();
  }, []);

  // Esc cancels; Tab stays inside the dialog (focus trap).
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onCancel();
        return;
      }
      if (e.key === 'Tab' && dialogRef.current) {
        const focusables = dialogRef.current.querySelectorAll<HTMLElement>('button');
        if (focusables.length === 0) return;
        const first = focusables[0];
        const last = focusables[focusables.length - 1];
        if (e.shiftKey && document.activeElement === first) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onCancel]);

  return (
    <div
      className="fixed inset-0 z-40 flex items-center justify-center bg-black/55"
      onClick={onCancel}
    >
      <div
        ref={dialogRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-title"
        aria-describedby="confirm-message"
        className="mx-4 w-full max-w-md rounded-tk border border-line bg-raised p-4 shadow-float"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="confirm-title" className="text-title font-semibold text-ink">
          {title}
        </h2>
        <p id="confirm-message" className="mt-2 text-body text-ink-2">
          {message}
        </p>
        <div className="mt-6 flex justify-end gap-2">
          <button
            ref={cancelRef}
            onClick={onCancel}
            className="inline-flex h-8 items-center rounded-tk border border-line bg-raised px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          >
            Cancelar
          </button>
          <button
            onClick={onConfirm}
            disabled={busy}
            aria-busy={busy}
            className="inline-flex h-8 items-center rounded-tk border border-neg px-3 text-body font-medium text-neg transition-colors duration-[120ms] hover:bg-neg/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-neg disabled:opacity-40"
          >
            {busy ? 'Borrando…' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
