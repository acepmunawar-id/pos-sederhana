export interface StoreSetting {
  name: string;
  address: string | null;
  phone: string | null;
  logo_url: string | null;
  tax_rate_bp: number;
  receipt_footer: string | null;
  updated_at: string;
}

export interface StoreSettingInput {
  name: string;
  address: string | null;
  phone: string | null;
  tax_rate_bp: number;
  receipt_footer: string | null;
}
