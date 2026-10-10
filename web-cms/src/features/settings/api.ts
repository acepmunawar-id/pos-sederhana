import { get, put, upload } from '@/lib/api';
import type { StoreSetting, StoreSettingInput } from './types';

export const settingApi = {
  get: () => get<StoreSetting>('/settings'),
  update: (input: StoreSettingInput) => put<StoreSetting>('/settings', input),
  uploadLogo: (file: File) => upload<StoreSetting>('/settings/logo', 'logo', file),
};
