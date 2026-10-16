/* Semantic status badge — dot + label, never color alone.
   Tinted background at 10%, text in the semantic color, 11px, radius 6. */

type Tone = 'success' | 'progress' | 'attention' | 'error' | 'neutral';

const TONES: Record<Tone, { dot: string; text: string; bg: string }> = {
  success: { dot: 'bg-pos', text: 'text-pos', bg: 'bg-pos/10' },
  progress: { dot: 'bg-accent', text: 'text-accent', bg: 'bg-accent/10' },
  attention: { dot: 'bg-warn', text: 'text-warn', bg: 'bg-warn/10' },
  error: { dot: 'bg-neg', text: 'text-neg', bg: 'bg-neg/10' },
  neutral: { dot: 'bg-ink-3', text: 'text-ink-2', bg: 'bg-raised' },
};

const STATUS_LABELS: Record<string, string> = {
  IMPORTED: 'Importado',
  CONFIRMED: 'Confirmado',
  SUCCESS: 'Completado',
  ACTIVE: 'Activo',
  VERIFIED: 'Verificado',
  COMPLETED: 'Completado',
  PAID: 'Pagado',
  PROCESSING: 'Procesando',
  EXTRACTED: 'Extraído',
  VALIDATING: 'Validando',
  IMPORTING: 'Importando',
  UPLOADED: 'Subido',
  UPLOAD: 'Subido',
  EXPECTED: 'Esperado',
  DOCUMENTED: 'Documentado',
  REVIEW_REQUIRED: 'Requiere revisión',
  PENDING_REVIEW: 'Pendiente de revisión',
  UNVERIFIED: 'Sin verificar',
  UNKNOWN: 'Desconocido',
  FAILED: 'Fallido',
  REJECTED: 'Rechazado',
  ERROR: 'Error',
  DUPLICATE: 'Duplicado',
  CANCELLED: 'Cancelado',
  REVERSED: 'Revertido',
};

const STATUS_TONES: Record<string, Tone> = {
  IMPORTED: 'success',
  CONFIRMED: 'success',
  SUCCESS: 'success',
  ACTIVE: 'success',
  VERIFIED: 'success',
  COMPLETED: 'success',
  PAID: 'success',
  PROCESSING: 'progress',
  EXTRACTED: 'progress',
  VALIDATING: 'progress',
  IMPORTING: 'progress',
  UPLOADED: 'progress',
  UPLOAD: 'progress',
  EXPECTED: 'progress',
  DOCUMENTED: 'progress',
  REVIEW_REQUIRED: 'attention',
  PENDING_REVIEW: 'attention',
  UNVERIFIED: 'attention',
  UNKNOWN: 'attention',
  FAILED: 'error',
  REJECTED: 'error',
  ERROR: 'error',
  DUPLICATE: 'neutral',
  CANCELLED: 'neutral',
  REVERSED: 'neutral',
};

export const StatusBadge = ({ status }: { status: string }) => {
  const key = (status ?? '').toUpperCase();
  const tone = TONES[STATUS_TONES[key] ?? 'neutral'];
  const label = STATUS_LABELS[key] ?? status;
  return (
    <span
      className={`inline-flex shrink-0 items-center gap-1.5 rounded-tk px-1.5 py-0.5 text-meta font-medium ${tone.bg} ${tone.text}`}
    >
      <span className={`h-1.5 w-1.5 rounded-full ${tone.dot}`} aria-hidden="true" />
      {label}
    </span>
  );
};
