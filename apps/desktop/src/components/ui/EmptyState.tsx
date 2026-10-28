import React from 'react';

interface EmptyStateProps {
  title?: string;
  description?: string;
  action?: React.ReactNode;
}

/* Empty state — what is empty, why it can be empty, and the next step.
   Never framed as an error. */

export const EmptyState = ({ title = 'Sin datos', description, action }: EmptyStateProps) => (
  <div className="flex flex-col items-center justify-center rounded-card border border-dashed border-line bg-surface/50 px-8 py-14 text-center">
    <h2 className="text-overline font-semibold text-ink">{title}</h2>
    {description && <p className="mt-2 max-w-md text-body text-ink-2">{description}</p>}
    {action && <div className="mt-5">{action}</div>}
  </div>
);
