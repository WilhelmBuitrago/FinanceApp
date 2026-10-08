import React, { useState } from 'react';

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
      <section aria-label="Datos" className="space-y-2">
        <SectionHeader title="Datos" />
        <div className="rounded-tk border border-line bg-surface p-4">
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
            className="mt-4 inline-flex h-8 items-center rounded-tk border border-neg px-3 text-body font-medium text-neg transition-colors duration-[120ms] hover:bg-neg/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-neg"
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
