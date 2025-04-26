import { StarIcon } from '@heroicons/react/20/solid';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import {
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
  useParams
} from 'react-router';

import Banner from '~/components/Banner';
import Footer from '~/components/Footer';
import Header from '~/components/Header';
import Loading from '~/components/Loading';
import { NotificationProvider } from '~/components/NotificationProvider';
import Sidebar from '~/components/Sidebar';
import CommandDialog from '~/components/command/CommandDialog';
import ErrorNotification from '~/components/notifications/ErrorNotification';
import SuccessNotification from '~/components/notifications/SuccessNotification';

import { LoadingStatus } from '~/types/Meta';

import { useSession } from '~/data/hooks/session';
import { useAppDispatch } from '~/data/hooks/store';

import {
  selectCurrentEnvironment,
  useListEnvironmentsQuery
} from './environments/environmentsApi';
import { selectDismissedBanner } from './events/eventSlice';
import { fetchInfoAsync, selectInfo } from './meta/metaSlice';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  useListNamespacesQuery
} from './namespaces/namespacesApi';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const dismissedBanner = useSelector(selectDismissedBanner);
  const dispatch = useAppDispatch();

  const { namespaceKey } = useParams();
  const location = useLocation();
  const navigate = useNavigate();

  const environments = useListEnvironmentsQuery();

  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const currentNamespace = useSelector(selectCurrentNamespace);
  const info = useSelector(selectInfo);

  const namespaces = useListNamespacesQuery({
    environmentKey: currentEnvironment.key
  });

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

  useEffect(() => {
    dispatch(fetchInfoAsync());
  }, [dispatch]);

  if (!session) {
    return <Navigate to="/login" />;
  }

  if (
    environments.isLoading ||
    namespaces.isLoading ||
    info.status != LoadingStatus.SUCCEEDED
  ) {
    return <Loading fullScreen />;
  }

  return (
    <>
      <Sidebar setSidebarOpen={setSidebarOpen} sidebarOpen={sidebarOpen} />
      <div className="flex min-h-screen flex-col bg-background md:pl-64">
        <div className="sticky top-0 z-10">
          <Header setSidebarOpen={setSidebarOpen} />
          {!dismissedBanner && (
            <div className="mt-16 z-10">
              <Banner
                title="Like Flipt? Give us a star on GitHub!"
                description="It really means a lot to us. Thank you!"
                href="https://github.com/flipt-io/flipt"
                icon={<StarIcon className="mx-2 inline h-3 w-3" />}
              />
            </div>
          )}
        </div>
        <main
          className={`flex flex-1 relative ${!dismissedBanner ? 'pt-8' : 'pt-24'}`}
        >
          <div className="mx-auto w-full lg:max-w-(--breakpoint-lg) xl:max-w-(--breakpoint-xl) 2xl:max-w-(--breakpoint-2xl) overflow-x-auto px-4 sm:px-6 lg:px-8">
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
