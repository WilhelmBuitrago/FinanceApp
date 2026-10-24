/* The JSON payload a parked review item carries (backend review_items).
   Every key is optional by design: the payloads come from different
   triggers and from data parked by older versions, and the UI reads them
   defensively — parking untrusted data and reading it tolerantly is the
   product rule (never fabricate, never drop). The type names the keys the
   UI actually reads; anything else stays valid at runtime. */

export interface ReviewPayload {
  // CATEGORY: the movement that lost its category.
  movement_id?: string;
  category_id?: string | null;
  category_name?: string;
  // DUPLICATE / EXTRACTION: the movement data to confirm or correct.
  account_id?: string;
  amount_minor?: number;
  amount_original?: string;
  movement_date?: string;
  description?: string;
  description_original?: string;
  date?: string;
  date_original?: string;
  type?: string;
  reason?: string;
  collides_with?: string;
  // CATEGORY: the kind filter the card applies.
  category_type?: string;
  // DOCUMENT_PERIOD: what the processor could not read, plus the suggestion.
  raw_period?: string;
  document_id?: string;
  suggested_month?: number;
  suggested_year?: number;
}
