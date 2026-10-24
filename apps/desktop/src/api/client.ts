/// <reference types="vite/client" />
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8081';

/* Every call is traced with the same grammar the backend HTTP middleware
   writes to its daily log ([HTTP] METHOD path -> status duration), so a
   frontend request and its backend handling read as one story in the
   console and the log file. No payloads: only method, path, status, time. */

function elapsedMs(started: number): number {
  return Math.round(performance.now() - started);
}

function errorLabel(data: unknown): string {
  if (typeof data === 'string') return data.slice(0, 120);
  if (data && typeof data === 'object' && 'error' in data) {
    const msg = (data as { error?: unknown }).error;
    return typeof msg === 'string' ? msg : '';
  }
  return '';
}

export class ApiError extends Error {
  status: number;
  response?: unknown;
  constructor(message: string, status: number, response?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.response = response;
  }
}

export async function fetchJson<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const method = init?.method ?? 'GET';
  const path = typeof input === 'string' ? new URL(input).pathname : new URL(input.url).pathname;
  const started = performance.now();
  let res: Response;
  try {
    res = await fetch(input, init);
  } catch (err) {
    console.error(`[HTTP] ${method} ${path} FAILED (${elapsedMs(started)}ms): ${String(err)}`);
    throw err;
  }
  const ms = elapsedMs(started);
  // 204/205 carry no body by definition — some servers still set a
  // Content-Type header, which must NOT trigger a JSON parse of "".
  if (res.status === 204 || res.status === 205) {
    console.info(`[HTTP] ${method} ${path} -> ${res.status} (${ms}ms)`);
    return undefined as T;
  }
  const contentType = res.headers.get('content-type') ?? '';
  let data: unknown;
  if (contentType.includes('application/json')) {
    data = await res.json();
  } else {
    data = await res.text();
  }
  if (!res.ok) {
    console.warn(`[HTTP] ${method} ${path} -> ${res.status} (${ms}ms) ERROR ${errorLabel(data)}`);
    throw new ApiError(`Request failed with ${res.status}`, res.status, data);
  }
  console.info(`[HTTP] ${method} ${path} -> ${res.status} (${ms}ms)`);
  return data as T;
}

export function apiGet<T>(path: string): Promise<T> {
  return fetchJson<T>(`${BASE_URL}${path}`, {
    headers: { 'Accept': 'application/json' },
  });
}

export function apiPost<T>(path: string, body: unknown): Promise<T> {
  return fetchJson<T>(`${BASE_URL}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
    body: JSON.stringify(body),
  });
}

export function apiPut<T>(path: string, body: unknown): Promise<T> {
  return fetchJson<T>(`${BASE_URL}${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
    body: JSON.stringify(body),
  });
}

export function apiPostForm<T>(path: string, form: FormData): Promise<T> {
  return fetchJson<T>(`${BASE_URL}${path}`, {
    method: 'POST',
    body: form,
  });
}

export function apiDelete(path: string): Promise<void> {
  return fetchJson<void>(`${BASE_URL}${path}`, {
    method: 'DELETE',
    headers: { 'Accept': 'application/json' },
  });
}

export function apiGetWithParams<T>(path: string, params?: Record<string, string | number | undefined>): Promise<T> {
  const url = new URL(`${BASE_URL}${path}`);
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== null) url.searchParams.set(k, String(v));
    });
  }
  return fetchJson<T>(url.toString(), {
    headers: { 'Accept': 'application/json' },
  });
}
