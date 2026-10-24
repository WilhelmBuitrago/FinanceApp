import { useEffect, useState } from 'react';

import { getReviewEntries, type ReviewItem } from '../api/review';
import { payloadOf } from '../lib/reviewPayload';

/* The dashboard notification bell: its content is the real review queue
   (every trigger, not client-side storage), and a mutation can push one
   more line (a duplicate parked by the add-movement form). */

export function useReviewNotifications(): {
  notifications: string[];
  notify: (text: string) => void;
} {
  const [notifications, setNotifications] = useState<string[]>([]);

  useEffect(() => {
    getReviewEntries()
      .then((items) => {
        const texts = (items || []).map((it: ReviewItem) => {
          const p = payloadOf(it);
          if (it.Kind === 'CATEGORY') return `Movimiento sin categoría: revisá en Revisión`;
          if (it.Kind === 'DUPLICATE') return `Movimiento duplicado: quedó en Revisión`;
          return `Extracción requiere verificación (${p.reason || 'movimiento no normalizado'})`;
        });
        setNotifications(texts);
      })
      .catch(() => {});
  }, []);

  const notify = (text: string) => setNotifications((n) => [...n, text]);

  return { notifications, notify };
}
