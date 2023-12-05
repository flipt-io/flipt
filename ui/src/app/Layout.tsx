import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import {
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
  useParams
} from 'react-router-dom';
import Footer from '~/components/Footer';
import Header from '~/components/Header';
import { NotificationProvider } from '~/components/NotificationProvider';
import ErrorNotification from '~/components/notifications/ErrorNotification';
import SuccessNotification from '~/components/notifications/SuccessNotification';
import Sidebar from '~/components/Sidebar';
import { useSession } from '~/data/hooks/session';
import { useAppDispatch } from '~/data/hooks/store';
import { fetchConfigAsync, fetchInfoAsync } from './meta/metaSlice';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  useListNamespacesQuery
} from './namespaces/namespacesSlice';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const dispatch = useAppDispatch();

  const { namespaceKey } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const currentNamespace = useSelector(selectCurrentNamespace);

  useEffect(() => {
    if (!namespaceKey) {
      return;
    }

    // if the namespaceKey in the url is not the same as the currentNamespace, then
    // dispatch the currentNamespaceChanged action to update the currentNamespace in the store
    // this allows the namespace to be changed by the url and not just the namespace dropdown,
    // which is required for 'deep' linking
    if (currentNamespace?.key !== namespaceKey) {
      dispatch(currentNamespaceChanged({ key: namespaceKey }));
    }
  }, [namespaceKey, currentNamespace, dispatch, navigate, location.pathname]);

  useListNamespacesQuery();

  useEffect(() => {
    if(currentNamespace)
    {  
      dispatch(fetchInfoAsync());
      dispatch(fetchConfigAsync());
    }
  }, [currentNamespace, dispatch]);

  if (!session) {
    return <Navigate to="/login" />;
  }

  return (
    <>
      <Sidebar setSidebarOpen={setSidebarOpen} sidebarOpen={sidebarOpen} />
      <div className="bg-white flex min-h-screen flex-col md:pl-64">
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
