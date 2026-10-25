import { formatSignedMoney } from '../../lib/formatMoney';
import type { ReviewItem } from '../../api/review';
import type { Account } from '../../types/account';
import type { ReviewPayload } from '../../types/review-payload';

/* DUPLICATE/EXTRACTION card: the human confirms (or edits) the movement to
   write. A DUPLICATE says "create it anyway"; an EXTRACTION carries the
   reason the processor parked it. Both resolve with the same fields. */

interface Props {
  item: ReviewItem;
  payload: ReviewPayload;
  form: Record<string, string>;
  accounts: Account[];
  onChange: (patch: Record<string, string>) => void;
  onSubmit: () => void;
  onDiscard: () => void;
}

export function ReviewMovementCard({
  item,
  payload: p,
  form: f,
  accounts,
  onChange,
  onSubmit,
  onDiscard,
}: Props) {
  const isDuplicate = item.Kind === 'DUPLICATE';
  return (
    <div className="rounded-card border border-line bg-surface p-6">
      <p className="text-body font-medium">
        {isDuplicate ? 'Movimiento duplicado detectado' : 'Extracción requiere verificación'}
      </p>
      <p className="mt-1 text-meta text-ink-3">
        {isDuplicate
          ? 'Ya existe un movimiento con estos datos. Confirmá si corresponde crear uno nuevo.'
          : p.reason || 'Movimiento no normalizado'}
      </p>
      <ul className="mt-3 space-y-1 text-body">
        <li>
          <b>Descripción original:</b> {p.description ?? p.description_original ?? '—'}
        </li>
        <li>
          <b>Fecha original:</b> {p.date ?? p.date_original ?? '—'}
        </li>
        <li>
          <b>Monto:</b>{' '}
          {p.amount_minor ? formatSignedMoney(p.amount_minor) : p.amount_original || '—'}
        </li>
      </ul>
      <div className="mt-4 grid gap-3">
        <input
          value={f.description ?? p.description ?? p.description_original ?? ''}
          onChange={(e) => onChange({ description: e.target.value })}
          placeholder="Descripción corregida"
          className="w-full border rounded px-2 py-1 bg-black text-white"
        />
        <div className="flex gap-2">
          <input
            type="date"
            value={f.date ?? p.date ?? p.date_original ?? ''}
            onChange={(e) => onChange({ date: e.target.value })}
            className="w-full border rounded px-2 py-1 bg-black text-white"
          />
          <input
            type="number"
            step="0.01"
            value={f.amount ?? (p.amount_minor !== undefined ? Math.abs(p.amount_minor) / 100 : '')}
            onChange={(e) => onChange({ amount: e.target.value })}
            placeholder="Monto"
            className="w-full border rounded px-2 py-1 bg-black text-white"
          />
        </div>
        <div className="flex gap-2">
          <select
            value={f.type ?? p.type ?? 'EXPENSE'}
            onChange={(e) => onChange({ type: e.target.value })}
            className="w-full border rounded px-2 py-1 bg-black text-white"
          >
            <option value="INCOME">Ingreso</option>
            <option value="EXPENSE">Egreso</option>
          </select>
          <select
            value={f.account ?? p.account_id ?? ''}
            onChange={(e) => onChange({ account: e.target.value })}
            className="w-full border rounded px-2 py-1 bg-black text-white"
          >
            <option value="">Seleccionar cuenta…</option>
            {accounts.map((a) => (
              <option key={a.ID} value={a.ID}>
                {a.Name}
              </option>
            ))}
          </select>
        </div>
        <div className="flex gap-2">
          <button onClick={onSubmit} className="px-3 py-1 rounded bg-accent-strong text-white">
            {isDuplicate ? 'Confirmar y agregar' : 'Confirmar corrección'}
          </button>
          <button onClick={onDiscard} className="px-3 py-1 rounded border text-neg">
            Descartar
          </button>
        </div>
      </div>
    </div>
  );
}
