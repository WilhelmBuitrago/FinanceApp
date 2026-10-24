import { getMovements } from '../api/movements';
import type { Movement } from '../types/movement';
import { useListQuery } from './useListQuery';

export function useMovements(refreshKey = 0) {
  return useListQuery<Movement[]>(getMovements, refreshKey);
}
