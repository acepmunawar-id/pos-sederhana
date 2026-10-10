import { useEffect, useRef, useState } from 'react';
import { ChevronDown, KeyRound, LogOut, Menu } from 'lucide-react';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { ChangePasswordModal } from '@/features/auth/components/ChangePasswordModal';
import { ConfirmDialog } from '@/components/ui';
import { roleLabel } from '@/lib/format';

export function Topbar({ onMenu }: { onMenu: () => void }) {
  const { user, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [pwOpen, setPwOpen] = useState(false);
  const [confirmLogout, setConfirmLogout] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const onClick = (e: MouseEvent) => !ref.current?.contains(e.target as Node) && setMenuOpen(false);
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, [menuOpen]);

  if (!user) return null;
  const initials = user.name
    .split(' ')
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();

  return (
    <header className="sticky top-0 z-20 flex h-topbar items-center justify-between gap-3 border-b border-separator/60 bg-card/80 px-4 backdrop-blur-xl sm:px-6">
      <button
        type="button"
        onClick={onMenu}
        className="flex h-touch items-center gap-2 rounded-md px-2 text-primary hover:bg-fill lg:hidden"
        aria-label="Buka menu"
      >
        <Menu className="h-5 w-5" />
        <span className="text-subheadline font-medium">Menu</span>
      </button>
      <div className="hidden lg:block" />

      <div className="relative" ref={ref}>
        <button
          type="button"
          onClick={() => setMenuOpen((v) => !v)}
          className="flex h-touch items-center gap-3 rounded-md px-2 hover:bg-fill"
          aria-haspopup="menu"
          aria-expanded={menuOpen}
        >
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10 text-footnote font-bold text-primary">
            {initials}
          </span>
          <span className="hidden text-left sm:block">
            <span className="block text-subheadline font-semibold leading-tight">{user.name}</span>
            <span className="block text-caption text-text-secondary">{roleLabel[user.role.code]}</span>
          </span>
          <ChevronDown className="h-4 w-4 text-gray" />
        </button>
        {menuOpen && (
          <div role="menu" className="absolute right-0 mt-2 w-56 animate-fade-in overflow-hidden rounded-lg bg-card shadow-modal">
            <div className="border-b border-separator/60 px-4 py-3">
              <p className="truncate text-subheadline font-semibold">{user.name}</p>
              <p className="truncate text-footnote text-text-secondary">{user.email}</p>
            </div>
            <button
              role="menuitem"
              type="button"
              className="flex h-touch w-full items-center gap-3 px-4 text-subheadline hover:bg-fill"
              onClick={() => {
                setMenuOpen(false);
                setPwOpen(true);
              }}
            >
              <KeyRound className="h-4 w-4 text-gray" /> Ganti Password
            </button>
            <button
              role="menuitem"
              type="button"
              className="flex h-touch w-full items-center gap-3 px-4 text-subheadline text-danger hover:bg-fill"
              onClick={() => {
                setMenuOpen(false);
                setConfirmLogout(true);
              }}
            >
              <LogOut className="h-4 w-4" /> Keluar
            </button>
          </div>
        )}
      </div>

      <ChangePasswordModal open={pwOpen} onClose={() => setPwOpen(false)} />
      <ConfirmDialog
        open={confirmLogout}
        title="Keluar dari akun?"
        message="Anda perlu login kembali untuk mengakses dashboard."
        confirmLabel="Keluar"
        loading={loggingOut}
        onClose={() => setConfirmLogout(false)}
        onConfirm={async () => {
          setLoggingOut(true);
          await logout();
        }}
      />
    </header>
  );
}
