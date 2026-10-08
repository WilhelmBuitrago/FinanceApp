/* API contract (verified live): GET /institutions — lowercase snake keys. */

export interface Institution {
  code: string;
  name: string;
  supported_account_types: string[];
}
