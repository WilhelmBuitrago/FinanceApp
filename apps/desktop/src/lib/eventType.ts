/* Spanish labels for EventType (S11 domain — closed set from the DB CHECK). */

const EVENT_TYPE_LABELS: Record<string, string> = {
  INCOME: 'Ingreso',
  EXPENSE: 'Gasto',
  TRANSFER: 'Transferencia',
  CASH_WITHDRAWAL: 'Retiro de efectivo',
  CASH_DEPOSIT: 'Depósito de efectivo',
  DEBIT_PURCHASE: 'Compra con débito',
  CREDIT_PURCHASE: 'Compra con crédito',
  CREDIT_PAYMENT: 'Pago de tarjeta',
  CREDIT_ADVANCE: 'Adelanto en efectivo',
  REFUND: 'Reembolso',
  INTEREST_INCOME: 'Interés',
  ADJUSTMENT: 'Ajuste',
};

export function eventTypeLabel(eventType: string): string {
  return EVENT_TYPE_LABELS[eventType] ?? eventType;
}
