/*
 * Amount parsing for the manual drawer.
 *
 * The ledger stores integer centavos (amount_minor) and the form is
 * free text: a Colombian user writes "25,5" for twenty-five and a half
 * pesos, and parseFloat reads that as 25 — the comma is dropped and
 * the saved amount is wrong by a factor of ten without any complaint.
 * This parser reads both conventions and hands the integer the API
 * expects:
 *
 *   "25,5"        → 2550      (comma is the decimal separator)
 *   "25.5"        → 2550      (dot, one or two decimals)
 *   "25.500"      → 2550000   (dot with exactly three digits = thousands)
 *   "1.250.000"   → 125000000
 *   "1.250.000,5" → 125000050 (last separator is the decimal one)
 *   "2500"        → 250000
 *
 * Thousands groups carry exactly three digits, the first group any
 * count: anything else — letters, spaces, a sign, three decimals after
 * a comma — returns null and the caller reports the usual validation
 * error instead of guessing.
 */

/** Parses a human-written amount into integer centavos, or null. */
export function parseAmountMinor(input: string): number | null {
  const s = input.trim();
  if (!s) return null;
  if (!/^[\d.,]+$/.test(s)) return null;

  const m = s.match(/^(\d+)(?:([.,])(\d+))?(?:([.,])(\d+))?(?:([.,])(\d+))?$/);
  if (!m) return null;

  const groups = [m[1]];
  let decimals: string | null = null;

  const pairs = [
    [m[2], m[3]],
    [m[4], m[5]],
    [m[6], m[7]],
  ].filter((pair) => pair[0] !== undefined && pair[0] !== '');

  for (let i = 0; i < pairs.length; i++) {
    const sep = pairs[i][0];
    const digits = pairs[i][1] as string;
    const isLast = i === pairs.length - 1;
    const isDecimal = isLast && digits.length <= 2;
    if (isDecimal) {
      decimals = digits;
      continue;
    }
    // Every other separator groups thousands: exactly three digits.
    if (digits.length !== 3) return null;
    if (sep === ',' && !isLast) return null;
    groups.push(digits);
  }

  const units = Number(groups.join(''));
  if (!Number.isSafeInteger(units)) return null;
  if (decimals === null) return units * 100;

  const cents = Math.round(Number('0.' + decimals) * 100);
  return units * 100 + cents;
}
