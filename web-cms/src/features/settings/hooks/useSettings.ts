import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { settingApi } from '../api';
import type { StoreSettingInput } from '../types';

export function useSettings() {
  return useQuery({ queryKey: ['settings'], queryFn: settingApi.get });
}

export function useUpdateSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: StoreSettingInput) => settingApi.update(input),
    onSuccess: (data) => qc.setQueryData(['settings'], data),
  });
}

export function useUploadLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => settingApi.uploadLogo(file),
    onSuccess: (data) => qc.setQueryData(['settings'], data),
  });
}
