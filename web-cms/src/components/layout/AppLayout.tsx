import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Topbar } from './Topbar';

export function AppLayout() {
  const [drawer, setDrawer] = useState(false);
  return (
    <div className="min-h-screen">
      <Sidebar open={drawer} onClose={() => setDrawer(false)} />
      <div className="lg:pl-sidebar">
        <Topbar onMenu={() => setDrawer(true)} />
        <main className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:py-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
