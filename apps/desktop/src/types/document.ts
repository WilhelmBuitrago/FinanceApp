/* API contract (verified against live backend): PascalCase keys. */

export interface Document {
  ID: string;
  AccountID?: string | null;
  DocumentType: string;
  Institution?: string | null;
  DocumentName?: string | null;
  PeriodStart?: string | null;
  PeriodEnd?: string | null;
  DocumentDate?: string | null;
  FileHash?: string | null;
  SourceUri?: string | null;
  Status: string;
  MetadataJson?: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
