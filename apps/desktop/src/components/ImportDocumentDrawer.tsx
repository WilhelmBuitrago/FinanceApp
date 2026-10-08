import React, { useState } from 'react';
import { Link } from 'react-router-dom';

import { uploadDocument } from '../api/documents';
import { ApiError } from '../api/client';
import {
  DrawerBody,
  DrawerCloseButton,
  DrawerHeader,
  DrawerSectionLabel,
  DrawerShell,
} from './DrawerShell';
import { FormField, selectClass } from './FormField';
import type { Account } from '../types/account';

/* Import document drawer — an import ALWAYS belongs to a configured
   account (S12: no orphan documents). With one account it is shown fixed;
   with several, it is requested explicitly. No account → the configuration
   path, never an empty selector. */

interface ImportDocumentDrawerProps {
  accounts: Account[];
  onClose: () => void;
  onImported: () => void;
}

export function ImportDocumentDrawer({ accounts, onClose, onImported }: ImportDocumentDrawerProps) {
  const [accountId, setAccountId] = useState(accounts[0]?.ID ?? '');
  const [entityId, setEntityId] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    if (!accountId || !file) return;
    setSubmitting(true);
    setError(null);
    /* Optional entity: one of the user's registered accounts. The backend
       verifies the value against the target account's institution, so we
       send the selected entity-account's institution. */
    const entityAccount = accounts.find((a) => a.ID === entityId);
    const entity = entityAccount?.Institution ?? undefined;
    try {
      await uploadDocument(accountId, file, entity);
      onImported();
    } catch (err) {
      const detail =
        err instanceof ApiError && typeof err.response === 'object' && err.response !== null
          ? (err.response as { error?: string }).error
          : undefined;
      setError(detail ?? 'No se pudo importar el documento.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <DrawerShell ariaLabel="Importar documento" onClose={onClose}>
      <DrawerHeader>
        <p className="text-title font-semibold text-ink">Importar documento</p>
        <DrawerCloseButton onClose={onClose} />
      </DrawerHeader>

      <DrawerBody>
        {accounts.length === 0 ? (
          <div className="space-y-4">
            <p className="text-body text-ink-2">Primero necesitas configurar una cuenta.</p>
            <Link
              to="/accounts"
              onClick={onClose}
              className="inline-flex h-8 items-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
            >
              Crear cuenta
            </Link>
          </div>
        ) : (
          <>
            <section className="space-y-2">
              <DrawerSectionLabel>Cuenta</DrawerSectionLabel>
              {accounts.length === 1 ? (
                <p className="rounded-tk border border-line bg-bg px-3 py-2 text-body text-ink">
                  {accounts[0].Name}
                  <span className="ml-2 text-meta text-ink-3">
                    {accounts[0].Institution ?? ''} · {accounts[0].Currency}
                  </span>
                </p>
              ) : (
                <select
                  value={accountId}
                  onChange={(e) => setAccountId(e.target.value)}
                  aria-label="Cuenta"
                  className={selectClass}
                >
                  {accounts.map((a) => (
                    <option key={a.ID} value={a.ID}>
                      {a.Name}
                    </option>
                  ))}
                </select>
              )}
            </section>

            <section className="space-y-2">
              <DrawerSectionLabel>Entidad asociada</DrawerSectionLabel>
              <FormField
                label="Entidad"
                htmlFor="import-entity"
                note="Verificación opcional: debe corresponder a la institución de la cuenta a la que importas."
              >
                <select
                  id="import-entity"
                  value={entityId}
                  onChange={(e) => setEntityId(e.target.value)}
                  className={selectClass}
                >
                  <option value="">Sin especificar (opcional)</option>
                  {accounts.map((a) => (
                    <option key={a.ID} value={a.ID}>
                      {a.Name}
                    </option>
                  ))}
                </select>
              </FormField>
            </section>

            <section className="space-y-2">
              <DrawerSectionLabel>Archivo</DrawerSectionLabel>
              <FormField
                label="Documento PDF"
                htmlFor="import-file"
                note="El extracto se asociará a la cuenta seleccionada y quedará registrado como documento de origen."
              >
                <div className="flex items-center gap-3">
                  <label className="inline-flex h-8 shrink-0 cursor-pointer items-center rounded-tk border border-line bg-raised px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-accent">
                    Seleccionar PDF
                    <input
                      id="import-file"
                      type="file"
                      accept="application/pdf"
                      className="sr-only"
                      onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                    />
                  </label>
                  {file ? (
                    <span className="min-w-0 truncate text-body text-ink-2" title={file.name}>
                      {file.name}
                    </span>
                  ) : (
                    <span className="text-meta text-ink-3">Ningún archivo seleccionado</span>
                  )}
                </div>
              </FormField>
            </section>

            {error && (
              <p className="text-body text-neg" role="alert">
                {error}
              </p>
            )}

            <button
              onClick={submit}
              disabled={!file || !accountId || submitting}
              aria-busy={submitting}
              className="inline-flex h-8 items-center justify-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40 disabled:hover:bg-accent-strong"
            >
              {submitting ? 'Importando…' : 'Importar'}
            </button>
          </>
        )}
      </DrawerBody>
    </DrawerShell>
  );
}
