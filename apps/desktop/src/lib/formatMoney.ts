/* Money formatting for FinanceApp.
   amountMinor ahora es centavos. Se divide por 100 para mostrar pesos con 2 decimales.
   NOTA IMPORTANTE: El sistema usa pesos decimales internamente como centavos enteros.
   amount_minor = 150042 → UI $1.500,42
   amount_minor = 15044 → UI $150,44
   Nunca almacenar decimales en amount_minor, solo enteros de centavos.
*/

export function formatMoney(amountMinor: number, locale = 'es-CO'): string {
  const sign = amountMinor < 0 ? '-' : '';
  const value = Math.abs(amountMinor) / 100;
  try {
    return `${sign}$${new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value)}`;
  } catch {
    return `${sign}$${value.toFixed(2)}`;
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
