import React from 'react';

interface EmptyStateProps {
  title?: string;
  description?: string;
  action?: React.ReactNode;
}

/* Empty state — what is empty, why it can be empty, and the next step.
   Never framed as an error. */

export const EmptyState = ({ title = 'Sin datos', description, action }: EmptyStateProps) => (
  <div className="flex flex-col items-center justify-center px-6 py-10 text-center">
    <h2 className="text-body font-medium text-ink">{title}</h2>
    {description && <p className="mt-1.5 max-w-sm text-body text-ink-2">{description}</p>}
    {action && <div className="mt-4">{action}</div>}
  </div>
);
