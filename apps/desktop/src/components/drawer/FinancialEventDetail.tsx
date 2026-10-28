import type { ResourceHandle } from '../../hooks/useResource';
import { eventTypeLabel } from '../../lib/eventType';
import { formatDate } from '../../lib/formatDate';
import type { FinancialEvent } from '../../types/financial-event';
import { CopyId } from '../CopyId';
import { DrawerSectionLabel, KV, SectionError, SectionLoading } from '../DrawerShell';
import { StatusBadge } from '../ui/StatusBadge';

/* EVENTO FINANCIERO — the ledger entry behind the movement. Resolves,
   fails or retries on its own. */

export function FinancialEventDetail({ event }: { event: ResourceHandle<FinancialEvent> }) {
  return (
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
            {event.data.Confidence != null ? `${Math.round(event.data.Confidence * 100)}%` : '—'}
          </KV>
          <KV label="ID">
            <CopyId id={event.data.ID} />
          </KV>
        </dl>
      )}
    </section>
  );
}
