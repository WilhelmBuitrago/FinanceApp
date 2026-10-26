import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { AttentionSection } from '../../../components/dashboard/AttentionSection';
import type { Document } from '../../../types/document';

/* Documents that parked something: the section shows at most three and
   names the account and the financial period next to the status. */

const doc = (over: Partial<Document>): Document => ({
  ID: 'doc-1',
  AccountID: 'acc-1',
  DocumentType: 'PDF',
  Status: 'REVIEW_REQUIRED',
  CreatedAt: '2026-10-20T00:00:00Z',
  UpdatedAt: '2026-10-20T00:00:00Z',
  ...over,
});

const renderSection = (documents: Document[], accountNames = new Map([['acc-1', 'Nu']])) =>
  render(
    <MemoryRouter>
      <AttentionSection documents={documents} accountNames={accountNames} />
    </MemoryRouter>,
  );

describe('AttentionSection', () => {
  it('shows at most three documents and keeps the rest for the Documents page', () => {
    renderSection([
      doc({ ID: 'a', DocumentName: 'extracto-a.pdf' }),
      doc({ ID: 'b', DocumentName: 'extracto-b.pdf' }),
      doc({ ID: 'c', DocumentName: 'extracto-c.pdf' }),
      doc({ ID: 'd', DocumentName: 'extracto-d.pdf' }),
      doc({ ID: 'e', DocumentName: 'extracto-e.pdf' }),
    ]);

    expect(screen.getByText('extracto-a.pdf')).toBeTruthy();
    expect(screen.getByText('extracto-c.pdf')).toBeTruthy();
    expect(screen.queryByText('extracto-d.pdf')).toBeNull();
    expect(screen.queryByText('extracto-e.pdf')).toBeNull();
  });

  it('names the account and the financial period, and marks the status', () => {
    renderSection([
      doc({
        DocumentName: 'extracto.pdf',
        PeriodStart: '2026-10-01',
        PeriodEnd: '2026-10-31',
        Status: 'REVIEW_REQUIRED',
      }),
    ]);

    expect(screen.getByText('extracto.pdf')).toBeTruthy();
    expect(screen.getByText(/Nu · 01 oct 2026 — 31 oct 2026/)).toBeTruthy();
    expect(screen.getByText('Requiere revisión')).toBeTruthy();
  });

  it('falls back to a generic name when the document carries no name', () => {
    renderSection([doc({ DocumentName: null, AccountID: 'acc-9' })], new Map());

    expect(screen.getByText('Documento')).toBeTruthy();
    // An unknown account leaves the meta line without a name: no id is
    // invented in place of the missing one.
    expect(screen.getByText('Requiere revisión')).toBeTruthy();
  });
});
