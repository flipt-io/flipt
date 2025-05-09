import { StarIcon } from 'lucide-react';
import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import {
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
  useParams
} from 'react-router';

import { AppSidebar } from '~/components/AppSidebar';
import Banner from '~/components/Banner';
import Footer from '~/components/Footer';
import { Header } from '~/components/Header';
import Loading from '~/components/Loading';
import { Toaster } from '~/components/Sonner';
import CommandDialog from '~/components/command/CommandDialog';
import { SidebarInset, SidebarProvider } from '~/components/ui/sidebar';

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
    <SidebarProvider>
      <AppSidebar variant="inset" />
      <SidebarInset>
        <Header ns={currentEnvironment.key + '/' + currentNamespace.name} />
        <div className="min-h-[100vh] flex-1 md:min-h-min">
          <div className="sticky top-0 z-10">
            {!dismissedBanner && (
              <div className="z-10">
                <Banner
                  title="Like Flipt? Give us a star on GitHub!"
                  description="It really means a lot to us. Thank you!"
                  href="https://github.com/flipt-io/flipt"
                  icon={<StarIcon className="mx-2 mb-1 inline h-4 w-4" />}
                />
              </div>
            )}
          </div>
          <main className="flex flex-1 relative pt-8">
            <div className="mx-auto w-full lg:max-w-(--breakpoint-lg) xl:max-w-(--breakpoint-xl) 2xl:max-w-(--breakpoint-2xl) overflow-x-auto px-4 sm:px-6 lg:px-8">
              <Outlet />
            </div>
          </main>
          <Footer />
          <CommandDialog />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default function Layout() {
  return (
    <>
      <InnerLayout />
      <Toaster />
    </>
  );
}
