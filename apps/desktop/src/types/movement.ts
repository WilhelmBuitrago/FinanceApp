/* API contract (verified against live backend): Go domain structs are
   serialized WITHOUT json tags, so keys are PascalCase. The two category
   fields are the exception: the movement list handlers wrap the domain
   struct in a DTO with camelCase json tags (handlers_movements.go,
   handlers_accounts.go), so they arrive camelCase. */

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
  categoryName: string;
  categoryId?: string;
}
