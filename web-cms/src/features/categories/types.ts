export interface Category {
  id: number;
  name: string;
  description: string | null;
  product_count: number;
  created_at: string;
  updated_at: string;
}

export interface CategoryInput {
  name: string;
  description?: string | null;
}

export interface Unit {
  id: number;
  name: string;
  abbreviation: string;
  created_at: string;
  updated_at: string;
}

export interface UnitInput {
  name: string;
  abbreviation: string;
}
