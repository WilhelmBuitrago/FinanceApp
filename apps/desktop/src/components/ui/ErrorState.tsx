interface ErrorStateProps {
  message?: string;
  onRetry?: () => void;
}

/* Error state — what failed, in plain language (no status codes, no stack),
   plus a retry control when one is available. */

export const ErrorState = ({
  message = 'No se pudo cargar la información.',
  onRetry,
}: ErrorStateProps) => (
  <div className="flex flex-col items-center justify-center px-6 py-10 text-center">
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="square"
      className="text-neg"
      aria-hidden="true"
    >
      <path d="M12 3.5 21 20H3z" />
      <path d="M12 10v4.5M12 17.2v.3" />
    </svg>
    <h3 className="mt-2 text-body font-medium text-ink">No se pudo cargar</h3>
    <p className="mt-1.5 max-w-sm text-body text-ink-2">{message}</p>
    {onRetry && (
      <button
        onClick={onRetry}
        className="mt-4 inline-flex h-8 items-center rounded-tk border border-line bg-raised px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
      >
        Reintentar
      </button>
    )}
  </div>
);
