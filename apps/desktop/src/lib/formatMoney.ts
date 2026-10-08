/* Money formatting for FinanceApp.
   CONTRACT (verified against apps/processor/normalization/nu_normalizer.py):
   normalize_amount() stores INTEGER PESOS in amount_minor
   ("+$700.000,00" -> 700000), NOT hundredths. Values are formatted as-is,
   with no /100 division. Currency: COP, locale es-CO, compact "$3.641"
   (Intl's es-CO currency style inserts a non-breaking space, which breaks
   column alignment and reads wrong for the compact ledger style). */

export function formatMoney(amountMinor: number, locale = 'es-CO'): string {
  const sign = amountMinor < 0 ? '-' : '';
  const value = Math.abs(amountMinor);
  try {
    return `${sign}$${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)}`;
  } catch {
    return `${sign}$${value}`;
  }
}

export function formatMoneyFromMinorSafe(amountMinor: number | null | undefined): string {
  if (amountMinor == null || Number.isNaN(amountMinor)) return '—';
  return formatMoney(amountMinor);
}

/* Signed money for table cells and drawer headers: "+$41.000" / "-$12.500" / "$0". */
export function formatSignedMoney(amountMinor: number): string {
  if (amountMinor > 0) return `+${formatMoney(amountMinor)}`;
  return formatMoney(amountMinor);
}
