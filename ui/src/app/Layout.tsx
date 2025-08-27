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
import {
  selectDismissedV2Banner,
  v2BannerDismissed
} from './events/eventSlice';

function InnerLayout() {
  const { session } = useSession();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const dismissedV2Banner = useSelector(selectDismissedV2Banner);
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
      <div className="bg-background flex min-h-screen flex-col md:pl-64">
        <Header setSidebarOpen={setSidebarOpen} />
        {!dismissedV2Banner && (
          <Banner
            emoji="🎉"
            title="Flipt v2.0.0 is now available!"
            description="Git-native feature flags with multi-environment support, branching, and real-time updates"
            href="https://docs.flipt.io/v2/introduction"
            onDismiss={() => dispatch(v2BannerDismissed())}
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
