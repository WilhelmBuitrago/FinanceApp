/* Money formatting for FinanceApp.

   The ledger stores amount_minor as integer minor units and every
   movement now carries the currency of its account: COP renders by
   hand ("$1.500,00", the look the product was built with), and any
   other currency the account catalog admits formats through Intl so
   the symbol and the separators describe the actual money. An absent
   currency reads as COP: the product lives in Colombia and older
   payloads predate the field. */

const COP_LOCALE = 'es-CO';

function formatCop(amountMinor: number): string {
  const sign = amountMinor < 0 ? '-' : '';
  const value = Math.abs(amountMinor) / 100;
  try {
    return `${sign}$${new Intl.NumberFormat(COP_LOCALE, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value)}`;
  } catch {
    return `${sign}$${value.toFixed(2)}`;
  }
}

export function formatMoney(amountMinor: number, currency?: string): string {
  if (!currency || currency.toUpperCase() === 'COP') {
    return formatCop(amountMinor);
  }
  const sign = amountMinor < 0 ? '-' : '';
  const value = Math.abs(amountMinor) / 100;
  try {
    return (
      sign +
      new Intl.NumberFormat(COP_LOCALE, {
        style: 'currency',
        currency: currency.toUpperCase(),
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(value)
    );
  } catch {
    return formatCop(amountMinor);
  }
}

/* Signed money for table cells and drawer headers: "+$41.000" / "-$12.500" / "$0". */
export function formatSignedMoney(amountMinor: number, currency?: string): string {
  if (amountMinor > 0) return `+${formatMoney(amountMinor, currency)}`;
  return formatMoney(amountMinor, currency);
}

/* The color a signed amount carries in tables and drawer headers: the
   same three-way mapping was copy-pasted in four components before the
   audit. Sign semantics only, never decoration. */
export function signedAmountClass(amountMinor: number): string {
  if (amountMinor > 0) return 'text-pos';
  if (amountMinor < 0) return 'text-neg';
  return 'text-ink';
}
