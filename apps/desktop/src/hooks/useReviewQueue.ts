import { useEffect, useState } from 'react';

import { getAccounts } from '../api/accounts';
import { getCategories } from '../api/categories';
import { ApiError } from '../api/client';
import {
  discardReviewEntry,
  getReviewEntries,
  resolveReviewCategory,
  resolveReviewMovement,
  resolveReviewPeriod,
  type ReviewItem,
} from '../api/review';
import { invalidateCollections } from '../lib/dataBus';
import { payloadOf } from '../lib/reviewPayload';
import type { Account } from '../types/account';
import type { CategoryOption } from '../types/category';

/* All review-page state: the queue, the data its cards need (categories
   to assign, accounts to correct against), the per-item form, and the
   four resolutions. The resolutions keep the exact UX they had —alert
   copy, in-place removal; the bus is what refreshes the queue, the
   sidebar badge and the dashboard now. */

export function useReviewQueue() {
  const [items, setItems] = useState<ReviewItem[]>([]);
  // The cards read these already normalized: {id, name, type} and the
  // typed account contract — no anonymous objects downstream.
  const [categories, setCategories] = useState<CategoryOption[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  // Per-item form: every value is a string coming from the DOM.
  const [form, setForm] = useState<Record<string, Record<string, string>>>({});
  const [selectedCat, setSelectedCat] = useState<Record<string, string>>({});
  // Transient outcome of a resolution, shown inline (the old alert()).
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    getReviewEntries()
      .then((d) => setItems(d || []))
      .catch(() => setItems([]));
    getCategories()
      .then((d) =>
        setCategories(
          (d || []).map((c) => ({
            id: c.ID,
            name: c.Name,
            type: (c.CategoryType ?? '').toUpperCase(),
          })),
        ),
      )
      .catch(() => {});
    getAccounts()
      .then((d) => setAccounts(d || []))
      .catch(() => {});
  }, []);

  const setField = (id: string, patch: Record<string, string>) =>
    setForm((f) => ({ ...f, [id]: { ...(f[id] || {}), ...patch } }));

  // The queue closed this item; every screen that reads the queue or the
  // movements it produced refetches through the bus. The badge does too —
  // no custom event needed anymore.
  const dropItem = (id: string) => {
    setItems((its) => its.filter((i) => i.ID !== id));
    invalidateCollections('review', 'movements', 'dashboard', 'categories');
  };

  const dismissNotice = () => setNotice(null);

  const resolveMovement = async (item: ReviewItem) => {
    const p = payloadOf(item);
    const f = form[item.ID] || {};
    const description = f.description ?? p.description ?? p.description_original ?? '';
    const date = f.date ?? p.date ?? p.date_original ?? '';
    const amountMinor =
      f.amount !== undefined && f.amount !== ''
        ? Math.round(Number(f.amount) * 100) * (f.type === 'INCOME' ? 1 : -1)
        : (p.amount_minor ?? 0);
    const type = f.type ?? p.type ?? 'EXPENSE';
    const accountId = f.account ?? p.account_id ?? '';
    const categoryId = selectedCat[item.ID] || null;
    if (!date || !accountId || !type) {
      setNotice('Completá fecha, monto y cuenta para confirmar');
      return;
    }
    try {
      await resolveReviewMovement(item.ID, {
        description,
        amount_minor: amountMinor,
        date,
        type,
        account_id: accountId,
        category_id: categoryId,
      });
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setNotice('Este ítem ya fue resuelto');
        dropItem(item.ID);
      } else {
        setNotice('Error al confirmar la corrección');
      }
    }
  };

  // Document-level resolution: the human supplies the statement month and
  // year (ABR 2026); the backend computes the full period and derives the
  // parked movement dates of that document against it.
  const resolvePeriod = async (item: ReviewItem) => {
    const f = form[item.ID] || {};
    const month = Number(f.month);
    const year = Number(f.year);
    if (!month || !year) {
      setNotice('Elegí el mes y el año del estado de cuenta');
      return;
    }
    try {
      await resolveReviewPeriod(item.ID, month, year);
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setNotice('Este ítem ya fue resuelto');
        dropItem(item.ID);
      } else {
        setNotice('Error al confirmar el período');
      }
    }
  };

  const resolveCategory = async (item: ReviewItem, categoryId: string | null) => {
    try {
      await resolveReviewCategory(item.ID, categoryId);
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        dropItem(item.ID);
      } else {
        setNotice('Error al resolver la categoría');
      }
    }
  };

  const discard = async (item: ReviewItem) => {
    try {
      await discardReviewEntry(item.ID);
      dropItem(item.ID);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        dropItem(item.ID);
      }
    }
  };

  return {
    items,
    categories,
    accounts,
    form,
    setField,
    selectedCat,
    setSelectedCat,
    notice,
    dismissNotice,
    resolveMovement,
    resolvePeriod,
    resolveCategory,
    discard,
  };
}
