import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import {
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
  useParams
} from 'react-router';
import Footer from '~/components/Footer';
import Header from '~/components/Header';
import Loading from '~/components/Loading';
import { NotificationProvider } from '~/components/NotificationProvider';
import ErrorNotification from '~/components/notifications/ErrorNotification';
import SuccessNotification from '~/components/notifications/SuccessNotification';
import Sidebar from '~/components/Sidebar';
import { useSession } from '~/data/hooks/session';
import { useAppDispatch } from '~/data/hooks/store';
import { LoadingStatus } from '~/types/Meta';
import { fetchInfoAsync, selectConfig } from './meta/metaSlice';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  useListNamespacesQuery
} from './namespaces/namespacesSlice';
import CommandDialog from '~/components/command/CommandDialog';
import Banner from '~/components/Banner';
import { selectDismissedBanner } from './events/eventSlice';
import { StarIcon } from '@heroicons/react/20/solid';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const dismissedBanner = useSelector(selectDismissedBanner);
  const dispatch = useAppDispatch();

  const { namespaceKey } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const currentNamespace = useSelector(selectCurrentNamespace);
  const config = useSelector(selectConfig);

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

  const namespaces = useListNamespacesQuery();

  useEffect(() => {
    dispatch(fetchInfoAsync());
  }, [dispatch]);

  if (!session) {
    return <Navigate to="/login" />;
  }

  if (namespaces.isLoading || config.status != LoadingStatus.SUCCEEDED) {
    return <Loading fullScreen />;
  }

  return (
    <>
      <Sidebar setSidebarOpen={setSidebarOpen} sidebarOpen={sidebarOpen} />
      <div className="flex min-h-screen flex-col bg-background md:pl-64">
        <Header setSidebarOpen={setSidebarOpen} />
        {!dismissedBanner && (
          <Banner
            title="Like Flipt? Give us a star on GitHub!"
            description="It really means a lot to us. Thank you!"
            href="https://github.com/flipt-io/flipt"
            icon={<StarIcon className="mx-2 inline h-3 w-3" />}
          />
        )}
        <main className="flex pt-1 sm:pt-4">
          <div className="mx-auto w-full max-w-(--breakpoint-lg) overflow-x-auto px-4 sm:px-6 lg:px-8">
            <Outlet />
          </div>
        </main>
        <Footer />
        <CommandDialog />
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
