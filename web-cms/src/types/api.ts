/** Uniform envelope returned by every backend endpoint. */
export interface ApiResponse<T> {
  success: true;
  message: string;
  data: T;
  meta?: PageMeta;
}

export interface PageMeta {
  page: number;
  limit: number;
  total: number;
}

export interface Paginated<T> {
  items: T[];
  meta: PageMeta;
}

export interface FieldError {
  field: string;
  message: string;
}

export interface ApiErrorBody {
  success: false;
  message: string;
  errors: FieldError[];
}

export interface ListParams {
  page?: number;
  limit?: number;
  search?: string;
  sort?: string;
  order?: 'asc' | 'desc';
  [filter: string]: string | number | boolean | undefined;
}

export type RoleCode = 'owner' | 'admin' | 'cashier';

export interface IdName {
  id: number;
  name: string;
}
