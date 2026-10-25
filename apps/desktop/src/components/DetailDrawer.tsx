import { getDocument } from '../api/documents';
import { getFinancialEvent } from '../api/financialEvents';
import { getSourceEntry } from '../api/sourceEntries';
import { useMovementActions } from '../hooks/useMovementActions';
import { useResource } from '../hooks/useResource';
import { formatSignedMoney, signedAmountClass } from '../lib/formatMoney';
import type { Movement } from '../types/movement';
import { DocumentDetail } from './drawer/DocumentDetail';
import { FinancialEventDetail } from './drawer/FinancialEventDetail';
import { MovementDetail } from './drawer/MovementDetail';
import { SourceEntryDetail } from './drawer/SourceEntryDetail';
import { DrawerBody, DrawerCloseButton, DrawerHeader, DrawerShell } from './DrawerShell';

/* Traceability drawer — the product's core navigation:
   Movement -> Financial Event -> Source Entry -> Document.
   Push-style (content shrinks, list stays visible). Esc or ✕ closes.

   This file composes the graph: each section resolves, fails or retries
   on its own and lives in components/drawer/; the movement's mutations
   live in useMovementActions. */

interface DetailDrawerProps {
  movement: Movement | null;
  accountName: string | null;
  onClose: () => void;
}

export function DetailDrawer({ movement, accountName, onClose }: DetailDrawerProps) {
  // The traceability graph behind the movement: all three refetch when a
  // mutation invalidates 'movements' (the category lives on the event, a
  // delete takes the movement down).
  const event = useResource(movement?.FinancialEventID, getFinancialEvent, ['movements']);
  const entry = useResource(movement?.SourceEntryID ?? null, getSourceEntry, ['movements']);
  const document_ = useResource(
    entry.status === 'ready' ? entry.data.DocumentID : null,
    getDocument,
    ['movements'],
  );

  // Until the event resolves, every category stays available (the
  // create form's rule, same in useMovementActions).
  const eventType = event.status === 'ready' ? event.data.EventType : '';
  const actions = useMovementActions(movement, eventType, onClose);

  if (!movement) return null;

  return (
    <DrawerShell ariaLabel="Detalle del movimiento" onClose={onClose}>
      <DrawerHeader>
        <p
          className={`text-amount font-semibold tabular-nums ${signedAmountClass(movement.AmountMinor)}`}
        >
          {formatSignedMoney(movement.AmountMinor)}
        </p>
        <DrawerCloseButton onClose={onClose} />
      </DrawerHeader>

      <DrawerBody>
        <MovementDetail
          movement={movement}
          accountName={accountName}
          categories={actions.categories}
          savingCategory={actions.savingCategory}
          categoryError={actions.categoryError}
          onAssignCategory={actions.assignCategory}
          confirmingDelete={actions.confirmingDelete}
          deleting={actions.deleting}
          deleteError={actions.deleteError}
          onRequestDelete={actions.requestDelete}
          onCancelDelete={actions.cancelDelete}
          onConfirmDelete={actions.confirmDelete}
        />
        <FinancialEventDetail event={event} />
        <SourceEntryDetail entry={entry} manual={movement.SourceEntryID == null} />
        {entry.status === 'ready' && <DocumentDetail document={document_} />}
      </DrawerBody>
    </DrawerShell>
  );
}
