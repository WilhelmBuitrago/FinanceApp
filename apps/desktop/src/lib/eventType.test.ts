import { describe, expect, it } from 'vitest';

import { EVENT_TYPE_LABELS, eventTypeLabel } from './eventType';

describe('eventTypeLabel', () => {
  it('labels the domain event types', () => {
    expect(eventTypeLabel('INCOME')).toBe('Ingreso');
    expect(eventTypeLabel('EXPENSE')).toBe('Gasto');
    expect(eventTypeLabel('CREDIT_PAYMENT')).toBe('Pago de tarjeta');
  });

  it('falls back to the raw value for an event type outside the catalog', () => {
    expect(eventTypeLabel('BRAND_NEW_EVENT')).toBe('BRAND_NEW_EVENT');
    expect(EVENT_TYPE_LABELS.TRANSFER).toBe('Transferencia');
  });
});
