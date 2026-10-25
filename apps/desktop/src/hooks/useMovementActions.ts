import { useCallback, useEffect, useState } from 'react';

import { getCategories } from '../api/categories';
import { deleteMovement, setMovementCategory } from '../api/movements';
import { invalidateCollections } from '../lib/dataBus';
import type { Category } from '../types/category';
import type { Movement } from '../types/movement';

/* Category assignment and deletion for the traceability drawer.

   Category (v0.1.1): the human picks one of their own categories, or
   leaves the movement uncategorized — a missing category never forces
   the movement into the review queue. The category lives on the
   financial event; PUT /movements/{id}/category moves it there.

   Deletion is destructive: confirmed through ConfirmModal, never a
   native confirm(). Both mutations refresh through the bus. */

export interface MovementActions {
  categories: Category[];
  savingCategory: boolean;
  categoryError: string | null;
  assignCategory: (categoryId: string | null) => void;
  confirmingDelete: boolean;
  deleting: boolean;
  deleteError: string | null;
  requestDelete: () => void;
  cancelDelete: () => void;
  confirmDelete: () => void;
}

export function useMovementActions(
  movement: Movement | null,
  eventType: string,
  onClose: () => void,
): MovementActions {
  const [categories, setCategories] = useState<Category[]>([]);
  const [savingCategory, setSavingCategory] = useState(false);
  const [categoryError, setCategoryError] = useState<string | null>(null);
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
  const available = eventType
    ? categories.filter((c) => (c.CategoryType ?? '').toUpperCase() === eventType)
    : categories;

  const assignCategory = useCallback(
    async (categoryId: string | null) => {
      if (!movement) return;
      setSavingCategory(true);
      setCategoryError(null);
      try {
        await setMovementCategory(movement.ID, categoryId);
        // The event behind the movement changed: every movement view —
        // and this drawer's own traceability sections — refetches.
        invalidateCollections('movements');
      } catch {
        setCategoryError('No se pudo guardar la categoría.');
      } finally {
        setSavingCategory(false);
      }
    },
    [movement],
  );

  const confirmDelete = useCallback(async () => {
    if (!movement) return;
    setDeleting(true);
    try {
      await deleteMovement(movement.ID);
      setConfirmingDelete(false);
      onClose();
      // The movement is gone: lists and aggregates refetch through the
      // bus, no document reload.
      invalidateCollections('movements', 'dashboard');
    } catch {
      setDeleteError('No se pudo eliminar el movimiento');
      setConfirmingDelete(false);
    } finally {
      setDeleting(false);
    }
  }, [movement, onClose]);

  return {
    categories: available,
    savingCategory,
    categoryError,
    assignCategory,
    confirmingDelete,
    deleting,
    deleteError,
    requestDelete: () => setConfirmingDelete(true),
    cancelDelete: () => setConfirmingDelete(false),
    confirmDelete,
  };
}
