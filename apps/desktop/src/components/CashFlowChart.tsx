import React from 'react';

import { formatDate } from '../lib/formatDate';
import { formatSignedMoney } from '../lib/formatMoney';
import type { CashFlowPoint } from '../api/dashboard';

/* Cash flow chart — the money's protagonist. Time-proportional x axis,
   net line over a violet wash area, zero baseline, hover tooltip with the
   point's real inflow/outflow/net. Purely a display transform of backend
   aggregates (S12 §34). */

interface CashFlowChartProps {
  points: CashFlowPoint[]; // sorted by date
}

const W = 800;
const H = 200;
const PAD_X = 8;
const PAD_Y = 16;

export function CashFlowChart({ points }: CashFlowChartProps) {
  const [hovered, setHovered] = React.useState<number | null>(null);
  const containerRef = React.useRef<HTMLDivElement>(null);

  const times = points.map((p) => new Date(p.date).getTime());
  const t0 = times[0];
  const t1 = times[times.length - 1];
  const span = t1 - t0 || 1;

  const nets = points.map((p) => p.net);
  const rawMin = Math.min(0, ...nets);
  const rawMax = Math.max(0, ...nets);
  const pad = Math.max((rawMax - rawMin) * 0.08, 1);
  const yMin = rawMin - pad;
  const yMax = rawMax + pad;

  const xAt = (i: number) => PAD_X + ((times[i] - t0) / span) * (W - 2 * PAD_X);
  const yAt = (net: number) =>
    H - PAD_Y - ((net - yMin) / (yMax - yMin || 1)) * (H - 2 * PAD_Y);
  const yZero = yAt(0);

  const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${xAt(i)},${yAt(p.net)}`).join(' ');
  const areaPath = `${linePath} L${xAt(points.length - 1)},${yZero} L${xAt(0)},${yZero} Z`;

  const onMove = (e: React.MouseEvent) => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const relX = ((e.clientX - rect.left) / rect.width) * W;
    let nearest = 0;
    let best = Infinity;
    for (let i = 0; i < points.length; i++) {
      const d = Math.abs(xAt(i) - relX);
      if (d < best) {
        best = d;
        nearest = i;
      }
    }
    setHovered(nearest);
  };

  const hoveredPoint = hovered != null ? points[hovered] : null;
  const hoverLeftPct =
    hovered != null ? (xAt(hovered) / W) * 100 : 0;
  const flip = hoverLeftPct > 62;

  const dateLabels = [
    points[0],
    points[Math.floor((points.length - 1) / 2)],
    points[points.length - 1],
  ];

  return (
    <div className="space-y-2">
      <div
        ref={containerRef}
        className="relative"
        onMouseMove={onMove}
        onMouseLeave={() => setHovered(null)}
        role="img"
        aria-label={`Flujo de dinero por día: ${points.length} días con actividad`}
      >
        <svg viewBox={`0 0 ${W} ${H}`} className="block h-48 w-full" aria-hidden="true">
          {/* zero baseline */}
          <line x1={PAD_X} y1={yZero} x2={W - PAD_X} y2={yZero} stroke="#232330" strokeWidth="1" />
          {/* area */}
          <path d={areaPath} fill="rgb(124 58 237 / 0.14)" />
          {/* line */}
          <path d={linePath} fill="none" stroke="#a78bfa" strokeWidth="1.5" strokeLinejoin="round" />
          {/* hover guide + dot */}
          {hovered != null && (
            <>
              <line
                x1={xAt(hovered)}
                y1={PAD_Y}
                x2={xAt(hovered)}
                y2={H - PAD_Y}
                stroke="#232330"
                strokeWidth="1"
              />
              <circle cx={xAt(hovered)} cy={yAt(points[hovered].net)} r="3" fill="#a78bfa" />
            </>
          )}
        </svg>

        {hoveredPoint && (
          <div
            className={`pointer-events-none absolute top-2 z-10 rounded-tk border border-line bg-interactive px-2.5 py-2 text-meta shadow-float ${
              flip ? '' : ''
            }`}
            style={{
              left: `${hoverLeftPct}%`,
              transform: `translateX(${flip ? '-100%' : '0'})`,
            }}
          >
            <p className="font-mono uppercase tracking-[0.08em] text-ink-3">
              {formatDate(hoveredPoint.date)}
            </p>
            <p className="mt-1 flex justify-between gap-4 text-ink-2">
              <span>Ingresos</span>
              <span className="tabular-nums text-pos">{formatSignedMoney(hoveredPoint.inflow)}</span>
            </p>
            <p className="flex justify-between gap-4 text-ink-2">
              <span>Egresos</span>
              <span className="tabular-nums text-neg">-{formatSignedMoney(hoveredPoint.outflow).replace('+', '')}</span>
            </p>
            <p className="flex justify-between gap-4 text-ink">
              <span>Neto</span>
              <span className="tabular-nums font-medium">{formatSignedMoney(hoveredPoint.net)}</span>
            </p>
          </div>
        )}
      </div>

      {/* x labels: first / middle / last */}
      <div className="flex justify-between px-2 text-meta text-ink-3">
        {[...new Set(dateLabels)].map((p) => (
          <span key={p.date}>{formatDate(p.date)}</span>
        ))}
      </div>
    </div>
  );
}
