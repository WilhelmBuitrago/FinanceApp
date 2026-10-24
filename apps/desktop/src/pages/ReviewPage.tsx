import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getAccounts } from '../api/accounts';
import { getCategories } from '../api/categories';
import { ApiError } from '../api/client';
import {
  discardReviewEntry,
  getReviewEntries,
  resolveReviewCategory,
  resolveReviewMovement,
  resolveReviewPeriod,
  type ReviewItem,
} from '../api/review';
import { formatDate } from '../lib/formatDate';
import { formatSignedMoney } from '../lib/formatMoney';

/* Human review. One queue, three triggers: DUPLICATE (a movement that
   collided), CATEGORY (movements left without a category) and EXTRACTION (a
   movement the processor could not normalize). The queue lives in the backend
   (review_items), so a resolution survives restarts and every page reads the
   same source. */

const MONTHS = [
  { n: 1, label: 'Enero' }, { n: 2, label: 'Febrero' }, { n: 3, label: 'Marzo' },
  { n: 4, label: 'Abril' }, { n: 5, label: 'Mayo' }, { n: 6, label: 'Junio' },
  { n: 7, label: 'Julio' }, { n: 8, label: 'Agosto' }, { n: 9, label: 'Septiembre' },
  { n: 10, label: 'Octubre' }, { n: 11, label: 'Noviembre' }, { n: 12, label: 'Diciembre' },
];

function payloadOf(item: ReviewItem): any {
  try { return JSON.parse(item.PayloadJson || '{}'); } catch { return {}; }
}

export default function ReviewPage() {
  const [items, setItems] = useState<ReviewItem[]>([]);
  const [categories, setCategories] = useState<any[]>([]);
  const [accounts, setAccounts] = useState<any[]>([]);
  const [form, setForm] = useState<Record<string, any>>({});
  const [selectedCat, setSelectedCat] = useState<Record<string, string>>({});

  const loadQueue = () =>
    getReviewEntries()
      .then(d => setItems(d || []))
      .catch(() => setItems([]));

  useEffect(() => {
    loadQueue();
    getCategories()
      .then(d => setCategories((d || []).map(c => ({
        id: c.ID, name: c.Name,
        type: (c.CategoryType ?? '').toUpperCase(),
      }))))
      .catch(() => {});
    getAccounts()
      .then(d => setAccounts(d || []))
      .catch(() => {});
  }, []);

  const setField = (id: string, patch: any) =>
    setForm(f => ({ ...f, [id]: { ...(f[id] || {}), ...patch } }));

  // Refreshing the queue in place also tells the sidebar badge to reload.
  const dropItem = (id: string) => {
    setItems(its => its.filter(i => i.ID !== id));
    window.dispatchEvent(new Event('review-queue-changed'));
  };

  const resolveMovement = async (item: ReviewItem) => {
    const p = payloadOf(item);
    const f = form[item.ID] || {};
    const description = f.description ?? p.description ?? p.description_original ?? '';
    const date = f.date ?? p.date ?? p.date_original ?? '';
    const amountMinor = f.amount !== undefined && f.amount !== ''
      ? Math.round(Number(f.amount) * 100) * (f.type === 'INCOME' ? 1 : -1)
      : (p.amount_minor ?? 0);
    const type = f.type ?? p.type ?? 'EXPENSE';
    const accountId = f.account ?? p.account_id ?? '';
    if (!date || !accountId || !type) {
      alert('Completá fecha, monto y cuenta para confirmar');
      return;
    }
    try {
      await resolveReviewMovement(item.ID, {
        description, amount_minor: amountMinor, date, type, account_id: accountId,
      });
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        alert('Este ítem ya fue resuelto');
        dropItem(item.ID);
      } else {
        alert('Error al confirmar la corrección');
      }
    }
  };

  // Document-level resolution: the human supplies the statement month and
  // year (ABR 2026); the backend computes the full period and derives the
  // parked movement dates of that document against it.
  const resolvePeriod = async (item: ReviewItem) => {
    const f = form[item.ID] || {};
    const month = Number(f.month);
    const year = Number(f.year);
    if (!month || !year) {
      alert('Elegí el mes y el año del estado de cuenta');
      return;
    }
    try {
      await resolveReviewPeriod(item.ID, month, year);
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        alert('Este ítem ya fue resuelto');
        dropItem(item.ID);
      } else {
        alert('Error al confirmar el período');
      }
    }
  };

  const resolveCategory = async (item: ReviewItem, categoryId: string | null) => {
    try {
      await resolveReviewCategory(item.ID, categoryId);
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        dropItem(item.ID);
      } else {
        alert('Error al resolver la categoría');
      }
    }
  };

  const discard = async (item: ReviewItem) => {
    try {
      await discardReviewEntry(item.ID);
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        dropItem(item.ID);
      }
    }
  };

  if (items.length === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Revisión</h1>
        <div className="rounded-[12px] border border-dashed border-line bg-surface/50 p-12 text-center">
          <p className="text-body text-ink-2">No hay ítems pendientes de revisión.</p>
          <p className="mt-1 text-meta text-ink-3">Cuando un documento, un duplicado o una categoría requiera atención aparecerá aquí.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Revisión</h1>
      {items.map((item) => {
        const p = payloadOf(item);
        const f = form[item.ID] || {};

        if (item.Kind === 'CATEGORY') {
          const cats = categories.filter(c => c.type === (p.category_type || 'EXPENSE'));
          return (
            <div key={item.ID} className="rounded-[12px] border border-line bg-surface p-6">
              <p className="text-body font-medium">Movimiento sin categoría</p>
              <p className="mt-1 text-meta text-ink-3">La categoría "{p.category_name || '—'}" fue eliminada.</p>
              <ul className="mt-3 space-y-1 text-body">
                <li><b>Fecha:</b> {p.movement_date ? formatDate(p.movement_date) : '—'}</li>
                <li><b>Monto:</b> {p.amount_minor ? formatSignedMoney(p.amount_minor) : '—'}</li>
              </ul>
              <div className="mt-4 flex flex-wrap gap-2">
                <select
                  value={selectedCat[item.ID] || ''}
                  onChange={e => setSelectedCat(s => ({ ...s, [item.ID]: e.target.value }))}
                  className="w-full max-w-xs border rounded px-2 py-1 bg-black text-white"
                >
                  <option value="">Seleccionar categoría…</option>
                  {cats.map(c => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
                <button onClick={() => resolveCategory(item, selectedCat[item.ID] || null)}
                  className="px-3 py-1 rounded bg-accent-strong text-white">
                  {selectedCat[item.ID] ? 'Asignar categoría' : 'Guardar sin categoría'}
                </button>
                <button onClick={() => discard(item)} className="px-3 py-1 rounded border text-neg">Descartar</button>
              </div>
            </div>
          );
        }

        if (item.Kind === 'DOCUMENT_PERIOD') {
          const fPeriod = form[item.ID] || {};
          const suggestedMonth = fPeriod.month ?? (p.suggested_month ? String(p.suggested_month) : '');
          const suggestedYear = fPeriod.year ?? (p.suggested_year ? String(p.suggested_year) : '');
          return (
            <div key={item.ID} className="rounded-[12px] border border-line bg-surface p-6">
              <p className="text-body font-medium">Período del documento no reconocido</p>
              <p className="mt-1 text-meta text-ink-3">
                El extracto no traía un período legible
                {p.raw_period ? ` (“${p.raw_period}”)` : ''}. Elegí el mes y el año del
                estado de cuenta: con eso se calcula el período completo y se derivan las
                fechas de los movimientos estacionados de este documento contra él.
              </p>
              <div className="mt-4 flex flex-wrap gap-2">
                <select
                  value={suggestedMonth}
                  onChange={e => setField(item.ID, { month: e.target.value })}
                  className="border rounded px-2 py-1 bg-black text-white"
                >
                  <option value="">Mes…</option>
                  {MONTHS.map(m => (
                    <option key={m.n} value={m.n}>{m.label}</option>
                  ))}
                </select>
                <input
                  type="number"
                  min={2000}
                  max={2100}
                  value={suggestedYear}
                  onChange={e => setField(item.ID, { year: e.target.value })}
                  placeholder="Año"
                  className="w-28 border rounded px-2 py-1 bg-black text-white"
                />
                <button
                  onClick={() => resolvePeriod(item)}
                  disabled={!suggestedMonth || !suggestedYear}
                  className="px-3 py-1 rounded bg-accent-strong text-white disabled:opacity-40"
                >
                  Confirmar período
                </button>
                <button onClick={() => discard(item)} className="px-3 py-1 rounded border text-neg">Descartar</button>
              </div>
              {(p.suggested_month || p.suggested_year) && (
                <p className="mt-2 text-meta text-ink-3">
                  Mes y año precargados a partir de las fechas de los movimientos estacionados
                  de este documento: es una sugerencia, no un dato confirmado. Verificalos
                  contra el PDF antes de confirmar.
                </p>
              )}
            </div>
          );
        }

        const isDuplicate = item.Kind === 'DUPLICATE';
        return (
          <div key={item.ID} className="rounded-[12px] border border-line bg-surface p-6">
            <p className="text-body font-medium">
              {isDuplicate ? 'Movimiento duplicado detectado' : 'Extracción requiere verificación'}
            </p>
            <p className="mt-1 text-meta text-ink-3">
              {isDuplicate
                ? 'Ya existe un movimiento con estos datos. Confirmá si corresponde crear uno nuevo.'
                : (p.reason || 'Movimiento no normalizado')}
            </p>
            <ul className="mt-3 space-y-1 text-body">
              <li><b>Descripción original:</b> {p.description ?? p.description_original ?? '—'}</li>
              <li><b>Fecha original:</b> {p.date ?? p.date_original ?? '—'}</li>
              <li><b>Monto:</b> {p.amount_minor ? formatSignedMoney(p.amount_minor) : (p.amount_original || '—')}</li>
            </ul>
            <div className="mt-4 grid gap-3">
              <input
                value={f.description ?? p.description ?? p.description_original ?? ''}
                onChange={e => setField(item.ID, { description: e.target.value })}
                placeholder="Descripción corregida"
                className="w-full border rounded px-2 py-1 bg-black text-white"
              />
              <div className="flex gap-2">
                <input
                  type="date"
                  value={f.date ?? p.date ?? p.date_original ?? ''}
                  onChange={e => setField(item.ID, { date: e.target.value })}
                  className="w-full border rounded px-2 py-1 bg-black text-white"
                />
                <input
                  type="number"
                  step="0.01"
                  value={f.amount ?? (p.amount_minor !== undefined ? Math.abs(p.amount_minor) / 100 : '')}
                  onChange={e => setField(item.ID, { amount: e.target.value })}
                  placeholder="Monto"
                  className="w-full border rounded px-2 py-1 bg-black text-white"
                />
              </div>
              <div className="flex gap-2">
                <select
                  value={f.type ?? p.type ?? 'EXPENSE'}
                  onChange={e => setField(item.ID, { type: e.target.value })}
                  className="w-full border rounded px-2 py-1 bg-black text-white"
                >
                  <option value="INCOME">Ingreso</option>
                  <option value="EXPENSE">Egreso</option>
                </select>
                <select
                  value={f.account ?? p.account_id ?? ''}
                  onChange={e => setField(item.ID, { account: e.target.value })}
                  className="w-full border rounded px-2 py-1 bg-black text-white"
                >
                  <option value="">Seleccionar cuenta…</option>
                  {accounts.map((a: any) => (
                    <option key={a.ID || a.id} value={a.ID || a.id}>{a.Name || a.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex gap-2">
                <button onClick={() => resolveMovement(item)} className="px-3 py-1 rounded bg-accent-strong text-white">
                  {isDuplicate ? 'Confirmar y agregar' : 'Confirmar corrección'}
                </button>
                <button onClick={() => discard(item)} className="px-3 py-1 rounded border text-neg">Descartar</button>
              </div>
            </div>
          </div>
        );
      })}
      <Link to="/" className="px-3 py-1 rounded border inline-block">Volver</Link>
    </div>
  );
}
