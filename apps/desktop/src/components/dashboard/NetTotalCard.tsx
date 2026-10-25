import { formatSignedMoney } from '../../lib/formatMoney';

/* Net total card — the headline figure and its per-account breakdown,
   both real sums over the user's movements. */

export interface AccountNet {
  id: string;
  name: string;
  net: number;
}

export function NetTotalCard({
  totalNet,
  perAccount,
}: {
  totalNet: number;
  perAccount: AccountNet[];
}) {
  return (
    <section aria-label="Neto total">
      <div className="rounded-card border border-line bg-surface p-5">
        <p className="text-meta text-ink-3">Neto total</p>
        <p className="text-headline font-semibold tracking-tight text-ink">
          {formatSignedMoney(totalNet)}
        </p>
        <div className="mt-4 space-y-2">
          {perAccount.map((a) => (
            <div key={a.id} className="flex items-center justify-between text-body">
              <span className="text-ink-2">{a.name}</span>
              <span className={a.net >= 0 ? 'text-pos' : 'text-neg'}>
                {formatSignedMoney(a.net)}
              </span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
