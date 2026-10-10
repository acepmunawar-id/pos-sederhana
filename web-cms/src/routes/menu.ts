import type { LucideIcon } from 'lucide-react';
import {
  BarChart3,
  Boxes,
  LayoutDashboard,
  Package,
  Receipt,
  Settings,
  Tags,
  Timer,
  Users,
} from 'lucide-react';
import type { RoleCode } from '@/types/api';

export interface MenuItem {
  path: string;
  label: string;
  icon: LucideIcon;
  roles: RoleCode[];
}

const ALL: RoleCode[] = ['owner', 'admin', 'cashier'];
const MANAGERS: RoleCode[] = ['owner', 'admin'];

/** Single source of truth for navigation and route guards. */
export const menu: MenuItem[] = [
  { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, roles: MANAGERS },
  { path: '/products', label: 'Produk', icon: Package, roles: ALL },
  { path: '/categories', label: 'Kategori & Satuan', icon: Tags, roles: MANAGERS },
  { path: '/stock', label: 'Stok', icon: Boxes, roles: ALL },
  { path: '/transactions', label: 'Transaksi', icon: Receipt, roles: ALL },
  { path: '/shifts', label: 'Shift Kasir', icon: Timer, roles: ALL },
  { path: '/reports', label: 'Laporan', icon: BarChart3, roles: MANAGERS },
  { path: '/users', label: 'Pengguna', icon: Users, roles: MANAGERS },
  { path: '/settings', label: 'Pengaturan Toko', icon: Settings, roles: ['owner'] },
];

export function menuForRole(role: RoleCode): MenuItem[] {
  return menu.filter((m) => m.roles.includes(role));
}

export function defaultPathForRole(role: RoleCode): string {
  return menuForRole(role)[0]?.path ?? '/products';
}
