/* API contract (verified against live backend): Go domain structs are
   serialized WITHOUT json tags, so keys are PascalCase. */

export interface Category {
  ID: string;
  Name: string;
  ParentID?: string | null;
  CategoryType?: string | null;
  Status: string;
  CreatedAt: string;
  UpdatedAt: string;
}
