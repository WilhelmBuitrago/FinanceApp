/* API contract (verified against live backend): PascalCase keys. */

export interface FinancialEvent {
  ID: string;
  EventType: string;
  EventDate?: string | null;
  Status: string;
  VerificationStatus: string;
  Description?: string | null;
  CategoryID?: string | null;
  Confidence?: number | null;
  MetadataJson?: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
