import axios, { AxiosError, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios';
import { authStorage } from './auth-storage';
import type { ApiErrorBody, ApiResponse, FieldError, ListParams, Paginated } from '@/types/api';

export const API_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080/api/v1';

export const api = axios.create({ baseURL: API_URL, timeout: 30_000 });

/** Error thrown by every API helper, carrying the backend message and field errors. */
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public fieldErrors: FieldError[] = [],
    /** Raw error body for blob requests (parsed lazily by download()). */
    public blob?: Blob,
  ) {
    super(message);
  }
}

api.interceptors.request.use((config) => {
  const tokens = authStorage.get();
  if (tokens) config.headers.Authorization = `Bearer ${tokens.accessToken}`;
  return config;
});

// ---- Refresh-token handling -------------------------------------------------
// All requests that fail with 401 while a refresh is in flight wait for the same
// promise, so the refresh endpoint is called only once.
let refreshPromise: Promise<string> | null = null;
let onSessionExpired: () => void = () => {};

export function setSessionExpiredHandler(fn: () => void) {
  onSessionExpired = fn;
}

async function refreshAccessToken(): Promise<string> {
  const tokens = authStorage.get();
  if (!tokens) throw new Error('no session');
  const { data } = await axios.post<ApiResponse<{ access_token: string; refresh_token: string }>>(
    `${API_URL}/auth/refresh`,
    { refresh_token: tokens.refreshToken },
  );
  authStorage.set({ accessToken: data.data.access_token, refreshToken: data.data.refresh_token });
  return data.data.access_token;
}

type RetriableConfig = InternalAxiosRequestConfig & { _retried?: boolean };

api.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<ApiErrorBody>) => {
    const original = error.config as RetriableConfig | undefined;
    const isAuthCall = original?.url?.includes('/auth/login') || original?.url?.includes('/auth/refresh');

    if (error.response?.status === 401 && original && !original._retried && !isAuthCall && authStorage.get()) {
      original._retried = true;
      try {
        refreshPromise ??= refreshAccessToken().finally(() => {
          refreshPromise = null;
        });
        const token = await refreshPromise;
        original.headers.Authorization = `Bearer ${token}`;
        return api(original);
      } catch {
        authStorage.clear();
        onSessionExpired();
      }
    }
    return Promise.reject(toApiError(error));
  },
);

function toApiError(error: AxiosError<ApiErrorBody>): ApiError {
  if (error instanceof ApiError) return error;
  const body = error.response?.data;
  if (body instanceof Blob) {
    return new ApiError('Gagal mengunduh file', error.response!.status, [], body);
  }
  if (body && typeof body === 'object' && 'message' in body) {
    return new ApiError(body.message, error.response!.status, body.errors ?? []);
  }
  if (error.code === 'ECONNABORTED') return new ApiError('Server terlalu lama merespons. Coba lagi.', 0);
  if (!error.response) return new ApiError('Tidak dapat terhubung ke server. Periksa koneksi Anda.', 0);
  return new ApiError('Terjadi kesalahan tak terduga', error.response.status);
}

// ---- Typed helpers ----------------------------------------------------------

/** Drops empty values so they are not sent as query parameters. */
export function cleanParams(params?: ListParams): Record<string, string | number | boolean> {
  const out: Record<string, string | number | boolean> = {};
  Object.entries(params ?? {}).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== '') out[k] = v;
  });
  return out;
}

export async function get<T>(url: string, params?: ListParams): Promise<T> {
  const { data } = await api.get<ApiResponse<T>>(url, { params: cleanParams(params) });
  return data.data;
}

export async function getPage<T>(url: string, params?: ListParams): Promise<Paginated<T>> {
  const { data } = await api.get<ApiResponse<T[]>>(url, { params: cleanParams(params) });
  return { items: data.data, meta: data.meta ?? { page: 1, limit: data.data.length, total: data.data.length } };
}

export async function post<T>(url: string, body?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const { data } = await api.post<ApiResponse<T>>(url, body, config);
  return data.data;
}

export async function put<T>(url: string, body?: unknown): Promise<T> {
  const { data } = await api.put<ApiResponse<T>>(url, body);
  return data.data;
}

export async function del(url: string): Promise<void> {
  await api.delete(url);
}

export async function upload<T>(url: string, field: string, file: File): Promise<T> {
  const form = new FormData();
  form.append(field, file);
  const { data } = await api.post<ApiResponse<T>>(url, form);
  return data.data;
}

/** Downloads a binary endpoint and saves it using the server-provided filename. */
export async function download(url: string, params: ListParams, fallbackName: string): Promise<void> {
  const res = await api.get<Blob>(url, { params: cleanParams(params), responseType: 'blob' }).catch(async (err) => {
    // Error bodies of blob requests arrive as Blob and must be parsed manually.
    if (err instanceof ApiError && err.blob) {
      try {
        const body = JSON.parse(await err.blob.text()) as ApiErrorBody;
        throw new ApiError(body.message, err.status, body.errors);
      } catch (parsed) {
        if (parsed instanceof ApiError) throw parsed;
      }
    }
    throw err;
  });
  const disposition = String(res.headers['content-disposition'] ?? '');
  const name = /filename="?([^"]+)"?/.exec(disposition)?.[1] ?? fallbackName;
  const href = URL.createObjectURL(res.data);
  const a = document.createElement('a');
  a.href = href;
  a.download = name;
  a.click();
  URL.revokeObjectURL(href);
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message;
  if (err instanceof Error) return err.message;
  return 'Terjadi kesalahan';
}
