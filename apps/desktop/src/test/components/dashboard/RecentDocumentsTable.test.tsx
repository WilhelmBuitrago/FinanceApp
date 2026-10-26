import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { RecentDocumentsTable } from '../../../components/dashboard/RecentDocumentsTable';
import type { UseQueryState } from '../../../types/api';
import type { Document } from '../../../types/document';

/* Recent documents (backend-limited to three): name, account, financial
   period, state and upload date, plus the honest states around it. */

const doc = (over: Partial<Document>): Document => ({
  ID: 'doc-1',
  AccountID: 'acc-1',
  DocumentType: 'PDF',
  Institution: 'NU',
  DocumentName: 'extracto.pdf',
  Status: 'IMPORTED',
  CreatedAt: '2026-10-20T12:00:00Z',
  UpdatedAt: '2026-10-20T12:00:00Z',
  ...over,
});

const ready = (data: Document[]): UseQueryState<Document[]> => ({
  loading: false,
  data,
  empty: data.length === 0,
});

const renderTable = (documents: UseQueryState<Document[]>) =>
  render(
    <MemoryRouter>
      <RecentDocumentsTable documents={documents} accountNames={new Map([['acc-1', 'Nu']])} />
    </MemoryRouter>,
  );

describe('RecentDocumentsTable', () => {
  it('shows the skeleton while loading and no rows', () => {
    renderTable({ loading: true });

    expect(screen.queryAllByRole('row')).toHaveLength(0);
  });

  it('shows the empty state when there is nothing to show', () => {
    renderTable(ready([]));

    expect(screen.getByText('Sin documentos')).toBeTruthy();
  });

  it('renders name, account, period, status and upload date', () => {
    renderTable(
      ready([doc({ PeriodStart: '2026-10-01', PeriodEnd: '2026-10-31', Status: 'IMPORTED' })]),
    );

    expect(screen.getByText('extracto.pdf')).toBeTruthy();
    expect(screen.getByText('Nu')).toBeTruthy();
    expect(screen.getByText('01 oct 2026 — 31 oct 2026')).toBeTruthy();
    expect(screen.getByText('Importado')).toBeTruthy();
    expect(screen.getByText('20 oct 2026')).toBeTruthy();
  });

  it('shows at most three documents even when more arrive', () => {
    renderTable(
      ready([
        doc({ ID: 'a', DocumentName: 'a.pdf' }),
        doc({ ID: 'b', DocumentName: 'b.pdf' }),
        doc({ ID: 'c', DocumentName: 'c.pdf' }),
        doc({ ID: 'd', DocumentName: 'd.pdf' }),
      ]),
    );

    expect(screen.getByText('a.pdf')).toBeTruthy();
    expect(screen.getByText('c.pdf')).toBeTruthy();
    expect(screen.queryByText('d.pdf')).toBeNull();
  });

  it('falls back to a generic name when the document carries none', () => {
    renderTable(ready([doc({ DocumentName: null })]));

    expect(screen.getByText('Documento')).toBeTruthy();
  });
});
