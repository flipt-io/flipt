import nightwind from 'nightwind/helper';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useMediaQuery } from 'react-responsive';
import { Navigate, Outlet } from 'react-router-dom';
import Footer from '~/components/Footer';
import Header from '~/components/Header';
import { NotificationProvider } from '~/components/NotificationProvider';
import ErrorNotification from '~/components/notifications/ErrorNotification';
import SuccessNotification from '~/components/notifications/SuccessNotification';
import Sidebar from '~/components/Sidebar';
import { useSession } from '~/data/hooks/session';
import { useAppDispatch } from '~/data/hooks/store';
import { Theme } from '~/types/Preferences';
import { fetchNamespacesAsync } from './namespaces/namespacesSlice';
import { selectTheme } from './preferences/preferencesSlice';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const dispatch = useAppDispatch();
  dispatch(fetchNamespacesAsync());

  const theme = useSelector(selectTheme);

  const useColorScheme = () => {
    useMediaQuery(
      {
        query: '(prefers-color-scheme: dark)'
      },
      undefined,
      (prefersDark) => {
        if (theme === Theme.SYSTEM) {
          nightwind.enable(prefersDark);
        }
      }
    );
  };

  useColorScheme();

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
      <InnerLayout />
      <ErrorNotification />
      <SuccessNotification />
    </NotificationProvider>
  );
}
