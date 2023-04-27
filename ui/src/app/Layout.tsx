import { useEffect, useState } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import ErrorNotification from '~/components/ErrorNotification';
import Footer from '~/components/Footer';
import Header from '~/components/Header';
import NamespaceProvider from '~/components/NamespaceProvider';
import { NotificationProvider } from '~/components/NotificationProvider';
import Sidebar from '~/components/Sidebar';
import SuccessNotification from '~/components/SuccessNotification';
import { listNamespaces } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { INamespace } from '~/types/Namespace';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // TODO: replace with a proper state management solution like Redux
  // instead of passing around via context
  const [namespaces, setNamespaces] = useState<INamespace[]>([]);

  const { setError } = useError();

  useEffect(() => {
    if (!session) return;

    listNamespaces()
      .then((data) => {
        setNamespaces(data.namespaces);
      })
      .catch((err) => {
        setError(err);
      });
  }, [session, setError]);

  if (!session) {
    return <Navigate to="/login" />;
  }

  return (
    <>
      <Sidebar
        namespaces={namespaces}
        setSidebarOpen={setSidebarOpen}
        sidebarOpen={sidebarOpen}
      />
      <div className="flex min-h-screen flex-col bg-white md:pl-64">
        <Header setSidebarOpen={setSidebarOpen} />

        <main className="flex px-6 py-10">
          <div className="w-full overflow-x-auto px-4 sm:px-6 lg:px-8">
            <Outlet context={{ namespaces, setNamespaces }} />
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
      <NamespaceProvider>
        <InnerLayout />
      </NamespaceProvider>
      <ErrorNotification />
      <SuccessNotification />
    </NotificationProvider>
  );
}
