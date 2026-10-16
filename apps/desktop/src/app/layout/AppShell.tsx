import { useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import PageContainer from './PageContainer';
import Sidebar from './Sidebar';
import Topbar from './Topbar';

const STORAGE_KEY = 'financeapp:sidebar-collapsed';

function readCollapsed(): boolean {
  try {
    return window.localStorage.getItem(STORAGE_KEY) === '1';
  } catch {
    return false;
  }
}

function persistCollapsed(collapsed: boolean): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, collapsed ? '1' : '0');
  } catch {
    /* persistence is best-effort */
  }
}

export default function AppShell() {
  const [collapsed, setCollapsed] = useState<boolean>(readCollapsed);

  const toggleSidebar = () => {
    setCollapsed((prev) => {
      const next = !prev;
      persistCollapsed(next);
      return next;
    });
  };

  // Auto-collapse below 1280px (desktop minimum layout). Manual expand stays possible.
  useEffect(() => {
    const mq = window.matchMedia('(min-width: 1280px)');
    const handleChange = () => {
      if (!mq.matches) setCollapsed(true);
    };
    handleChange();
    mq.addEventListener('change', handleChange);
    return () => mq.removeEventListener('change', handleChange);
  }, []);

  return (
    <div className="flex h-screen w-screen flex-col bg-bg text-ink">
      <Topbar onToggleSidebar={toggleSidebar} />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar collapsed={collapsed} />
        <PageContainer>
          <Outlet />
        </PageContainer>
      </div>
    </div>
  );
}
