import { useMemo, useState } from 'react';

import { createAccount } from '../api/accounts';
import { ApiError } from '../api/client';
import { invalidateCollections } from '../lib/dataBus';
import { useInstitutions } from '../hooks/useDashboard';
import { accountTypeLabel } from '../lib/accountType';
import {
  DrawerBody,
  DrawerCloseButton,
  DrawerHeader,
  DrawerSectionLabel,
  DrawerShell,
} from './DrawerShell';
import { FormField, inputClass, selectClass } from './FormField';

/* New account wizard — the financial-source configuration entry point.
   Institution comes from the backend catalog (only what is really
   supported); products come from the institution's supported types.
   Validation data is stored for document matching — NOT bank access. */

interface NewAccountDrawerProps {
  onClose: () => void;
}

const PRODUCT_LABELS: Record<string, string> = {
  BANK_ACCOUNT: 'Cuenta Nu',
  CREDIT_CARD: 'Tarjeta de crédito Nu',
};
export function NewAccountDrawer({ onClose }: NewAccountDrawerProps) {
  const institutions = useInstitutions();

  const [institutionCode, setInstitutionCode] = useState('');
  const [accountType, setAccountType] = useState('');
  const [validationData, setValidationData] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const institution = useMemo(
    () => institutions.data?.find((i) => i.code === institutionCode) ?? null,
    [institutions.data, institutionCode],
  );

  const effectiveInstitution = institution ?? institutions.data?.[0] ?? null;
  const effectiveCode = institutionCode || effectiveInstitution?.code || '';
  const products = effectiveInstitution?.supported_account_types ?? [];
  const effectiveType = accountType || (products.length > 0 ? products[0] : '');

  const accountName = useMemo(() => {
    if (!effectiveInstitution) return '';
    // Name derives from institution, not user input
    return `${effectiveInstitution.name}`;
  }, [effectiveInstitution]);

  const submit = async () => {
    if (!effectiveCode || !effectiveType || !accountName) return;
    setSubmitting(true);
    setError(null);
    try {
      await createAccount({
        id: crypto.randomUUID(),
        user_id: 'local',
        name: accountName,
        account_type: effectiveType as 'BANK_ACCOUNT' | 'CREDIT_CARD',
        currency: 'COP',
        institution: effectiveCode,
        validation_data: validationData.trim() || undefined,
      });
      // Every screen listing accounts refetches through the bus; the drawer
      // closes itself, as the page used to do for it.
      invalidateCollections('accounts');
      onClose();
    } catch (err) {
      const detail =
        err instanceof ApiError && err.response && typeof err.response === 'object'
          ? (err.response as { error?: string }).error
          : undefined;
      setError(detail ?? 'No se pudo crear la cuenta.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <DrawerShell ariaLabel="Nueva cuenta" onClose={onClose}>
      <DrawerHeader>
        <p className="text-title font-semibold text-ink">Nueva cuenta</p>
        <DrawerCloseButton onClose={onClose} />
      </DrawerHeader>

      <DrawerBody>
        {institutions.loading ? (
          <div className="h-40 w-full animate-pulse rounded-tk bg-interactive" aria-hidden="true" />
        ) : institutions.error ? (
          <p className="text-body text-ink-2">
            No se pudo cargar el catálogo de instituciones. Cierra e inténtalo de nuevo.
          </p>
        ) : (
          <>
            <section className="space-y-2">
              <DrawerSectionLabel>Institución</DrawerSectionLabel>
              <select
                value={effectiveCode}
                onChange={(e) => {
                  setInstitutionCode(e.target.value);
                  setAccountType('');
                }}
                aria-label="Institución"
                className={selectClass}
              >
                {(institutions.data ?? []).map((i) => (
                  <option key={i.code} value={i.code}>
                    {i.name}
                  </option>
                ))}
              </select>
            </section>

            <section className="space-y-2">
              <DrawerSectionLabel>Producto</DrawerSectionLabel>
              <div className="space-y-2" role="radiogroup" aria-label="Producto">
                {products.map((type) => {
                  const selected = effectiveType === type;
                  return (
                    <button
                      key={type}
                      type="button"
                      role="radio"
                      aria-checked={selected}
                      onClick={() => setAccountType(type)}
                      className={`flex w-full items-center gap-3 rounded-tk border px-3 py-2.5 text-left transition-colors duration-[120ms] ${
                        selected
                          ? 'border-accent bg-accent-wash'
                          : 'border-line hover:bg-interactive focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent'
                      }`}
                    >
                      <span
                        className={`h-3 w-3 shrink-0 rounded-full border ${
                          selected ? 'border-accent bg-accent' : 'border-ink-3'
                        }`}
                        aria-hidden="true"
                      />
                      <span className="min-w-0">
                        <span className="block text-body font-medium text-ink">
                          {PRODUCT_LABELS[type] ?? type}
                        </span>
                        <span className="block text-meta text-ink-3">{accountTypeLabel(type)}</span>
                      </span>
                    </button>
                  );
                })}
              </div>
            </section>

            <section className="space-y-2">
              <DrawerSectionLabel>Datos de validación</DrawerSectionLabel>
              <FormField
                label="Número de cuenta"
                htmlFor="new-account-validation"
                note="Se utiliza para validar que los documentos que importes correspondan a esta cuenta. No es un mecanismo de acceso bancario."
              >
                <input
                  id="new-account-validation"
                  type="text"
                  value={validationData}
                  onChange={(e) => setValidationData(e.target.value)}
                  placeholder="Opcional"
                  className={inputClass}
                />
              </FormField>
            </section>

            {error && (
              <p className="text-body text-neg" role="alert">
                {error}
              </p>
            )}

            <button
              onClick={submit}
              disabled={!effectiveType || submitting}
              aria-busy={submitting}
              className="inline-flex h-8 items-center justify-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40 disabled:hover:bg-accent-strong"
            >
              {submitting ? 'Creando…' : 'Crear cuenta'}
            </button>
          </>
        )}
      </DrawerBody>
    </DrawerShell>
  );
}
