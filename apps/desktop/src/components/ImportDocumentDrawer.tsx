import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { uploadDocument } from '../api/documents';
import { getJob } from '../api/jobs';
import { getReviewEntries, type ReviewItem } from '../api/review';
import { REVIEW_QUEUE_CHANGED_EVENT } from '../lib/reviewEvents';
import { ApiError } from '../api/client';
import {
  DrawerBody,
  DrawerCloseButton,
  DrawerHeader,
  DrawerSectionLabel,
  DrawerShell,
} from './DrawerShell';
import { FormField, selectClass } from './FormField';
import { CopyId } from './CopyId';
import type { Account } from '../types/account';

/* Import document drawer — an import ALWAYS belongs to a configured
   account (S12: no orphan documents). With one account it is shown fixed;
   with several, it is requested explicitly. No account → the configuration
   path, never an empty selector. */

interface ImportDocumentDrawerProps {
  accounts: Account[];
  onClose: () => void;
  onImported: () => void;
}

export function ImportDocumentDrawer({ accounts, onClose, onImported }: ImportDocumentDrawerProps) {
  const [accountId, setAccountId] = useState(accounts[0]?.ID ?? '');
  const [file, setFile] = useState<File | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadResult, setUploadResult] = useState<{
    document_id: string;
    job_id: string;
    processing_id?: string;
  } | null>(null);
  const [jobState, setJobState] = useState<'RUNNING' | 'SUCCESS' | 'FAILED' | null>(null);
  const [reviewCount, setReviewCount] = useState<number | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  // After an upload, poll the job until it reaches a terminal state and then
  // report what landed in the human-review queue for this document: the user
  // learns here, not by browsing to Revisión.
  useEffect(() => {
    if (!uploadResult) return;
    let cancelled = false;
    let attempts = 0;
    setJobState('RUNNING');
    setReviewCount(null);
    const tick = async () => {
      try {
        const job = await getJob(uploadResult.job_id);
        if (cancelled) return;
        if (job.Status === 'SUCCESS' || job.Status === 'FAILED') {
          setJobState(job.Status);
          const items = await getReviewEntries().catch(() => [] as ReviewItem[]);
          if (cancelled) return;
          const mine = (items || []).filter((it: ReviewItem) => {
            try {
              return JSON.parse(it.PayloadJson || '{}').document_id === uploadResult.document_id;
            } catch {
              return false;
            }
          });
          setReviewCount(mine.length);
          window.dispatchEvent(new Event(REVIEW_QUEUE_CHANGED_EVENT));
          onImported();
          return;
        }
      } catch {
        /* job still processing or transient error: retry */
      }
      if (cancelled) return;
      attempts += 1;
      if (attempts >= 20) {
        setJobState('FAILED');
        return;
      }
      setTimeout(tick, 1500);
    };
    tick();
    return () => {
      cancelled = true;
    };
  }, [uploadResult]);

  const submit = async () => {
    if (!accountId || !file) return;
    setSubmitting(true);
    setError(null);
    // Entity is derived from the selected account's institution
    const account = accounts.find((a) => a.ID === accountId);
    const entity = account?.Institution ?? undefined;
    try {
      const result = await uploadDocument(accountId, file, entity);
      setUploadResult(result);
    } catch (err) {
      let detail: string | undefined;
      if (err instanceof ApiError) {
        const r = err.response;
        if (typeof r === 'object' && r && 'error' in (r as any)) {
          detail = (r as any).error;
        } else if (typeof r === 'string') {
          try {
            const parsed = JSON.parse(r);
            if (parsed && typeof parsed === 'object' && 'error' in parsed) {
              detail = parsed.error;
            } else {
              detail = r;
            }
          } catch {
            detail = r;
          }
        }
      }
      setError(detail ?? 'No se pudo importar el documento.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleFileChange = (f: File | null) => {
    setFile(f);
    setError(null);
    setUploadResult(null);
  };

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const dropped = e.dataTransfer.files?.[0];
    if (dropped && dropped.type === 'application/pdf') {
      handleFileChange(dropped);
    } else if (dropped) {
      setError('Solo se permiten archivos PDF.');
    }
  };

  const onDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const onDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  return (
    <DrawerShell ariaLabel="Importar documento" onClose={onClose}>
      <DrawerHeader>
        <p className="text-title font-semibold text-ink">Importar documento</p>
        <DrawerCloseButton onClose={onClose} />
      </DrawerHeader>

      <DrawerBody>
        {accounts.length === 0 ? (
          <div className="space-y-4">
            <p className="text-body text-ink-2">Primero necesitas configurar una cuenta.</p>
            <Link
              to="/accounts"
              onClick={onClose}
              className="inline-flex h-8 items-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
            >
              Crear cuenta
            </Link>
          </div>
        ) : (
          <>
            <section className="space-y-2">
              <DrawerSectionLabel>Cuenta</DrawerSectionLabel>
              {accounts.length === 1 ? (
                <p className="rounded-tk border border-line bg-bg px-3 py-2 text-body text-ink">
                  {accounts[0].Name}
                  <span className="ml-2 text-meta text-ink-3">
                    {accounts[0].Institution ?? ''} · {accounts[0].Currency}
                  </span>
                </p>
              ) : (
                <select
                  value={accountId}
                  onChange={(e) => setAccountId(e.target.value)}
                  aria-label="Cuenta"
                  className={selectClass}
                >
                  {accounts.map((a) => (
                    <option key={a.ID} value={a.ID}>
                      {a.Name}
                    </option>
                  ))}
                </select>
              )}
            </section>

            <section className="space-y-2">
              <DrawerSectionLabel>Archivo</DrawerSectionLabel>
              <FormField
                label="Documento PDF"
                htmlFor="import-file"
                note="El extracto se asociará a la cuenta seleccionada y quedará registrado como documento de origen."
              >
                <div
                  onDragOver={onDragOver}
                  onDragLeave={onDragLeave}
                  onDrop={onDrop}
                  className={`group relative rounded-tk border ${isDragging ? 'border-accent bg-interactive' : file ? 'border-line bg-bg' : 'border-dashed border-line bg-bg'} px-4 py-3 transition-colors`}
                  aria-dropeffect="execute"
                >
                  <div className="flex items-start gap-3">
                    <div
                      className={`mt-0.5 rounded-md bg-raised p-2 text-ink-3 group-[.is-dragging]:text-accent`}
                    >
                      <svg
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        aria-hidden="true"
                      >
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                        <polyline points="14 2 14 8 20 8" />
                      </svg>
                    </div>
                    <div className="min-w-0 flex-1">
                      {file ? (
                        <>
                          <p className="text-body text-ink font-medium truncate" title={file.name}>
                            {file.name}
                          </p>
                          <p className="text-meta text-ink-3">{(file.size / 1024).toFixed(0)} KB</p>
                        </>
                      ) : (
                        <>
                          <p className="text-body text-ink">Arrastra un PDF aquí</p>
                          <p className="text-meta text-ink-3">o selecciona con el botón</p>
                        </>
                      )}
                    </div>
                    <label className="inline-flex h-8 shrink-0 cursor-pointer items-center rounded-tk border border-line bg-raised px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-accent">
                      Seleccionar PDF
                      <input
                        id="import-file"
                        type="file"
                        accept="application/pdf"
                        className="sr-only"
                        onChange={(e) => handleFileChange(e.target.files?.[0] ?? null)}
                      />
                    </label>
                  </div>
                </div>
              </FormField>
            </section>

            <div aria-live="polite" aria-atomic="true" className="space-y-3">
              {error && (
                <div
                  className="rounded-tk border border-neg/30 bg-neg/10 px-3 py-2 text-body text-neg"
                  role="alert"
                >
                  <p className="font-medium">No se pudo importar</p>
                  <p className="text-meta">{error}</p>
                </div>
              )}
              {uploadResult && (
                <div className="rounded-tk border border-line bg-raised px-4 py-3 space-y-3">
                  <div className="flex items-start gap-3">
                    <span className="mt-0.5 text-pos">
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        aria-hidden="true"
                      >
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    </span>
                    <div>
                      <p className="text-body font-medium text-ink">Documento recibido</p>
                      <p className="text-meta text-ink-3">
                        El extracto se ha enviado a la cuenta seleccionada. Puedes seguir su estado
                        en Documentos.
                      </p>
                    </div>
                  </div>
                  <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-meta">
                    <dt className="text-ink-3">Documento</dt>
                    <dd className="text-ink">
                      <CopyId id={uploadResult.document_id} />
                    </dd>
                    <dt className="text-ink-3">Trabajo</dt>
                    <dd className="text-ink">
                      <CopyId id={uploadResult.job_id} />
                    </dd>
                    {uploadResult.processing_id && (
                      <>
                        <dt className="text-ink-3">Procesamiento</dt>
                        <dd className="text-ink">
                          <CopyId id={uploadResult.processing_id} />
                        </dd>
                      </>
                    )}
                  </dl>
                  {/* Import outcome: what the pipeline parked for a human. */}
                  {jobState === 'RUNNING' && (
                    <p className="text-meta text-ink-3">Procesando el extracto…</p>
                  )}
                  {jobState === 'FAILED' && (
                    <p className="text-body text-neg">
                      El procesamiento falló.Revisá el documento en Documentos.
                    </p>
                  )}
                  {jobState === 'SUCCESS' &&
                    reviewCount !== null &&
                    (reviewCount > 0 ? (
                      <p className="text-body text-ink">
                        {reviewCount} movimiento(s) requieren verificación.{' '}
                        <a href="/review" className="text-accent-strong underline">
                          Ir a Revisión
                        </a>
                      </p>
                    ) : (
                      <p className="text-body text-ink-2">
                        Extracto importado sin ítems pendientes.
                      </p>
                    ))}
                </div>
              )}
            </div>

            <div className="flex items-center gap-3">
              {!uploadResult ? (
                <button
                  onClick={submit}
                  disabled={!file || !accountId || submitting}
                  aria-busy={submitting}
                  className="inline-flex h-8 items-center justify-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40 disabled:hover:bg-accent-strong"
                >
                  {submitting ? 'Importando…' : 'Importar'}
                </button>
              ) : (
                <button
                  onClick={() => {
                    onImported();
                    onClose();
                  }}
                  className="inline-flex h-8 items-center justify-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
                >
                  Ver documentos
                </button>
              )}
              {(error || uploadResult) && (
                <button
                  onClick={() => {
                    setError(null);
                    setUploadResult(null);
                    setFile(null);
                  }}
                  className="inline-flex h-8 items-center justify-center rounded-tk border border-line px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
                >
                  {uploadResult ? 'Importar otro' : 'Reiniciar'}
                </button>
              )}
            </div>
          </>
        )}
      </DrawerBody>
    </DrawerShell>
  );
}
