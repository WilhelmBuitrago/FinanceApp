interface NoticeProps {
  message: string | null;
  onDismiss?: () => void;
}

/* Inline notice (role=status) for operation feedback: the same message
   alert() used to give, without blocking the UI. Errors that ARE the
   section's state use ErrorState instead; this is for transient
   outcomes ("este ítem ya fue resuelto", "no se pudo guardar"). */

export function Notice({ message, onDismiss }: NoticeProps) {
  if (!message) return null;
  return (
    <div
      role="status"
      className="flex items-center justify-between gap-3 rounded-tk border border-line bg-raised px-3 py-2 text-body text-ink"
    >
      <span className="min-w-0 break-words">{message}</span>
      {onDismiss && (
        <button
          onClick={onDismiss}
          aria-label="Cerrar aviso"
          className="shrink-0 text-ink-3 transition-colors hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
        >
          ✕
        </button>
      )}
    </div>
  );
}
