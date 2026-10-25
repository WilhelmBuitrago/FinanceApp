export type UseQueryState<T> =
  | { loading: true; data?: undefined; error?: undefined; empty?: undefined }
  | { loading: false; data: T; error?: undefined; empty: boolean }
  | { loading: false; data?: undefined; error: Error; empty?: undefined };
