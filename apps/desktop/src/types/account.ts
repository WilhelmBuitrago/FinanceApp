/* API contract (verified against live backend): PascalCase keys. */

export interface Account {
  ID: string;
  UserID: string;
  Name: string;
  AccountType: string;
  ParentAccountID?: string | null;
  Currency: string;
  Status: string;
  Institution?: string | null;
  MetadataJson?: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
