import { del, get, getPage, post, put, upload } from '@/lib/api';
import type { ListParams } from '@/types/api';
import type { CreateProductInput, Product, ProductInput } from './types';

export const productApi = {
  list: (params: ListParams) => getPage<Product>('/products', params),
  get: (id: number) => get<Product>(`/products/${id}`),
  create: (input: CreateProductInput) => post<Product>('/products', input),
  update: (id: number, input: ProductInput) => put<Product>(`/products/${id}`, input),
  remove: (id: number) => del(`/products/${id}`),
  uploadImage: (id: number, file: File) => upload<Product>(`/products/${id}/image`, 'image', file),
};
