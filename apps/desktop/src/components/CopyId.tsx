import { useState } from 'react';

/* IDs live only in drawers: mono 12px muted, click to copy. */
export function CopyId({ id }: { id: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      onClick={() => {
        navigator.clipboard
          ?.writeText(id)
          .then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
          })
          .catch(() => undefined);
      }}
      className="font-mono text-meta text-ink-3 transition-colors hover:text-ink"
      title="Copiar ID"
    >
      {copied ? 'Copiado' : `${id.slice(0, 13)}…`}
    </button>
  );
}
