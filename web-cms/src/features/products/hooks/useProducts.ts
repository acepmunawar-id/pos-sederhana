import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListParams } from '@/types/api';
import { productApi } from '../api';
import type { CreateProductInput, ProductInput } from '../types';

export const productKeys = {
  all: ['products'] as const,
  list: (p: ListParams) => ['products', 'list', p] as const,
};

export function useProducts(params: ListParams) {
  return useQuery({ queryKey: productKeys.list(params), queryFn: () => productApi.list(params), placeholderData: keepPreviousData });
}

/** Invalidates everything that shows product stock or names. */
export function useInvalidateProducts() {
  const qc = useQueryClient();
  return () =>
    Promise.all([
      qc.invalidateQueries({ queryKey: productKeys.all }),
      qc.invalidateQueries({ queryKey: ['categories'] }),
      qc.invalidateQueries({ queryKey: ['stock'] }),
      qc.invalidateQueries({ queryKey: ['dashboard'] }),
    ]);
}

export function useCreateProduct() {
  const invalidate = useInvalidateProducts();
  return useMutation({ mutationFn: (input: CreateProductInput) => productApi.create(input), onSuccess: invalidate });
}

export function useUpdateProduct() {
  const invalidate = useInvalidateProducts();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: ProductInput }) => productApi.update(id, input),
    onSuccess: invalidate,
  });
}

export function useDeleteProduct() {
  const invalidate = useInvalidateProducts();
  return useMutation({ mutationFn: (id: number) => productApi.remove(id), onSuccess: invalidate });
}

export function useUploadProductImage() {
  const invalidate = useInvalidateProducts();
  return useMutation({
    mutationFn: ({ id, file }: { id: number; file: File }) => productApi.uploadImage(id, file),
    onSuccess: invalidate,
  });
}
