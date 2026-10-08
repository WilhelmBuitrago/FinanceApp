/* Date formatting — es-CO default (project UI language: Spanish).
   Compact form "31 ene 2026" assembled from formatToParts (Intl's default
   Spanish rendering inserts "de" twice). Date-only values ("2026-01-31")
   parse as UTC midnight; formatting them in local time shifts the day in
   negative-offset timezones, so date-only strings are formatted in UTC
   to display the written date exactly. */

const DATE_ONLY_RE = /^\d{4}-\d{2}-\d{2}$/;

function formatCompact(
  date: Date,
  locale: string,
  options: Intl.DateTimeFormatOptions,
): string {
  const parts = new Intl.DateTimeFormat(locale, options).formatToParts(date);
  const get = (type: string) => parts.find((part) => part.type === type)?.value ?? '';
  return `${get('day')} ${get('month')} ${get('year')}`.trim();
}

export function formatDate(iso: string | null | undefined, locale = 'es-CO'): string {
  if (!iso) return '—';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  const dateOnly = DATE_ONLY_RE.test(iso.trim());
  return formatCompact(
    date,
    locale,
    dateOnly
      ? { year: 'numeric', month: 'short', day: '2-digit', timeZone: 'UTC' }
      : { year: 'numeric', month: 'short', day: '2-digit' },
  );
}

export function formatDateTime(iso: string | null | undefined, locale = 'es-CO'): string {
  if (!iso) return '—';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  const parts = new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).formatToParts(date);
  const get = (type: string) => parts.find((part) => part.type === type)?.value ?? '';
  return `${get('day')} ${get('month')} ${get('year')}, ${get('hour')}:${get('minute')}`;
}
