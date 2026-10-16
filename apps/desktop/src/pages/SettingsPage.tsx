import { useEffect, useState } from 'react';

import { purgeAllData } from '../api/admin';
import { ConfirmModal } from '../components/ConfirmModal';
import { SectionHeader } from '../components/SectionHeader';

/* Configuración — data management. The purge deletes EVERYTHING (accounts,
   documents, movements, events, relations, installments, categories) via
   DELETE /admin/purge. Irreversible: guarded by an explicit confirmation
   that names the exact consequence. */

export default function SettingsPage() {
  const [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Categorías desde backend
  const [categories, setCategories] = useState<any[]>([]);
  const [newCat, setNewCat] = useState({ name: '', type: 'EXPENSE' });
  const [loadingCats, setLoadingCats] = useState(true);

  useEffect(() => {
    fetch('http://localhost:8081/categories')
      .then(r => r.json())
      .then(data => {
        const norm = (data || []).map((c:any) => ({
          id: c.id ?? c.ID,
          name: c.name ?? c.Name,
          type: (c.category_type ?? c.CategoryType ?? '').toUpperCase()
        }));
        setCategories(norm);
      })
      .finally(()=> setLoadingCats(false));
  }, []);

  const addCategory = async () => {
    if (!newCat.name.trim()) return;
    const id = 'cat_' + Date.now();
    const res = await fetch('http://localhost:8081/categories', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({ id, name: newCat.name.trim(), category_type: newCat.type })
    });
    if (res.ok) {
      const cat = await res.json();
      setCategories(c => [...c, cat]);
      setNewCat({ name: '', type: 'EXPENSE' });
      // Auto-refresh to ensure persistence
      setTimeout(()=> window.location.reload(), 300);
    }
  };

  const deleteCategory = async (id: string) => {
    const cat = categories.find(c=> c.id === id);
    const res = await fetch(`http://localhost:8081/categories/${id}`, { method: 'DELETE' });
    if (res.ok) {
      setCategories(c => c.filter(x => x.id !== id));
      // Notificar movimientos asociados para revisión
      try {
        const movRes = await fetch('http://localhost:8081/movements');
        const movs = await movRes.json();
        const affected = movs.filter((m:any) => m.categoryId === id);
        if (affected.length > 0) {
          const queue = JSON.parse(localStorage.getItem('reviewQueue') || '[]');
          const newItems = affected.map((m:any)=> ({
            id: `rev_${m.ID}_${Date.now()}`,
            date: m.MovementDate,
            description: m.Description ?? '',
            categoryId: id,
            categoryName: cat?.name || '',
            categoryType: cat?.type || '',
            movementId: m.ID,
            amount: m.AmountMinor,
            accountId: m.AccountID
          }));
          localStorage.setItem('reviewQueue', JSON.stringify([...queue, ...newItems]));
          const notifs = JSON.parse(localStorage.getItem('notifications') || '[]');
          notifs.push({ id: Date.now(), text: `Categoría "${cat?.name}" eliminada con ${affected.length} movimiento(s) asociados. Revisa en Revisión.` });
          localStorage.setItem('notifications', JSON.stringify(notifs));
        }
      } catch {}
    }
  };

  const confirm = async () => {
    setBusy(true);
    setError(null);
    try {
      await purgeAllData();
      setBusy(false);
      setConfirming(false);
      setDone(true);
      // Every list in the app refetches on mount — a reload presents the
      // emptied system after the success message is legible.
      setTimeout(() => window.location.reload(), 2000);
    } catch {
      setBusy(false);
      setError('No se pudieron borrar los datos. Inténtalo de nuevo.');
    }
  };

  return (
    <div className="max-w-2xl space-y-8">
      <header className="space-y-1">
        <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Configuración</h1>
        <p className="text-body text-ink-2">Gestiona datos y preferencias de la aplicación.</p>
      </header>

      <section aria-label="Categorías" className="space-y-3">
        <SectionHeader title="Categorías" />
        <div className="rounded-[12px] border border-line bg-surface p-5 space-y-4">
          <div className="flex gap-2">
            <input value={newCat.name} onChange={e=>setNewCat({ ...newCat, name:e.target.value })} placeholder="Nombre categoría" className="flex-1 bg-bg border border-line rounded px-2 py-1 text-ink" />
            <select value={newCat.type} onChange={e=>setNewCat({ ...newCat, type:e.target.value })} className="bg-bg border border-line rounded px-2 py-1 text-ink">
              <option value="EXPENSE">Egreso</option>
              <option value="INCOME">Ingreso</option>
            </select>
            <button onClick={addCategory} className="px-3 py-1 rounded bg-accent-strong text-white">Agregar</button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h3 className="text-sm font-medium text-ink-2 mb-2">Categorías de Ingresos</h3>
              <ul className="space-y-2">
                {categories.filter(c=>c.type==='INCOME').map(c => (
                  <li key={c.id} className="flex items-center justify-between border border-line rounded px-3 py-2">
                    <span>{c.name}</span>
                    <button onClick={()=>deleteCategory(c.id)} className="text-neg text-sm">Eliminar</button>
                  </li>
                ))}
                {categories.filter(c=>c.type==='INCOME').length===0 && <li className="text-meta text-ink-3 text-sm">Sin categorías</li>}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-medium text-ink-2 mb-2">Categorías de Egresos</h3>
              <ul className="space-y-2">
                {categories.filter(c=>c.type==='EXPENSE').map(c => (
                  <li key={c.id} className="flex items-center justify-between border border-line rounded px-3 py-2">
                    <span>{c.name}</span>
                    <button onClick={()=>deleteCategory(c.id)} className="text-neg text-sm">Eliminar</button>
                  </li>
                ))}
                {categories.filter(c=>c.type==='EXPENSE').length===0 && <li className="text-meta text-ink-3 text-sm">Sin categorías</li>}
              </ul>
            </div>
          </div>
          <p className="text-meta text-ink-3">Si una categoría es editada/eliminada y tiene movimientos asociados, esos movimientos pasarán a Revisión.</p>
        </div>
      </section>

      <section aria-label="Datos" className="space-y-3">
        <SectionHeader title="Datos" />
        <div className="rounded-[12px] border border-line bg-surface p-5">
          <p className="text-body text-ink-2">
            Borrar todos los datos elimina de forma permanente las cuentas,
            documentos, movimientos, eventos, relaciones y cuotas registradas
            en FinanceApp. Esta acción no se puede deshacer.
          </p>
          <button
            onClick={() => {
              setConfirming(true);
              setError(null);
            }}
            className="mt-4 inline-flex h-8 items-center rounded-[8px] border border-neg px-3 text-body font-medium text-neg transition-colors duration-[120ms] hover:bg-neg/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-neg"
          >
            Borrar todos los datos
          </button>
          {done && (
            <p className="mt-3 text-body text-pos" role="status">
              Todos los datos fueron eliminados.
            </p>
          )}
          {error && (
            <p className="mt-3 text-body text-neg" role="alert">
              {error}
            </p>
          )}
        </div>
      </section>

      {confirming && (
        <ConfirmModal
          title="¿Borrar todos los datos?"
          message="Se eliminarán todas las cuentas, documentos, movimientos y eventos de FinanceApp. Esta acción no se puede deshacer."
          confirmLabel="Borrar todos los datos"
          onConfirm={confirm}
          onCancel={() => setConfirming(false)}
          busy={busy}
        />
      )}
    </div>
  );
}
