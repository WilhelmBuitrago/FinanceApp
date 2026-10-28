import { useCallback, useSyncExternalStore } from 'react';

/* Minimal shared invalidation bus. A mutation publishes WHICH collections
   went stale; every hook subscribed to them refetches on the next notify.
   It carries no data — no cache, only "these collections changed" — so it
   can never serve a stale value, and a screen showing shared data stays
   fresh without the mutator knowing who is listening.

   One tradeoff, documented: two invalidations in different ticks refetch
   twice (no coalescing). Correct over clever; if it ever matters, the
   coalescing goes here and nowhere else. */

export type DataCollection =
  'accounts' | 'movements' | 'categories' | 'documents' | 'review' | 'dashboard';

const versions = new Map<DataCollection, number>();
const listeners = new Set<() => void>();

export function invalidateCollections(...collections: DataCollection[]): void {
  if (collections.length === 0) return;
  for (const collection of collections) {
    versions.set(collection, (versions.get(collection) ?? 0) + 1);
  }
  for (const listener of listeners) listener();
}

/* Version stamp of the given collections: a string signature that changes
   whenever any of them is invalidated. Equal stamp means "unchanged".
   Strings over arithmetic so different tuples can never collide. */
export function useCollectionVersion(...collections: DataCollection[]): string {
  const key = collections.join(',');
  const subscribe = useCallback((listener: () => void) => {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  }, []);
  const getSnapshot = useCallback(
    () =>
      key
        .split(',')
        .map((c) => `${c}=${versions.get(c as DataCollection) ?? 0}`)
        .join('&'),
    [key],
  );
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}
