/* API contract (verified against live backend): PascalCase keys. */

export interface SourceEntry {
  ID: string;
  DocumentID: string;
  SourceReference?: string | null;
  RawDescription?: string | null;
  RawAmountMinor?: number | null;
  RawDate?: string | null;
  NormalizedDescription?: string | null;
  NormalizedAmountMinor?: number | null;
  NormalizedDate?: string | null;
  PageNumber?: number | null;
  RowNumber?: number | null;
  Section?: string | null;
  ExtractionID?: string | null;
  MetadataJson?: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
