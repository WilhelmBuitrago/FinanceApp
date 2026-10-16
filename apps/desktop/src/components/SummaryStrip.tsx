import { formatSignedMoney } from '../lib/formatMoney';

/* Income/expense/net summary strip — the Overview's primary surface.
   Real sums only; the caller decides the period and the label.
   Net is the dominant figure (24px tabular, violet hairline underline). */

interface SummaryStripProps {
  income: number;
  expense: number;
  periodLabel: string;
}

export function SummaryStrip({ income, expense, periodLabel }: SummaryStripProps) {
  return (
    <div className="rounded-[12px] border border-line bg-surface p-6">
      <div className="flex flex-wrap items-end gap-x-10 gap-y-6">
        <div className="min-w-40">
          <p className="font-mono text-meta uppercase tracking-[0.08em] text-ink-3">
            Ingresos · {periodLabel}
          </p>
          <p className="mt-2 text-[1.375rem] font-medium tabular-nums text-pos">
            {formatSignedMoney(income)}
          </p>
        </div>
        <div className="w-px self-stretch bg-line" aria-hidden="true" />
        <div className="min-w-40">
          <p className="font-mono text-meta uppercase tracking-[0.08em] text-ink-3">
            Egresos · {periodLabel}
          </p>
          <p className="mt-2 text-[1.375rem] font-medium tabular-nums text-neg">
            {formatSignedMoney(expense)}
          </p>
        </div>
        <div className="w-px self-stretch bg-line" aria-hidden="true" />
        <div className="min-w-40">
          <p className="font-mono text-meta uppercase tracking-[0.08em] text-ink-3">
            Neto · {periodLabel}
          </p>
          <p className="mt-2 inline-block border-b border-accent pb-1 text-[1.75rem] font-semibold tabular-nums text-ink">
            {formatSignedMoney(income + expense)}
          </p>
        </div>
      </div>
    </div>
  );
}
