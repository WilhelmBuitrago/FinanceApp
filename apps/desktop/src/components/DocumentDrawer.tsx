import React from 'react';
import { Link } from 'react-router-dom';

import { CopyId } from './CopyId';
import {
  DrawerBody,
  DrawerCloseButton,
  DrawerHeader,
  DrawerSectionLabel,
  DrawerShell,
} from './DrawerShell';
import { StatusBadge } from './ui/StatusBadge';
import { formatDate } from '../lib/formatDate';
import { formatPeriod } from '../lib/formatPeriod';
import type { Document } from '../types/document';

/* Document drawer — provenance: which account owns it, its financial
   period (≠ upload date), its state, and its origin file info. */

interface DocumentDrawerProps {
  document: Document | null;
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

function fileNameFrom(sourceUri?: string | null): string | null {
  if (!sourceUri) return null;
  const parts = sourceUri.split(/[\\/]/);
  return parts[parts.length - 1] || null;
}

export function DocumentDrawer({ document: doc, accountName, onClose }: DocumentDrawerProps) {
  if (!doc) return null;

  return (
    <DrawerShell ariaLabel="Detalle del documento" onClose={onClose}>
      <DrawerHeader>
        <p className="min-w-0 truncate text-title font-semibold text-ink">
          {doc.DocumentName ?? 'Documento'}
        </p>
        <DrawerCloseButton onClose={onClose} />
      </DrawerHeader>

      <DrawerBody>
        <section className="space-y-1.5">
          <DrawerSectionLabel>Documento</DrawerSectionLabel>
          <dl>
            <KV label="Cuenta">
              {doc.AccountID ? (
                <Link
                  to={`/accounts/${doc.AccountID}`}
                  className="text-accent transition-colors hover:underline"
                >
                  {accountName ?? doc.AccountID}
                </Link>
              ) : (
                '—'
              )}
            </KV>
            <KV label="Período">{formatPeriod(doc.PeriodStart, doc.PeriodEnd)}</KV>
            <KV label="Fecha de subida">{formatDate(doc.CreatedAt)}</KV>
            <KV label="Estado">
              <StatusBadge status={doc.Status} />
            </KV>
          </dl>
        </section>

        <section className="space-y-1.5">
          <DrawerSectionLabel>Información de origen</DrawerSectionLabel>
          <dl>
            <KV label="Archivo">
              {fileNameFrom(doc.SourceUri) ? (
                <span className="break-all font-mono text-meta text-ink-2">
                  {fileNameFrom(doc.SourceUri)}
                </span>
              ) : (
                '—'
              )}
            </KV>
            <KV label="Hash">
              {doc.FileHash ? (
                <span className="font-mono text-meta text-ink-3">
                  {doc.FileHash.slice(0, 13)}…
                </span>
              ) : (
                '—'
              )}
            </KV>
            <KV label="ID">
              <CopyId id={doc.ID} />
            </KV>
          </dl>
        </section>
      </DrawerBody>
    </DrawerShell>
  );
}
