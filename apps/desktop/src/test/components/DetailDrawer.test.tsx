import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { getDocument } from '../../api/documents';
import { getFinancialEvent } from '../../api/financialEvents';
import { getSourceEntry } from '../../api/sourceEntries';
import { getCategories } from '../../api/categories';
import { deleteMovement, setMovementCategory } from '../../api/movements';
import { invalidateCollections } from '../../lib/dataBus';
import { DetailDrawer } from '../../components/DetailDrawer';
import type { Movement } from '../../types/movement';

vi.mock('../../api/documents', () => ({ getDocument: vi.fn() }));
vi.mock('../../api/financialEvents', () => ({ getFinancialEvent: vi.fn() }));
vi.mock('../../api/sourceEntries', () => ({ getSourceEntry: vi.fn() }));
vi.mock('../../api/categories', () => ({ getCategories: vi.fn() }));
vi.mock('../../api/movements', () => ({ setMovementCategory: vi.fn(), deleteMovement: vi.fn() }));
vi.mock('../../lib/dataBus', () => ({
  invalidateCollections: vi.fn(),
  useCollectionVersion: () => '0',
}));

const movement: Movement = {
  ID: 'mv_1',
  FinancialEventID: 'fe_1',
  AccountID: 'acc_1',
  SourceEntryID: 'se_1',
  MovementDate: '2026-10-01',
  AmountMinor: -2500,
  Description: 'Mercado',
  Status: 'ACTIVE',
  categoryName: '',
  CreatedAt: '2026-10-01T00:00:00Z',
  UpdatedAt: '2026-10-01T00:00:00Z',
};

function setup(overrides: Partial<{ movement: Movement | null; onClose: () => void }> = {}) {
  const props = {
    movement,
    accountName: 'Nu',
    onClose: vi.fn(),
    ...overrides,
  };
  render(
    <MemoryRouter>
      <DetailDrawer {...props} />
    </MemoryRouter>,
  );
  return props;
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getFinancialEvent).mockResolvedValue({
    ID: 'fe_1',
    EventType: 'EXPENSE',
    EventDate: '2026-10-01',
    Status: 'ACTIVE',
    VerificationStatus: 'VERIFIED',
    Description: 'Mercado',
    CreatedAt: '2026-10-01T00:00:00Z',
    UpdatedAt: '2026-10-01T00:00:00Z',
  } as never);
  vi.mocked(getSourceEntry).mockResolvedValue({
    ID: 'se_1',
    DocumentID: 'doc_1',
    RawDescription: 'MERCADO',
    NormalizedDescription: 'Mercado',
    PageNumber: 1,
    RowNumber: 2,
  } as never);
  vi.mocked(getDocument).mockResolvedValue({
    ID: 'doc_1',
    DocumentType: 'STATEMENT',
    Status: 'IMPORTED',
    CreatedAt: '2026-10-02T00:00:00Z',
    UpdatedAt: '2026-10-02T00:00:00Z',
  } as never);
  vi.mocked(getCategories).mockResolvedValue([
    {
      ID: 'cat_1',
      Name: 'Supermercado',
      CategoryType: 'EXPENSE',
      Status: 'ACTIVE',
      CreatedAt: '',
      UpdatedAt: '',
    },
    {
      ID: 'cat_2',
      Name: 'Sueldo',
      CategoryType: 'INCOME',
      Status: 'ACTIVE',
      CreatedAt: '',
      UpdatedAt: '',
    },
  ] as never);
  vi.mocked(setMovementCategory).mockResolvedValue(undefined);
  vi.mocked(deleteMovement).mockResolvedValue(undefined);
});

describe('DetailDrawer', () => {
  it('walks the whole traceability chain: movement, event, entry, document', async () => {
    setup();
    await waitFor(() => expect(screen.getByText('Movimiento')).toBeInTheDocument());
    expect(screen.getByText('Evento financiero')).toBeInTheDocument();
    expect(screen.getByText('Entrada documental')).toBeInTheDocument();
    expect(screen.getByText('Documento')).toBeInTheDocument();
    // The evidence branch shows the raw line, not just the normalized one.
    expect(screen.getByText('MERCADO')).toBeInTheDocument();
  });

  it('offers the user categories of the movement type, plus "Sin categoría"', async () => {
    setup();
    await waitFor(() =>
      expect(screen.getByLabelText('Categoría del movimiento')).toBeInTheDocument(),
    );
    expect(screen.getByRole('option', { name: 'Supermercado' })).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'Sueldo' })).not.toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Sin categoría' })).toBeInTheDocument();
  });

  it('assigns a category and publishes the invalidation', async () => {
    setup();
    const select = await waitFor(() => screen.getByLabelText('Categoría del movimiento'));
    fireEvent.change(select, { target: { value: 'cat_1' } });
    await waitFor(() => expect(setMovementCategory).toHaveBeenCalledWith('mv_1', 'cat_1'));
    expect(invalidateCollections).toHaveBeenCalledWith('movements');
  });

  it('leaves the movement uncategorized when the human picks the empty option', async () => {
    setup();
    const select = await waitFor(() => screen.getByLabelText('Categoría del movimiento'));
    fireEvent.change(select, { target: { value: '' } });
    await waitFor(() => expect(setMovementCategory).toHaveBeenCalledWith('mv_1', null));
  });

  it('confirms deletion through the modal, then deletes and invalidates', async () => {
    const props = setup();
    await waitFor(() => expect(screen.getByText('Movimiento')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'Eliminar movimiento' }));
    // The consequence is named before anything happens.
    expect(screen.getByText(/no se puede deshacer/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Eliminar' }));
    await waitFor(() => expect(deleteMovement).toHaveBeenCalledWith('mv_1'));
    expect(invalidateCollections).toHaveBeenCalledWith('movements', 'dashboard');
    expect(props.onClose).toHaveBeenCalled();
  });

  it('cancels the deletion without touching the data', async () => {
    setup();
    await waitFor(() => expect(screen.getByText('Movimiento')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'Eliminar movimiento' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancelar' }));
    expect(deleteMovement).not.toHaveBeenCalled();
  });
});
