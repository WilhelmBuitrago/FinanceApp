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

/* The shape the UI actually consumes: the selectors filter by type and
   the review cards read id/name, so the pages normalize once instead of
   each reinventing an anonymous object. */

export interface CategoryOption {
  id: string;
  name: string;
  type: string;
}
