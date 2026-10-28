import type { ResourceHandle } from '../../hooks/useResource';
import { formatDate } from '../../lib/formatDate';
import { formatMoney } from '../../lib/formatMoney';
import type { SourceEntry } from '../../types/source-entry';
import { CopyId } from '../CopyId';
import { DrawerSectionLabel, KV, SectionError, SectionLoading } from '../DrawerShell';

/* ENTRADA DOCUMENTAL — the evidence branch: what the document actually
   said, next to what the normalizer produced. A manually registered
   movement has no entry, and says so instead of failing. */

export function SourceEntryDetail({
  entry,
  manual,
}: {
  entry: ResourceHandle<SourceEntry>;
  manual: boolean;
}) {
  return (
    <section className="space-y-1.5">
      <DrawerSectionLabel>Entrada documental</DrawerSectionLabel>
      {manual ? (
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
  );
}
