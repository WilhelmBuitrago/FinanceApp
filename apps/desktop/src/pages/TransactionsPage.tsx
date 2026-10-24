import { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useAccounts } from '../hooks/useAccounts';
import { useMovements } from '../hooks/useMovements';
import { invalidateCollections } from '../lib/dataBus';
import { formatDate } from '../lib/formatDate';
import { formatSignedMoney } from '../lib/formatMoney';
import { DetailDrawer } from '../components/DetailDrawer';
import { IconClose, IconSearch } from '../components/icons';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { SkeletonRows } from '../components/ui/SkeletonRows';
import { StatusBadge } from '../components/ui/StatusBadge';
import type { Movement } from '../types/movement';

/* Movimientos — the product's main query screen.
   NO VISUAL DEDUPLICATION, ever: every documented record renders as its own
   row regardless of identical dates, descriptions or amounts. Search and
   filters are URL-addressable (back/refresh preserve them). Rows open the
   traceability drawer: Movement -> Event -> Source Entry -> Document. */

const MOVEMENT_STATUS_LABELS: Record<string, string> = {
  ACTIVE: 'Activos',
  CANCELLED: 'Cancelados',
  REVERSED: 'Revertidos',
};

const tableHead =
  'px-3 py-1.5 text-left font-medium uppercase tracking-[0.08em] text-meta text-ink-3';

function FilterChip({ label, onRemove }: { label: string; onRemove: () => void }) {
  return (
    <span className="inline-flex h-7 items-center gap-1.5 rounded-[8px] border border-line bg-surface px-2 text-meta text-ink-2">
      {label}
      <button
        onClick={onRemove}
        aria-label={`Quitar filtro ${label}`}
        className="flex items-center text-ink-3 transition-colors hover:text-ink"
      >
        <IconClose />
      </button>
    </span>
  );
}

const secondaryButton =
  'inline-flex h-8 items-center rounded-[8px] border border-line bg-raised px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';

const selectClass =
  'h-8 rounded-[8px] border border-line bg-bg px-2 text-body text-ink focus:border-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';

export default function TransactionsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const q = searchParams.get('q') ?? '';
  const cuenta = searchParams.get('cuenta') ?? '';
  const estado = searchParams.get('estado') ?? '';

  const setParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams);
    if (value) next.set(key, value);
    else next.delete(key);
    setSearchParams(next, { replace: true });
    // A different filter set invalidates the open drawer's row.
    setSelected(null);
  };
  const clearFilters = () => {
    setSearchParams(new URLSearchParams(), { replace: true });
    setSelected(null);
  };

  const movements = useMovements();
  const accounts = useAccounts();

  const accountNames = useMemo(() => {
    const map = new Map<string, string>();
    for (const account of accounts.data ?? []) map.set(account.ID, account.Name);
    return map;
  }, [accounts.data]);

  /* Every documented record stays an independent row — filter only, never merge. */
  const filtered = useMemo(() => {
    if (!movements.data) return [];
    const needle = q.trim().toLowerCase();
    return movements.data.filter(
      (mv) =>
        (needle === '' || (mv.Description ?? '').toLowerCase().includes(needle)) &&
        (cuenta === '' || mv.AccountID === cuenta) &&
        (estado === '' || mv.Status === estado),
    );
  }, [movements.data, q, cuenta, estado]);

  const accountOptions = useMemo(() => {
    const map = new Map<string, string>();
    for (const mv of movements.data ?? []) {
      if (!map.has(mv.AccountID))
        map.set(mv.AccountID, accountNames.get(mv.AccountID) ?? mv.AccountID);
    }
    return [...map.entries()].map(([id, label]) => ({ id, label }));
  }, [movements.data, accountNames]);

  const statusOptions = useMemo(() => {
    const set = new Set<string>();
    for (const mv of movements.data ?? []) set.add(mv.Status);
    return [...set];
  }, [movements.data]);

  /* Estado column only exists when there is something to show (non-ACTIVE).
     Silence = active: an all-ACTIVE table keeps its density. */
  const showEstado = useMemo(
    () => (movements.data ?? []).some((mv) => mv.Status !== 'ACTIVE'),
    [movements.data],
  );

  const [selected, setSelected] = useState<Movement | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  // Ctrl+K focuses search
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        searchRef.current?.focus();
        searchRef.current?.select();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const filtersActive = q !== '' || cuenta !== '' || estado !== '';
  const total = movements.data?.length ?? 0;

  return (
    <div
      className={`transition-[padding] duration-[220ms] ease-out-tk ${
        selected ? 'pr-[calc(min(440px,40vw)_+_16px)]' : ''
      }`}
    >
      {/* Toolbar: search + filters + count */}
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <div className="relative">
          <span className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-ink-3">
            <IconSearch />
          </span>
          <input
            ref={searchRef}
            type="text"
            value={q}
            onChange={(e) => setParam('q', e.target.value)}
            placeholder="Buscar por descripción… (Ctrl+K)"
            aria-label="Buscar movimientos por descripción"
            className="h-8 w-72 rounded-tk border border-line bg-bg pl-8 pr-3 text-body text-ink placeholder:text-ink-3 focus:border-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          />
        </div>
        <select
          value={cuenta}
          onChange={(e) => setParam('cuenta', e.target.value)}
          aria-label="Filtrar por cuenta"
          className={selectClass}
        >
          <option value="">Todas las cuentas</option>
          {accountOptions.map((opt) => (
            <option key={opt.id} value={opt.id}>
              {opt.label}
            </option>
          ))}
        </select>
        <select
          value={estado}
          onChange={(e) => setParam('estado', e.target.value)}
          aria-label="Filtrar por estado"
          className={selectClass}
        >
          <option value="">Todos los estados</option>
          {statusOptions.map((s) => (
            <option key={s} value={s}>
              {MOVEMENT_STATUS_LABELS[s] ?? s}
            </option>
          ))}
        </select>
        <span className="ml-auto text-meta text-ink-3">
          {filtersActive ? `${filtered.length} de ${total}` : total} movimientos
        </span>
      </div>

      {/* Active filter chips */}
      {filtersActive && (
        <div className="mb-2 flex flex-wrap items-center gap-2">
          {q !== '' && <FilterChip label={`“${q}”`} onRemove={() => setParam('q', '')} />}
          {cuenta !== '' && (
            <FilterChip
              label={accountNames.get(cuenta) ?? cuenta}
              onRemove={() => setParam('cuenta', '')}
            />
          )}
          {estado !== '' && (
            <FilterChip
              label={MOVEMENT_STATUS_LABELS[estado] ?? estado}
              onRemove={() => setParam('estado', '')}
            />
          )}
          <button
            onClick={clearFilters}
            className="text-meta font-medium text-ink-3 transition-colors hover:text-ink"
          >
            Limpiar filtros
          </button>
        </div>
      )}

      {/* List */}
      {movements.loading ? (
        <SkeletonRows rows={8} />
      ) : movements.error ? (
        <ErrorState
          message="No se pudo cargar los movimientos desde el servidor."
          onRetry={() => invalidateCollections('movements', 'accounts')}
        />
      ) : total === 0 ? (
        <EmptyState
          title="Sin movimientos"
          description="Los movimientos se generan al importar documentos bancarios."
          action={
            <Link to="/documents" className={secondaryButton}>
              Ir a Documentos
            </Link>
          }
        />
      ) : filtered.length === 0 ? (
        <EmptyState
          title="Sin resultados"
          description={`Ningún movimiento coincide con “${q || 'los filtros aplicados'}”.`}
          action={
            <button onClick={clearFilters} className={secondaryButton}>
              Limpiar filtros
            </button>
          }
        />
      ) : (
        <div className="overflow-x-auto rounded-tk border border-line bg-surface">
          <table
            className="w-full text-body"
            onKeyDown={(e) => {
              if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
              const rows = Array.from(e.currentTarget.querySelectorAll('tbody tr'));
              const index = rows.indexOf(document.activeElement as HTMLElement);
              const next = e.key === 'ArrowDown' ? index + 1 : index - 1;
              if (next >= 0 && next < rows.length) {
                e.preventDefault();
                (rows[next] as HTMLElement).focus();
              }
            }}
          >
            <thead className="bg-raised">
              <tr>
                <th className={`${tableHead} w-24`}>Fecha</th>
                <th className={tableHead}>Descripción</th>
                <th className={`${tableHead} w-32`}>Categoría</th>
                <th className={`${tableHead} w-40`}>Cuenta</th>
                {showEstado && <th className={`${tableHead} w-28`}>Estado</th>}
                <th className={`${tableHead.replace('text-left', 'text-right')} w-32`}>Monto</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((mv) => {
                const isSelected = selected?.ID === mv.ID;
                return (
                  <tr
                    key={mv.ID}
                    tabIndex={0}
                    onClick={() => setSelected(mv)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') setSelected(mv);
                    }}
                    aria-label={`${formatDate(mv.MovementDate)} — ${mv.Description ?? 'Sin descripción'} — ${formatSignedMoney(mv.AmountMinor)}`}
                    className={`relative cursor-pointer border-t border-line/60 transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent ${
                      isSelected ? 'bg-accent-wash' : ''
                    }`}
                  >
                    {isSelected && (
                      <span
                        className="absolute bottom-0 left-0 top-0 w-[2px] bg-accent"
                        aria-hidden="true"
                      />
                    )}
                    <td className="whitespace-nowrap px-3 py-2 text-meta text-ink-3">
                      {formatDate(mv.MovementDate)}
                    </td>
                    <td
                      className="w-full max-w-0 truncate px-3 py-2 text-ink"
                      title={mv.Description ?? undefined}
                    >
                      {mv.Description ?? '—'}
                    </td>
                    <td className="truncate px-3 py-2 text-ink-2">{mv.categoryName || '—'}</td>
                    <td
                      className="truncate px-3 py-2 text-ink-2"
                      title={accountNames.get(mv.AccountID) ?? undefined}
                    >
                      {accountNames.get(mv.AccountID) ?? '—'}
                    </td>
                    {showEstado && (
                      <td className="px-3 py-2">
                        {mv.Status !== 'ACTIVE' ? <StatusBadge status={mv.Status} /> : null}
                      </td>
                    )}
                    <td
                      className={`whitespace-nowrap px-3 py-2 text-right font-medium tabular-nums ${
                        mv.AmountMinor > 0
                          ? 'text-pos'
                          : mv.AmountMinor < 0
                            ? 'text-neg'
                            : 'text-ink'
                      }`}
                    >
                      {formatSignedMoney(mv.AmountMinor)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <DetailDrawer
        movement={selected}
        accountName={selected ? (accountNames.get(selected.AccountID) ?? null) : null}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
