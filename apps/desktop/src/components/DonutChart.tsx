import React from 'react';

import { formatMoney } from '../lib/formatMoney';
import type { DistributionSlice } from '../api/dashboard';

/* Distribution donut — real backend slices only. Categorical colors come
   from existing tokens; >5 slices group the smallest as "Otros" (visual
   grouping of received data, S12 §24). The "uncategorized" key is the
   backend's honest label for events without category. */

/* Categorical palette from existing tokens — data encoding, not decoration. */
const SLICE_COLORS = ['#a78bfa', '#34d399', '#fbbf24', '#fb7185', '#858591'];

const CATEGORY_LABELS: Record<string, string> = {
  uncategorized: 'Sin categorizar',
};

interface DonutChartProps {
  slices: DistributionSlice[];
}

interface LegendEntry {
  label: string;
  value: number;
  pct: number;
  color: string;
}

function buildLegend(slices: DistributionSlice[]): { entries: LegendEntry[]; total: number } {
  const total = slices.reduce((acc, s) => acc + s.amount_minor, 0);
  if (total <= 0) return { entries: [], total };

  const sorted = [...slices].sort((a, b) => b.amount_minor - a.amount_minor);
  let shown = sorted.slice(0, 5);
  const rest = sorted.slice(5);
  if (rest.length > 0) {
    shown = [
      ...shown,
      {
        category: '__other__',
        amount_minor: rest.reduce((acc, s) => acc + s.amount_minor, 0),
      },
    ];
  }
  const entries: LegendEntry[] = shown.map((s, i) => ({
    label: s.category === '__other__' ? 'Otros' : CATEGORY_LABELS[s.category] ?? s.category,
    value: s.amount_minor,
    pct: (s.amount_minor / total) * 100,
    color: s.category === '__other__' ? '#3a3a4a' : SLICE_COLORS[i % SLICE_COLORS.length],
  }));
  return { entries, total };
}

export function DonutChart({ slices }: DonutChartProps) {
  const { entries, total } = buildLegend(slices);

  if (entries.length === 0) return null;

  const R = 52;
  const CIRC = 2 * Math.PI * R;
  let offset = 0;

  return (
    <div className="flex items-center gap-6">
      <svg viewBox="0 0 128 128" className="h-32 w-32 shrink-0" role="img" aria-label="Distribución">
        {entries.map((e) => {
          const dash = (e.pct / 100) * CIRC;
          const el = (
            <circle
              key={e.label}
              cx="64"
              cy="64"
              r={R}
              fill="none"
              stroke={e.color}
              strokeWidth="16"
              strokeDasharray={`${dash} ${CIRC - dash}`}
              strokeDashoffset={-offset}
              transform="rotate(-90 64 64)"
            />
          );
          offset += dash;
          return el;
        })}
      </svg>
      <dl className="min-w-0 flex-1 space-y-1.5">
        {entries.map((e) => (
          <div key={e.label} className="flex items-center justify-between gap-3">
            <dt className="flex min-w-0 items-center gap-2">
              <span className="h-2 w-2 shrink-0 rounded-full" style={{ backgroundColor: e.color }} aria-hidden="true" />
              <span className="truncate text-body text-ink-2">{e.label}</span>
            </dt>
            <dd className="shrink-0 text-right">
              <span className="tabular-nums text-body font-medium text-ink">{formatMoney(e.value)}</span>
              <span className="ml-2 tabular-nums text-meta text-ink-3">
                {Math.round(e.pct)}%
              </span>
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
