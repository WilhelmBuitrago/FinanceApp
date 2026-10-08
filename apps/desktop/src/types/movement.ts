/* API contract (verified against live backend): Go domain structs are
   serialized WITHOUT json tags, so keys are PascalCase. */

export interface Movement {
  ID: string;
  FinancialEventID: string;
  AccountID: string;
  SourceEntryID?: string | null;
  MovementDate: string;
  AmountMinor: number;
  Description?: string | null;
  Status: string;
  MetadataJson?: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
