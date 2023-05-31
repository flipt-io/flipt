import { useState } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import Footer from '~/components/Footer';
import Header from '~/components/Header';
import NamespaceProvider from '~/components/NamespaceProvider';
import { NotificationProvider } from '~/components/NotificationProvider';
import ErrorNotification from '~/components/notifications/ErrorNotification';
import SuccessNotification from '~/components/notifications/SuccessNotification';
import PreferencesProvider from '~/components/PreferencesProvider';
import Sidebar from '~/components/Sidebar';
import { useSession } from '~/data/hooks/session';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  if (!session) {
    return <Navigate to="/login" />;
  }

  return (
    <>
      <Sidebar setSidebarOpen={setSidebarOpen} sidebarOpen={sidebarOpen} />
      <div className="flex min-h-screen flex-col bg-white md:pl-64">
        <Header setSidebarOpen={setSidebarOpen} />

        <main className="flex px-6 py-10">
          <div className="w-full overflow-x-auto px-4 sm:px-6 lg:px-8">
            <Outlet />
          </div>
        </main>
        <Footer />
      </div>
    </>
  );
}

export default function Layout() {
  return (
    <NotificationProvider>
      <PreferencesProvider>
        <NamespaceProvider>
          <InnerLayout />
        </NamespaceProvider>
      </PreferencesProvider>
      <ErrorNotification />
      <SuccessNotification />
    </NotificationProvider>
  );
}
