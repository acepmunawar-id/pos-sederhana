import { lazy, Suspense, type ReactNode } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { AppLayout } from '@/components/layout/AppLayout';
import { EmptyState } from '@/components/ui';
import { useAuth } from '@/features/auth/hooks/useAuth';
import type { RoleCode } from '@/types/api';
import { defaultPathForRole, menu } from './menu';

const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage'));
const DashboardPage = lazy(() => import('@/features/dashboard/pages/DashboardPage'));
const ProductsPage = lazy(() => import('@/features/products/pages/ProductsPage'));
const CategoriesPage = lazy(() => import('@/features/categories/pages/CategoriesPage'));
const StockPage = lazy(() => import('@/features/stock/pages/StockPage'));
const TransactionsPage = lazy(() => import('@/features/transactions/pages/TransactionsPage'));
const TransactionDetailPage = lazy(() => import('@/features/transactions/pages/TransactionDetailPage'));
const ShiftsPage = lazy(() => import('@/features/transactions/pages/ShiftsPage'));
const ReportsPage = lazy(() => import('@/features/reports/pages/ReportsPage'));
const UsersPage = lazy(() => import('@/features/users/pages/UsersPage'));
const SettingsPage = lazy(() => import('@/features/settings/pages/SettingsPage'));

function FullPageSpinner() {
  return (
    <div className="flex min-h-[50vh] items-center justify-center text-gray">
      <Loader2 className="h-6 w-6 animate-spin" aria-label="Memuat" />
    </div>
  );
}

function RequireAuth({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  const location = useLocation();
  if (status === 'loading') return <FullPageSpinner />;
  if (status === 'guest') return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  return <>{children}</>;
}

/** Guards a route with the roles declared for it in the menu. */
function RequireRole({ path, children }: { path: string; children: ReactNode }) {
  const { user } = useAuth();
  const roles: RoleCode[] = menu.find((m) => m.path === path)?.roles ?? [];
  if (!user || !roles.includes(user.role.code)) {
    return (
      <EmptyState
        title="Akses ditolak"
        description="Akun Anda tidak memiliki izin untuk membuka halaman ini."
      />
    );
  }
  return <>{children}</>;
}

function HomeRedirect() {
  const { user } = useAuth();
  return <Navigate to={user ? defaultPathForRole(user.role.code) : '/login'} replace />;
}

const guarded = (path: string, el: ReactNode) => <RequireRole path={path}>{el}</RequireRole>;

export function AppRoutes() {
  return (
    <BrowserRouter>
      <Suspense fallback={<FullPageSpinner />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            element={
              <RequireAuth>
                <AppLayout />
              </RequireAuth>
            }
          >
            <Route index element={<HomeRedirect />} />
            <Route path="/dashboard" element={guarded('/dashboard', <DashboardPage />)} />
            <Route path="/products" element={guarded('/products', <ProductsPage />)} />
            <Route path="/categories" element={guarded('/categories', <CategoriesPage />)} />
            <Route path="/stock" element={guarded('/stock', <StockPage />)} />
            <Route path="/transactions" element={guarded('/transactions', <TransactionsPage />)} />
            <Route path="/transactions/:id" element={guarded('/transactions', <TransactionDetailPage />)} />
            <Route path="/shifts" element={guarded('/shifts', <ShiftsPage />)} />
            <Route path="/reports" element={guarded('/reports', <ReportsPage />)} />
            <Route path="/users" element={guarded('/users', <UsersPage />)} />
            <Route path="/settings" element={guarded('/settings', <SettingsPage />)} />
            <Route path="*" element={<EmptyState title="Halaman tidak ditemukan" description="Periksa kembali alamat yang Anda buka." />} />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
