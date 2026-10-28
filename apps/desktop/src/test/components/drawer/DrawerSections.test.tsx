import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import { FinancialEventDetail } from '../../../components/drawer/FinancialEventDetail';
import { SourceEntryDetail } from '../../../components/drawer/SourceEntryDetail';
import { DocumentDetail } from '../../../components/drawer/DocumentDetail';
import type { ResourceHandle } from '../../../hooks/useResource';
import type { FinancialEvent } from '../../../types/financial-event';
import type { SourceEntry } from '../../../types/source-entry';
import type { Document } from '../../../types/document';

/* The three async sections of the traceability drawer: each resolves,
   fails (with its retry) or shows data on its own — one section failing
   never blanks the others. */

const event: FinancialEvent = {
  ID: 'fe_1',
  EventType: 'EXPENSE',
  EventDate: '2026-10-20',
  Status: 'ACTIVE',
  VerificationStatus: 'UNVERIFIED',
  Description: 'Mercado',
  Confidence: 0.85,
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
};

const entry: SourceEntry = {
  ID: 'se_1',
  DocumentID: 'doc_1',
  RawDescription: 'MERCADO DEL BARRIO',
  RawAmountMinor: -150000,
  RawDate: '20/10/2026',
  NormalizedDescription: 'Mercado del barrio',
  NormalizedAmountMinor: -150000,
  NormalizedDate: '2026-10-20',
  PageNumber: 3,
  RowNumber: 2,
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
};

const document: Document = {
  ID: 'doc_1',
  AccountID: 'acc-1',
  DocumentType: 'PDF',
  Institution: 'NU',
  DocumentName: 'extracto.pdf',
  PeriodStart: '2026-10-01',
  PeriodEnd: '2026-10-31',
  Status: 'IMPORTED',
  CreatedAt: '2026-10-20T12:00:00Z',
  UpdatedAt: '2026-10-20T12:00:00Z',
};

const handle = <T,>(state: ResourceHandle<T>['status'], data?: T): ResourceHandle<T> =>
  ({ status: state, data, retry: vi.fn() }) as ResourceHandle<T>;

describe('FinancialEventDetail', () => {
  it('shows the section loading while the event resolves', () => {
    render(<FinancialEventDetail event={handle<FinancialEvent>('loading')} />);

    expect(screen.queryByText('Tipo')).toBeNull();
  });

  it('offers the retry when the event fails, without blanking the drawer', () => {
    const retry = vi.fn();
    render(
      <FinancialEventDetail event={{ status: 'error', retry } as ResourceHandle<FinancialEvent>} />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Reintentar' }));
    expect(retry).toHaveBeenCalledTimes(1);
  });

  it('renders the event data: type label, date, verification and confidence', () => {
    render(<FinancialEventDetail event={handle<FinancialEvent>('ready', event)} />);

    expect(screen.getByText('Gasto')).toBeTruthy();
    expect(screen.getByText('20 oct 2026')).toBeTruthy();
    expect(screen.getByText('Sin verificar')).toBeTruthy();
    expect(screen.getByText('Mercado')).toBeTruthy();
    expect(screen.getByText('85%')).toBeTruthy();
  });

  it('shows a dash when the event carries no confidence', () => {
    render(
      <FinancialEventDetail
        event={handle<FinancialEvent>('ready', { ...event, Confidence: null })}
      />,
    );

    expect(screen.getByText('—')).toBeTruthy();
  });
});

describe('SourceEntryDetail', () => {
  it('says a manually registered movement has no entry, instead of failing', () => {
    render(<SourceEntryDetail entry={handle<SourceEntry>('loading')} manual />);

    expect(screen.getByText('Sin entrada documental — registrado manualmente.')).toBeTruthy();
  });

  it('shows the document evidence next to the normalized values', () => {
    render(<SourceEntryDetail entry={handle<SourceEntry>('ready', entry)} manual={false} />);

    expect(screen.getByText('Evidencia del documento')).toBeTruthy();
    expect(screen.getByText('MERCADO DEL BARRIO')).toBeTruthy();
    expect(screen.getByText('Fecha original: 20/10/2026')).toBeTruthy();
    expect(screen.getByText('-$1.500,00')).toBeTruthy();
    expect(screen.getByText('Mercado del barrio')).toBeTruthy();
    expect(screen.getByText('pág. 3 · fila 2')).toBeTruthy();
  });

  it('falls back to a dash for every missing value', () => {
    render(
      <SourceEntryDetail
        entry={handle<SourceEntry>('ready', {
          ID: 'se_2',
          DocumentID: 'doc_1',
          CreatedAt: '2026-10-20T00:00:00Z',
          UpdatedAt: '2026-10-20T00:00:00Z',
        })}
        manual={false}
      />,
    );

    expect(screen.getAllByText('—').length).toBeGreaterThanOrEqual(4);
  });

  it('offers the retry when the entry fails', () => {
    const retry = vi.fn();
    render(
      <SourceEntryDetail
        entry={{ status: 'error', retry } as ResourceHandle<SourceEntry>}
        manual={false}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Reintentar' }));
    expect(retry).toHaveBeenCalledTimes(1);
  });
});

describe('DocumentDetail', () => {
  const renderDoc = (doc: ResourceHandle<Document>) =>
    render(
      <MemoryRouter>
        <DocumentDetail document={doc} />
      </MemoryRouter>,
    );

  it('shows the section loading while the document resolves', () => {
    renderDoc(handle<Document>('loading'));

    expect(screen.queryByText('Nombre')).toBeNull();
  });

  it('renders the document data with its financial period, not the upload date', () => {
    renderDoc(handle<Document>('ready', document));

    expect(screen.getByText('extracto.pdf')).toBeTruthy();
    expect(screen.getByText('PDF')).toBeTruthy();
    expect(screen.getByText('01 oct 2026 — 31 oct 2026')).toBeTruthy();
    expect(screen.getByText('Importado')).toBeTruthy();
  });

  it('shows a dash when the document carries no period', () => {
    renderDoc(handle<Document>('ready', { ...document, PeriodStart: null, PeriodEnd: null }));

    expect(screen.getByText('—')).toBeTruthy();
  });

  it('links back to the documents page', () => {
    renderDoc(handle<Document>('ready', document));

    expect(screen.getByRole('link', { name: 'Ver en Documentos' }).getAttribute('href')).toBe(
      '/documents',
    );
  });
});
