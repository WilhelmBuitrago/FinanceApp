import type { ReviewItem } from '../../api/review';

const MONTHS = [
  { n: 1, label: 'Enero' },
  { n: 2, label: 'Febrero' },
  { n: 3, label: 'Marzo' },
  { n: 4, label: 'Abril' },
  { n: 5, label: 'Mayo' },
  { n: 6, label: 'Junio' },
  { n: 7, label: 'Julio' },
  { n: 8, label: 'Agosto' },
  { n: 9, label: 'Septiembre' },
  { n: 10, label: 'Octubre' },
  { n: 11, label: 'Noviembre' },
  { n: 12, label: 'Diciembre' },
];

/* DOCUMENT_PERIOD card: the processor could not read the statement
   period. The human picks month and year; the backend derives the full
   period and the parked movement dates of this document against it. The
   suggestion comes from the parked raws and is explicitly a suggestion. */

interface Props {
  item: ReviewItem;
  payload: any;
  monthValue: string;
  yearValue: string;
  onMonth: (value: string) => void;
  onYear: (value: string) => void;
  onConfirm: () => void;
  onDiscard: () => void;
}

export function ReviewPeriodCard({
  payload: p,
  monthValue,
  yearValue,
  onMonth,
  onYear,
  onConfirm,
  onDiscard,
}: Props) {
  return (
    <div className="rounded-[12px] border border-line bg-surface p-6">
      <p className="text-body font-medium">Período del documento no reconocido</p>
      <p className="mt-1 text-meta text-ink-3">
        El extracto no traía un período legible
        {p.raw_period ? ` (“${p.raw_period}”)` : ''}. Elegí el mes y el año del estado de cuenta:
        con eso se calcula el período completo y se derivan las fechas de los movimientos
        estacionados de este documento contra él.
      </p>
      <div className="mt-4 flex flex-wrap gap-2">
        <select
          value={monthValue}
          onChange={(e) => onMonth(e.target.value)}
          className="border rounded px-2 py-1 bg-black text-white"
        >
          <option value="">Mes…</option>
          {MONTHS.map((m) => (
            <option key={m.n} value={m.n}>
              {m.label}
            </option>
          ))}
        </select>
        <input
          type="number"
          min={2000}
          max={2100}
          value={yearValue}
          onChange={(e) => onYear(e.target.value)}
          placeholder="Año"
          className="w-28 border rounded px-2 py-1 bg-black text-white"
        />
        <button
          onClick={onConfirm}
          disabled={!monthValue || !yearValue}
          className="px-3 py-1 rounded bg-accent-strong text-white disabled:opacity-40"
        >
          Confirmar período
        </button>
        <button onClick={onDiscard} className="px-3 py-1 rounded border text-neg">
          Descartar
        </button>
      </div>
      {(p.suggested_month || p.suggested_year) && (
        <p className="mt-2 text-meta text-ink-3">
          Mes y año precargados a partir de las fechas de los movimientos estacionados de este
          documento: es una sugerencia, no un dato confirmado. Verificalos contra el PDF antes de
          confirmar.
        </p>
      )}
    </div>
  );
}
