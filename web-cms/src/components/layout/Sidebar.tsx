import { NavLink } from 'react-router-dom';
import { ShoppingCart, X } from 'lucide-react';
import { cn } from '@/lib/cn';
import { menuForRole } from '@/routes/menu';
import { useAuth } from '@/features/auth/hooks/useAuth';

export function Sidebar({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { user } = useAuth();
  if (!user) return null;
  const items = menuForRole(user.role.code);

  return (
    <>
      {/* Backdrop for the tablet drawer */}
      <div
        className={cn('fixed inset-0 z-30 bg-black/30 lg:hidden', open ? 'block animate-fade-in' : 'hidden')}
        onClick={onClose}
        aria-hidden
      />
      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-40 flex w-sidebar flex-col border-r border-separator/60 bg-card transition-transform lg:translate-x-0',
          open ? 'translate-x-0' : '-translate-x-full',
        )}
        aria-label="Menu utama"
      >
        <div className="flex h-topbar items-center gap-3 px-5">
          <span className="flex h-9 w-9 items-center justify-center rounded-md bg-primary text-white">
            <ShoppingCart className="h-5 w-5" />
          </span>
          <div className="flex-1">
            <p className="text-headline leading-tight">POS UMKM</p>
            <p className="text-caption text-text-secondary">Dashboard Toko</p>
          </div>
          <button
            type="button"
            className="flex h-touch w-touch items-center justify-center rounded-full text-gray hover:bg-fill lg:hidden"
            onClick={onClose}
            aria-label="Tutup menu"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <nav className="flex-1 space-y-0.5 overflow-y-auto px-3 py-3">
          {items.map(({ path, label, icon: Icon }) => (
            <NavLink
              key={path}
              to={path}
              onClick={onClose}
              className={({ isActive }) =>
                cn(
                  'flex h-touch items-center gap-3 rounded-md px-3 text-subheadline font-medium transition-colors',
                  isActive ? 'bg-primary text-white' : 'text-text hover:bg-fill',
                )
              }
            >
              <Icon className="h-5 w-5 shrink-0" />
              {label}
            </NavLink>
          ))}
        </nav>
        <p className="px-5 py-4 text-caption text-text-secondary">v1.0.0</p>
      </aside>
    </>
  );
}
