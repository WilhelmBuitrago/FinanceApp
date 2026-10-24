import { formatDate } from '../../lib/formatDate';
import { formatSignedMoney } from '../../lib/formatMoney';

/* CATEGORY card: a movement whose category was deleted. The human either
   assigns one of their categories or leaves the movement uncategorized —
   both are valid decisions (a missing category never forces the item). */

interface Props {
  payload: any;
  categories: any[];
  selectedId: string;
  onSelectChange: (value: string) => void;
  onAssign: () => void;
  onDiscard: () => void;
}

export function ReviewCategoryCard({ payload: p, categories, selectedId, onSelectChange, onAssign, onDiscard }: Props) {
  const cats = categories.filter(c => c.type === (p.category_type || 'EXPENSE'));
  return (
    <div className="rounded-[12px] border border-line bg-surface p-6">
      <p className="text-body font-medium">Movimiento sin categoría</p>
      <p className="mt-1 text-meta text-ink-3">La categoría "{p.category_name || '—'}" fue eliminada.</p>
      <ul className="mt-3 space-y-1 text-body">
        <li><b>Fecha:</b> {p.movement_date ? formatDate(p.movement_date) : '—'}</li>
        <li><b>Monto:</b> {p.amount_minor ? formatSignedMoney(p.amount_minor) : '—'}</li>
      </ul>
      <div className="mt-4 flex flex-wrap gap-2">
        <select
          value={selectedId}
          onChange={e => onSelectChange(e.target.value)}
          className="w-full max-w-xs border rounded px-2 py-1 bg-black text-white"
        >
          <option value="">Seleccionar categoría…</option>
          {cats.map(c => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
        <button onClick={onAssign}
          className="px-3 py-1 rounded bg-accent-strong text-white">
          {selectedId ? 'Asignar categoría' : 'Guardar sin categoría'}
        </button>
        <button onClick={onDiscard} className="px-3 py-1 rounded border text-neg">Descartar</button>
      </div>
    </div>
  );
}
