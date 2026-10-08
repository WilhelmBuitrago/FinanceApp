/* API contract: PascalCase keys (Go domain structs without json tags). */

export interface ImportJob {
  ID: string;
  DocumentID: string;
  Status: string;
  CreatedAt: string;
  StartedAt?: string | null;
  FinishedAt?: string | null;
  Error?: string | null;
  ProcessingID?: string | null;
}
