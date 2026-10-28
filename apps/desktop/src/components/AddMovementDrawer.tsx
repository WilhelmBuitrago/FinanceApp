import { useEffect, useState } from 'react';

import { ApiError, apiErrorDetail } from '../api/client';
import { invalidateCollections } from '../lib/dataBus';
import { createMovement } from '../api/movements';
import { getCategories } from '../api/categories';
import { formatMoney } from '../lib/formatMoney';
import { parseAmountMinor } from '../lib/amount';
import type { Category, CategoryOption } from '../types/category';
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
    destino: '',
    categoria: '',
    descripcion: '',
    fecha: '',
  });
  const [categories, setCategories] = useState<CategoryOption[]>([]);
  const [error, setError] = useState<string | null>(null);
  // The effective account: the picked one, or the first available. Derived
  // during render instead of synced through an effect (the effect version
  // is what react-hooks flags, and it buys nothing).
  const effectiveCuenta = form.cuenta || accounts[0]?.ID || '';

  useEffect(() => {
    getCategories()
      .then((data) => {
        const norm = (data || []).map((c: Category) => ({
          id: c.ID,
          name: c.Name,
          type: (c.CategoryType ?? '').toUpperCase(),
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
            // parseAmountMinor reads the amount the way a Colombian writes
            // it ("25,5" is twenty-five and a half) and hands the integer
            // centavos the ledger speaks; parseFloat read the comma away and
            // silently divided the amount by ten.
            const amountMinor = parseAmountMinor(form.cantidad);
            if (amountMinor === null || amountMinor <= 0) {
              setError('Cantidad debe ser positiva');
              return;
            }
            // A transfer moves money between two own accounts: the backend
            // derives the signs (it leaves the origin, it enters the
            // destination), so the amount stays positive here.
            const isTransfer = form.tipo === 'Transferencia';
            if (isTransfer) {
              if (!form.destino) {
                setError('Elegí la cuenta destino');
                return;
              }
              if (form.destino === effectiveCuenta) {
                setError('La cuenta destino debe ser distinta a la de origen');
                return;
              }
            }
            const sign = form.tipo === 'Egreso' ? -1 : 1;
            const signedMinor = amountMinor * sign;
            try {
              await createMovement({
                amount_minor: signedMinor,
                type: isTransfer ? 'TRANSFER' : form.tipo === 'Ingreso' ? 'INCOME' : 'EXPENSE',
                account_id: effectiveCuenta,
                ...(isTransfer ? { to_account_id: form.destino } : {}),
                description: form.descripcion,
                date: form.fecha,
                category_id: isTransfer ? '' : form.categoria,
              });
              // Every screen showing movements or their aggregates refetches
              // through the bus; no document reload.
              invalidateCollections('movements', 'dashboard');
            } catch (err) {
              if (err instanceof ApiError && err.status === 409) {
                // The backend parks the duplicate in the review queue; the UI
                // only notifies (no client-side queue to lose).
                onDuplicate(
                  `Movimiento duplicado detectado el ${form.fecha} por ${formatMoney(signedMinor)}. Quedó en Revisión.`,
                );
                onClose();
                return;
              }
              // Same copy as before: the raw server body when there is one.
              const detail =
                apiErrorDetail(err) ?? (err instanceof Error ? err.message : String(err));
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
              <label className="text-body" htmlFor="add-movement-amount">
                Cantidad *
              </label>
              <input
                id="add-movement-amount"
                type="text"
                inputMode="decimal"
                required
                value={form.cantidad}
                onChange={(e) => setForm({ ...form, cantidad: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              />
            </div>
            <div>
              <label className="text-body" htmlFor="add-movement-type">
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
              <label className="text-body" htmlFor="add-movement-account">
                {form.tipo === 'Transferencia' ? 'Cuenta origen *' : 'Cuenta *'}
              </label>
              <select
                id="add-movement-account"
                value={effectiveCuenta}
                onChange={(e) => setForm({ ...form, cuenta: e.target.value })}
                className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
              >
                {accounts.map((a) => (
                  <option key={a.ID} value={a.ID}>
                    {a.Name} ({a.Currency ?? 'COP'})
                  </option>
                ))}
              </select>
            </div>
            {form.tipo === 'Transferencia' && (
              <div>
                <label className="text-body" htmlFor="add-movement-destination">
                  Cuenta destino *
                </label>
                <select
                  id="add-movement-destination"
                  value={form.destino}
                  onChange={(e) => setForm({ ...form, destino: e.target.value })}
                  className="w-full bg-surface border border-line rounded px-2 py-1 text-ink"
                >
                  <option value="">Elegí una cuenta</option>
                  {accounts
                    .filter((a) => a.ID !== effectiveCuenta)
                    .map((a) => (
                      <option key={a.ID} value={a.ID}>
                        {a.Name} ({a.Currency ?? 'COP'})
                      </option>
                    ))}
                </select>
              </div>
            )}
            {form.tipo !== 'Transferencia' && (
              <div>
                <label className="text-body" htmlFor="add-movement-category">
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
                      return c.type === typeMap[form.tipo];
                    })
                    .map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                </select>
              </div>
            )}
            <div>
              <label className="text-body" htmlFor="add-movement-description">
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
              <label className="text-body" htmlFor="add-movement-date">
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
