import { useEffect, useState } from 'react';

import { purgeAllData } from '../api/admin';
import {
  createCategory,
  deleteCategory as deleteCategoryApi,
  getCategories,
} from '../api/categories';
import { ConfirmModal } from '../components/ConfirmModal';
import { SectionHeader } from '../components/SectionHeader';
import { check } from '@tauri-apps/plugin-updater';

/* Configuración — data management. The purge deletes EVERYTHING (accounts,
   documents, movements, events, relations, installments, categories) via
   DELETE /admin/purge. Irreversible: guarded by an explicit confirmation
   that names the exact consequence. */

export default function SettingsPage() {
  const [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Actualizaciones
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [updateAvailable, setUpdateAvailable] = useState<any>(null);
  const [updateConfirming, setUpdateConfirming] = useState(false);
  const [updateProgress, setUpdateProgress] = useState<string | null>(null);
  const [updateError, setUpdateError] = useState<string | null>(null);

  // Categorías desde backend
  const [categories, setCategories] = useState<any[]>([]);
  const [newCat, setNewCat] = useState({ name: '', type: 'EXPENSE' });

  useEffect(() => {
    getCategories()
      .then((data) => {
        const norm = (data || []).map((c) => ({
          id: c.ID,
          name: c.Name,
          type: (c.CategoryType ?? '').toUpperCase(),
        }));
        setCategories(norm);
      })
      .catch(() => {});
  }, []);

  const addCategory = async () => {
    if (!newCat.name.trim()) return;
    const id = 'cat_' + Date.now();
    try {
      const cat = await createCategory(id, newCat.name.trim(), newCat.type);
      setCategories((c) => [...c, cat]);
      setNewCat({ name: '', type: 'EXPENSE' });
      // Auto-refresh to ensure persistence
      setTimeout(() => window.location.reload(), 300);
    } catch {
      // A failed create leaves the list untouched, exactly as before.
    }
  };

  const deleteCategory = async (id: string) => {
    try {
      await deleteCategoryApi(id);
      setCategories((c) => c.filter((x) => x.id !== id));
      // The backend parks the affected movements in the unified review queue;
      // the UI only refreshes the list.
    } catch {
      // A failed delete leaves the list untouched, exactly as before.
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

  const handleCheckUpdates = async () => {
    setCheckingUpdate(true);
    setUpdateError(null);
    setUpdateProgress(null);
    try {
      const update = await check();
      if (update) {
        setUpdateAvailable(update);
        setUpdateConfirming(true);
      } else {
        setUpdateProgress('FinanceApp está actualizado.');
        setTimeout(() => setUpdateProgress(null), 3000);
      }
    } catch (e: any) {
      setUpdateError(e?.message || 'Error verificando actualizaciones.');
    } finally {
      setCheckingUpdate(false);
    }
  };

  const handleInstallUpdate = async () => {
    if (!updateAvailable) return;
    setUpdateConfirming(false);
    setUpdateProgress(`Descargando versión ${updateAvailable.version}...`);
    try {
      // El plugin permite callback de progreso
      await updateAvailable.downloadAndInstall(
        (event: { data?: { progress?: { percent?: number } } }) => {
          const percent = event?.data?.progress?.percent ?? 0;
          if (percent) {
            setUpdateProgress(
              `Descargando ${updateAvailable.version}... ${Math.round(percent * 100)}%`,
            );
          }
        },
      );
      setUpdateProgress('Instalación completada. Reiniciando...');
    } catch (e: any) {
      setUpdateError(e?.message || 'Falló la instalación de la actualización.');
      setUpdateProgress(null);
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
            <input
              value={newCat.name}
              onChange={(e) => setNewCat({ ...newCat, name: e.target.value })}
              placeholder="Nombre categoría"
              className="flex-1 bg-bg border border-line rounded px-2 py-1 text-ink"
            />
            <select
              value={newCat.type}
              onChange={(e) => setNewCat({ ...newCat, type: e.target.value })}
              className="bg-bg border border-line rounded px-2 py-1 text-ink"
            >
              <option value="EXPENSE">Egreso</option>
              <option value="INCOME">Ingreso</option>
            </select>
            <button onClick={addCategory} className="px-3 py-1 rounded bg-accent-strong text-white">
              Agregar
            </button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h3 className="text-sm font-medium text-ink-2 mb-2">Categorías de Ingresos</h3>
              <ul className="space-y-2">
                {categories
                  .filter((c) => c.type === 'INCOME')
                  .map((c) => (
                    <li
                      key={c.id}
                      className="flex items-center justify-between border border-line rounded px-3 py-2"
                    >
                      <span>{c.name}</span>
                      <button onClick={() => deleteCategory(c.id)} className="text-neg text-sm">
                        Eliminar
                      </button>
                    </li>
                  ))}
                {categories.filter((c) => c.type === 'INCOME').length === 0 && (
                  <li className="text-meta text-ink-3 text-sm">Sin categorías</li>
                )}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-medium text-ink-2 mb-2">Categorías de Egresos</h3>
              <ul className="space-y-2">
                {categories
                  .filter((c) => c.type === 'EXPENSE')
                  .map((c) => (
                    <li
                      key={c.id}
                      className="flex items-center justify-between border border-line rounded px-3 py-2"
                    >
                      <span>{c.name}</span>
                      <button onClick={() => deleteCategory(c.id)} className="text-neg text-sm">
                        Eliminar
                      </button>
                    </li>
                  ))}
                {categories.filter((c) => c.type === 'EXPENSE').length === 0 && (
                  <li className="text-meta text-ink-3 text-sm">Sin categorías</li>
                )}
              </ul>
            </div>
          </div>
          <p className="text-meta text-ink-3">
            Si una categoría es editada/eliminada y tiene movimientos asociados, esos movimientos
            pasarán a Revisión.
          </p>
        </div>
      </section>

      <section aria-label="Datos" className="space-y-3">
        <SectionHeader title="Datos" />
        <div className="rounded-[12px] border border-line bg-surface p-5">
          <p className="text-body text-ink-2">
            Borrar todos los datos elimina de forma permanente las cuentas, documentos, movimientos,
            eventos, relaciones y cuotas registradas en FinanceApp. Esta acción no se puede
            deshacer.
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

      <section aria-label="Actualizaciones" className="space-y-3">
        <SectionHeader title="Actualizaciones" />
        <div className="rounded-[12px] border border-line bg-surface p-5 space-y-3">
          <p className="text-body text-ink-2">
            Verifica si hay una nueva versión de FinanceApp disponible.
          </p>
          <button
            onClick={handleCheckUpdates}
            disabled={checkingUpdate}
            className="inline-flex h-8 items-center rounded-[8px] bg-violet-600 px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-violet-700 disabled:opacity-60"
          >
            {checkingUpdate ? 'Buscando...' : 'Buscar Actualizaciones'}
          </button>
          {updateProgress && (
            <p className="text-body text-ink-2" role="status">
              {updateProgress}
            </p>
          )}
          {updateError && (
            <p className="text-body text-neg" role="alert">
              {updateError}
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

      {updateConfirming && updateAvailable && (
        <ConfirmModal
          title="Actualización disponible"
          message={`Se encontró la versión ${updateAvailable.version}. ¿Deseas instalarla? La aplicación se reiniciará al finalizar.`}
          confirmLabel="Instalar"
          onConfirm={handleInstallUpdate}
          onCancel={() => setUpdateConfirming(false)}
          busy={!!updateProgress}
        />
      )}
    </div>
  );
}
