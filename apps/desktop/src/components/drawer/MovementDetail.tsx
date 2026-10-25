import { Link } from 'react-router-dom';

import { formatDate } from '../../lib/formatDate';
import type { Category } from '../../types/category';
import type { Movement } from '../../types/movement';
import { ConfirmModal } from '../ConfirmModal';
import { CopyId } from '../CopyId';
import { DrawerSectionLabel, KV } from '../DrawerShell';
import { StatusBadge } from '../ui/StatusBadge';

/* MOVIMIENTO — local data, always present. Category assignment and
   deletion (their machine lives in useMovementActions) render inside
   this section. */

interface MovementDetailProps {
  movement: Movement;
  accountName: string | null;
  categories: Category[];
  savingCategory: boolean;
  categoryError: string | null;
  onAssignCategory: (categoryId: string | null) => void;
  confirmingDelete: boolean;
  deleting: boolean;
  deleteError: string | null;
  onRequestDelete: () => void;
  onCancelDelete: () => void;
  onConfirmDelete: () => void;
}

export function MovementDetail({
  movement,
  accountName,
  categories,
  savingCategory,
  categoryError,
  onAssignCategory,
  confirmingDelete,
  deleting,
  deleteError,
  onRequestDelete,
  onCancelDelete,
  onConfirmDelete,
}: MovementDetailProps) {
  return (
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
              onChange={(e) => onAssignCategory(e.target.value || null)}
              disabled={savingCategory}
              aria-label="Categoría del movimiento"
              className="rounded-tk border border-line bg-bg px-2 py-1 text-body text-ink disabled:opacity-40"
            >
              <option value="">Sin categoría</option>
              {categories.map((c) => (
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
          onClick={onRequestDelete}
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
              onCancel={onCancelDelete}
              onConfirm={onConfirmDelete}
            />
          </div>
        )}
      </div>
    </section>
  );
}
