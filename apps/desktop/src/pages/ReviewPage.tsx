import { Link } from 'react-router-dom';

import { Notice } from '../components/ui/Notice';
import { ReviewCategoryCard } from '../components/review/ReviewCategoryCard';
import { ReviewMovementCard } from '../components/review/ReviewMovementCard';
import { ReviewPeriodCard } from '../components/review/ReviewPeriodCard';
import { useReviewQueue } from '../hooks/useReviewQueue';
import { payloadOf } from '../lib/reviewPayload';

/* Human review. One queue, four triggers: DUPLICATE (a movement that
   collided), CATEGORY (movements left without a category), EXTRACTION (a
   movement the processor could not normalize) and DOCUMENT_PERIOD (a
   statement period nobody could read). The queue lives in the backend
   (review_items), so a resolution survives restarts and every page reads
   the same source.

   This page only composes: queue and resolutions live in useReviewQueue,
   and each Kind renders its own card under components/review/. */

export default function ReviewPage() {
  const {
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
  } = useReviewQueue();

  if (items.length === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Revisión</h1>
        <div className="rounded-[12px] border border-dashed border-line bg-surface/50 p-12 text-center">
          <p className="text-body text-ink-2">No hay ítems pendientes de revisión.</p>
          <p className="mt-1 text-meta text-ink-3">
            Cuando un documento, un duplicado o una categoría requiera atención aparecerá aquí.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Revisión</h1>
      <Notice message={notice} onDismiss={dismissNotice} />
      {items.map((item) => {
        const p = payloadOf(item);
        const f = form[item.ID] || {};

        if (item.Kind === 'CATEGORY') {
          return (
            <ReviewCategoryCard
              key={item.ID}
              payload={p}
              categories={categories}
              selectedId={selectedCat[item.ID] || ''}
              onSelectChange={(value) => setSelectedCat((s) => ({ ...s, [item.ID]: value }))}
              onAssign={() => resolveCategory(item, selectedCat[item.ID] || null)}
              onDiscard={() => discard(item)}
            />
          );
        }

        if (item.Kind === 'DOCUMENT_PERIOD') {
          return (
            <ReviewPeriodCard
              key={item.ID}
              item={item}
              payload={p}
              monthValue={f.month ?? (p.suggested_month ? String(p.suggested_month) : '')}
              yearValue={f.year ?? (p.suggested_year ? String(p.suggested_year) : '')}
              onMonth={(value) => setField(item.ID, { month: value })}
              onYear={(value) => setField(item.ID, { year: value })}
              onConfirm={() => resolvePeriod(item)}
              onDiscard={() => discard(item)}
            />
          );
        }

        return (
          <ReviewMovementCard
            key={item.ID}
            item={item}
            payload={p}
            form={f}
            accounts={accounts}
            onChange={(patch) => setField(item.ID, patch)}
            onSubmit={() => resolveMovement(item)}
            onDiscard={() => discard(item)}
          />
        );
      })}
      <Link to="/" className="px-3 py-1 rounded border inline-block">
        Volver
      </Link>
    </div>
  );
}
