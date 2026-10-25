import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getCategories } from '../api/categories';
import { getDocument } from '../api/documents';
import { getFinancialEvent } from '../api/financialEvents';
import { deleteMovement, setMovementCategory } from '../api/movements';
import { getSourceEntry } from '../api/sourceEntries';
import { useResource } from '../hooks/useResource';
import { eventTypeLabel } from '../lib/eventType';
import { formatDate } from '../lib/formatDate';
import { formatMoney, formatSignedMoney, signedAmountClass } from '../lib/formatMoney';
import { CopyId } from './CopyId';
import {
  DrawerBody,
  DrawerCloseButton,
  DrawerHeader,
  DrawerSectionLabel,
  DrawerShell,
} from './DrawerShell';
import { StatusBadge } from './ui/StatusBadge';
import { ConfirmModal } from './ConfirmModal';
import { invalidateCollections } from '../lib/dataBus';
import type { Category } from '../types/category';
import type { Movement } from '../types/movement';

/* Traceability drawer — the product's core navigation:
   Movement -> Financial Event -> Source Entry -> Document.
   Push-style (content shrinks, list stays visible). Esc or ✕ closes.
   Each fetched section resolves/fails/retries independently. */

interface DetailDrawerProps {
  movement: Movement | null;
  accountName: string | null;
  onClose: () => void;
}

function KV({ label, children }: { label: string; children?: React.ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-1">
      <dt className="shrink-0 text-meta text-ink-3">{label}</dt>
      <dd className="min-w-0 break-words text-right text-body text-ink">{children ?? '—'}</dd>
    </div>
  );
}

function SectionLoading() {
  return (
    <div className="space-y-2" aria-hidden="true">
      <div className="h-2.5 w-24 animate-pulse rounded-tk bg-interactive" />
      <div className="h-2.5 w-40 animate-pulse rounded-tk bg-interactive" />
    </div>
  );
}

function SectionError({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex items-center gap-3">
      <p className="text-meta text-ink-2">No se pudo cargar esta sección.</p>
      <button onClick={onRetry} className="text-meta font-medium text-accent hover:underline">
        Reintentar
      </button>
    </div>
  );
}

export function DetailDrawer({ movement, accountName, onClose }: DetailDrawerProps) {
  // The traceability graph behind the movement: all three refetch when a
  // mutation invalidates 'movements' (the category lives on the event, a
  // delete takes the movement down).
  const event = useResource(movement?.FinancialEventID, getFinancialEvent, ['movements']);
  const entry = useResource(movement?.SourceEntryID ?? null, getSourceEntry, ['movements']);
  const document_ = useResource(
    entry.status === 'ready' ? entry.data.DocumentID : null,
    getDocument,
    ['movements'],
  );

  // Category assignment (v0.1.1): the human picks one of their own
  // categories for this movement, or leaves it uncategorized. Leaving it
  // uncategorized is a valid state — a missing category never forces the
  // movement into the review queue. The category lives on the financial
  // event; PUT /movements/{id}/category moves it there.
  const [categories, setCategories] = useState<Category[]>([]);
  const [savingCategory, setSavingCategory] = useState(false);
  const [categoryError, setCategoryError] = useState<string | null>(null);
  // Movement deletion: confirmed through ConfirmModal (the destructive
  // action the repo already has), never through a native confirm().
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getCategories()
      .then((data) => {
        if (!cancelled) setCategories(data);
      })
      .catch(() => {
        if (!cancelled) setCategories([]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Same filtering the create form uses: only categories of the movement's
  // event type. Until the event resolves, every category stays available.
  const eventType = event.status === 'ready' ? event.data.EventType : '';
  const categoryOptions = eventType
    ? categories.filter((c) => (c.CategoryType ?? '').toUpperCase() === eventType)
    : categories;

  const assignCategory = async (categoryId: string | null) => {
    if (!movement) return;
    setSavingCategory(true);
    setCategoryError(null);
    try {
      await setMovementCategory(movement.ID, categoryId);
      // The event behind the movement changed: every movement view — and
      // this drawer's own traceability sections — refetches through the bus.
      invalidateCollections('movements');
    } catch {
      setCategoryError('No se pudo guardar la categoría.');
    } finally {
      setSavingCategory(false);
    }
  };

  if (!movement) return null;

  return (
    <DrawerShell ariaLabel="Detalle del movimiento" onClose={onClose}>
      <DrawerHeader>
        <p
          className={`text-amount font-semibold tabular-nums ${signedAmountClass(movement.AmountMinor)}`}
        >
          {formatSignedMoney(movement.AmountMinor)}
        </p>
        <DrawerCloseButton onClose={onClose} />
      </DrawerHeader>

      <DrawerBody>
        {/* MOVIMIENTO — local data, always present */}
        <section className="space-y-1.5">
          <DrawerSectionLabel>Movimiento</DrawerSectionLabel>
          <dl>
            <KV label="Fecha">{formatDate(movement.MovementDate)}</KV>
            <KV label="Cuenta">
              <Link
                to={`/accounts/${movement.AccountID}`}
                className="text-accent transition-colors hover:underline"
              >
                {accountName ?? movement.AccountID}
              </Link>
            </KV>
            <KV label="Categoría">
              {categoryError ? (
                <span className="text-neg">{categoryError}</span>
              ) : (
                <select
                  value={movement.categoryId ?? ''}
                  onChange={(e) => assignCategory(e.target.value || null)}
                  disabled={savingCategory}
                  aria-label="Categoría del movimiento"
                  className="rounded-tk border border-line bg-bg px-2 py-1 text-body text-ink disabled:opacity-40"
                >
                  <option value="">Sin categoría</option>
                  {categoryOptions.map((c) => (
                    <option key={c.ID} value={c.ID}>
                      {c.Name}
                    </option>
                  ))}
                </select>
              )}
            </KV>
            <KV label="Estado">
              <StatusBadge status={movement.Status} />
            </KV>
            <KV label="ID">
              <CopyId id={movement.ID} />
            </KV>
          </dl>
          <div className="pt-2">
            <button
              onClick={() => setConfirmingDelete(true)}
              className="rounded-tk border border-line bg-bg px-3 py-1.5 text-sm font-medium text-neg hover:bg-red-50"
            >
              Eliminar movimiento
            </button>
            {deleteError && (
              <p className="mt-2 text-meta text-neg" role="alert">
                {deleteError}
              </p>
            )}
            {confirmingDelete && (
              <div className="mt-3">
                <ConfirmModal
                  title="Eliminar movimiento"
                  message="¿Eliminar este movimiento? Esta acción no se puede deshacer."
                  confirmLabel="Eliminar"
                  busy={deleting}
                  onCancel={() => setConfirmingDelete(false)}
                  onConfirm={async () => {
                    setDeleting(true);
                    try {
                      await deleteMovement(movement.ID);
                      setConfirmingDelete(false);
                      onClose();
                      // The movement is gone: lists and aggregates refetch
                      // through the bus, no document reload.
                      invalidateCollections('movements', 'dashboard');
                    } catch {
                      setDeleteError('No se pudo eliminar el movimiento');
                      setConfirmingDelete(false);
                    } finally {
                      setDeleting(false);
                    }
                  }}
                />
              </div>
            )}
          </div>
        </section>

        {/* EVENTO FINANCIERO */}
        <section className="space-y-1.5">
          <DrawerSectionLabel>Evento financiero</DrawerSectionLabel>
          {event.status === 'loading' && <SectionLoading />}
          {event.status === 'error' && <SectionError onRetry={event.retry} />}
          {event.status === 'ready' && (
            <dl>
              <KV label="Tipo">{eventTypeLabel(event.data.EventType)}</KV>
              <KV label="Fecha">{formatDate(event.data.EventDate)}</KV>
              <KV label="Verificación">
                <StatusBadge status={event.data.VerificationStatus} />
              </KV>
              <KV label="Descripción">{event.data.Description}</KV>
              <KV label="Confianza">
                {event.data.Confidence != null
                  ? `${Math.round(event.data.Confidence * 100)}%`
                  : '—'}
              </KV>
              <KV label="ID">
                <CopyId id={event.data.ID} />
              </KV>
            </dl>
          )}
        </section>

        {/* ENTRADA DOCUMENTAL — the evidence branch */}
        <section className="space-y-1.5">
          <DrawerSectionLabel>Entrada documental</DrawerSectionLabel>
          {movement.SourceEntryID == null ? (
            <p className="text-body text-ink-2">Sin entrada documental — registrado manualmente.</p>
          ) : entry.status === 'loading' ? (
            <SectionLoading />
          ) : entry.status === 'error' ? (
            <SectionError onRetry={entry.retry} />
          ) : entry.status === 'ready' ? (
            <div className="space-y-3">
              <div className="rounded-tk border border-line bg-bg p-3">
                <p className="font-mono text-meta text-ink-3">Evidencia del documento</p>
                <p className="mt-1.5 break-words font-mono text-body text-ink-2">
                  {entry.data.RawDescription ?? '—'}
                </p>
                {entry.data.RawDate && (
                  <p className="mt-1 font-mono text-meta text-ink-3">
                    Fecha original: {entry.data.RawDate}
                  </p>
                )}
                {entry.data.RawAmountMinor != null && (
                  <p className="font-mono text-meta text-ink-3">
                    Monto original: {entry.data.RawAmountMinor}
                  </p>
                )}
              </div>
              <dl>
                <KV label="Descripción normalizada">{entry.data.NormalizedDescription}</KV>
                <KV label="Monto normalizado">
                  {entry.data.NormalizedAmountMinor != null
                    ? formatMoney(entry.data.NormalizedAmountMinor)
                    : '—'}
                </KV>
                <KV label="Fecha normalizada">{formatDate(entry.data.NormalizedDate)}</KV>
                <KV label="Página / fila">
                  {entry.data.PageNumber != null
                    ? `pág. ${entry.data.PageNumber}${
                        entry.data.RowNumber != null ? ` · fila ${entry.data.RowNumber}` : ''
                      }`
                    : '—'}
                </KV>
                <KV label="ID">
                  <CopyId id={entry.data.ID} />
                </KV>
              </dl>
            </div>
          ) : null}
        </section>

        {/* DOCUMENTO — appears progressively once the entry resolves */}
        {entry.status === 'ready' && (
          <section className="space-y-1.5">
            <DrawerSectionLabel>Documento</DrawerSectionLabel>
            {document_.status === 'loading' && <SectionLoading />}
            {document_.status === 'error' && <SectionError onRetry={document_.retry} />}
            {document_.status === 'ready' && (
              <div className="space-y-2">
                <dl>
                  <KV label="Nombre">{document_.data.DocumentName ?? 'Documento'}</KV>
                  <KV label="Tipo">{document_.data.DocumentType}</KV>
                  <KV label="Institución">{document_.data.Institution}</KV>
                  <KV label="Período">
                    {document_.data.PeriodStart && document_.data.PeriodEnd
                      ? `${formatDate(document_.data.PeriodStart)} — ${formatDate(document_.data.PeriodEnd)}`
                      : '—'}
                  </KV>
                  <KV label="Estado">
                    <StatusBadge status={document_.data.Status} />
                  </KV>
                  <KV label="ID">
                    <CopyId id={document_.data.ID} />
                  </KV>
                </dl>
                <Link
                  to="/documents"
                  className="inline-flex text-meta font-medium text-accent hover:underline"
                >
                  Ver en Documentos
                </Link>
              </div>
            )}
          </section>
        )}
      </DrawerBody>
    </DrawerShell>
  );
}
