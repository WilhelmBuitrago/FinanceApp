import React from 'react';

/* ChartCard — the chart's panel: eyebrow title + optional context on the
   right, body below. Same surface language as tables. */

interface ChartCardProps {
  title: string;
  context?: string;
  children: React.ReactNode;
}

export function ChartCard({ title, context, children }: ChartCardProps) {
  return (
    <section className="overflow-hidden rounded-tk border border-line bg-surface">
      <div className="flex items-center justify-between border-b border-line bg-raised px-3 py-2">
        <h2 className="font-mono text-meta font-medium uppercase tracking-[0.08em] text-ink-3">
          {title}
        </h2>
        {context && <span className="text-meta text-ink-3">{context}</span>}
      </div>
      <div className="p-4">{children}</div>
    </section>
  );
}
