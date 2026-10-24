import { useEffect, useState } from 'react';

import { ApiError } from '../api/client';
import { createMovement } from '../api/movements';
import { getCategories } from '../api/categories';
import type { Account } from '../types/account';

/* Add-movement drawer: manual capture. A zero or negative amount is
   rejected client-side with the same criterion the backend applies, the
   category list is the user's own filtered by the chosen type, and a
   colliding movement (409) is parked by the backend in the review queue:
   the drawer only reports it through onDuplicate. */

interface Props {
  accounts: Account[];
  onClose: () => void;
  onDuplicate: (text: string) => void;
}

export function AddMovementDrawer({ accounts, onClose, onDuplicate }: Props) {
  const [form, setForm] = useState({
    cantidad: '',
    tipo: 'Ingreso',
    cuenta: '',
    categoria: '',
    descripcion: '',
    fecha: '',
  });
  const [categories, setCategories] = useState<any[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (accounts.length > 0 && !form.cuenta) {
      setForm((f) => ({ ...f, cuenta: accounts[0].ID }));
    }
  }, [accounts]);

  useEffect(() => {
    getCategories()
      .then((data) => {
        const norm = (data || []).map((c) => ({
          id: c.ID,
          name: c.Name,
          category_type: (c.CategoryType ?? '').toUpperCase(),
        }));
        setCategories(norm);
      })
      .catch(() => {});
  }, []);

  return (
    <div className="fixed inset-0 z-[10000] flex">
      <div className="absolute inset-0 bg-black/40 z-0" onClick={onClose} />
      <div
        className="ml-auto h-full w-[420px] bg-surface border-l border-line p-6 overflow-auto relative z-10"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Agregar movimiento</h2>
        <form
          onSubmit={async (e) => {
            e.preventDefault();
            setError(null);
            const amount = Number(parseFloat(form.cantidad).toFixed(2));
            if (!amount || amount <= 0) {
              setError('Cantidad debe ser positiva');
              return;
            }
            const sign = form.tipo === 'Egreso' ? -1 : 1;
            const amountMinor = Math.round(amount * 100) * sign;
            try {
              await createMovement({
                amount_minor: amountMinor,
                type:
                  form.tipo === 'Ingreso'
                    ? 'INCOME'
                    : form.tipo === 'Egreso'
                      ? 'EXPENSE'
                      : 'TRANSFER',
                account_id: form.cuenta,
                description: form.descripcion,
                date: form.fecha,
                category_id: form.categoria,
              });
              window.location.reload();
            } catch (err) {
              if (err instanceof ApiError && err.status === 409) {
                // The backend parks the duplicate in the review queue; the UI
                // only notifies (no client-side queue to lose).
                onDuplicate(
                  `Movimiento duplicado detectado el ${form.fecha} por $${amount}. Quedó en Revisión.`,
                );
                onClose();
                return;
              }
              // Same copy as before: the raw server body when there is one.
              const detail =
                err instanceof ApiError && typeof err.response === 'string'
                  ? err.response
                  : err instanceof Error
                    ? err.message
                    : String(err);
              setError(detail);
            }
          }}
        >
          {error && (
            <p className="text-body text-neg" role="alert">
              {error}
            </p>
          )}
          <div className="space-y-4">
            <div>
              <label className="text-sm" htmlFor="add-movement-amount">
                Cantidad *
              </label>
              <input
                id="add-movement-amount"
                type="number"
                min="0"
                step="0.01"
                required
                value={form.cantidad}
                onChange={(e) => setForm({ ...form, cantidad: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              />
            </div>
            <div>
              <label className="text-sm" htmlFor="add-movement-type">
                Tipo *
              </label>
              <select
                id="add-movement-type"
                value={form.tipo}
                onChange={(e) => setForm({ ...form, tipo: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              >
                <option>Ingreso</option>
                <option>Egreso</option>
                {accounts.length > 1 && <option>Transferencia</option>}
              </select>
            </div>
            <div>
              <label className="text-sm" htmlFor="add-movement-account">
                Cuenta *
              </label>
              <select
                id="add-movement-account"
                value={form.cuenta}
                onChange={(e) => setForm({ ...form, cuenta: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              >
                {accounts.map((a) => (
                  <option key={a.ID} value={a.ID}>
                    {a.Name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm" htmlFor="add-movement-category">
                Categoría
              </label>
              <select
                id="add-movement-category"
                value={form.categoria}
                onChange={(e) => setForm({ ...form, categoria: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              >
                <option value="">Sin categoría</option>
                {categories
                  .filter((c) => {
                    const typeMap: Record<string, string> = {
                      Ingreso: 'INCOME',
                      Egreso: 'EXPENSE',
                      Transferencia: 'TRANSFER',
                    };
                    return c.category_type === typeMap[form.tipo];
                  })
                  .map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
              </select>
            </div>
            <div>
              <label className="text-sm" htmlFor="add-movement-description">
                Descripción
              </label>
              <input
                id="add-movement-description"
                type="text"
                value={form.descripcion}
                onChange={(e) => setForm({ ...form, descripcion: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              />
            </div>
            <div>
              <label className="text-sm" htmlFor="add-movement-date">
                Fecha *
              </label>
              <input
                id="add-movement-date"
                type="date"
                required
                value={form.fecha}
                onChange={(e) => setForm({ ...form, fecha: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              />
            </div>
            <div className="flex justify-end gap-2 pt-4">
              <button type="button" onClick={onClose} className="px-3 py-1 rounded border">
                Cancelar
              </button>
              <button type="submit" className="px-3 py-1 rounded bg-accent-strong text-white">
                Aceptar
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
