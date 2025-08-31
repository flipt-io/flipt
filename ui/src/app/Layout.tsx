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
import Footer from '~/components/Footer';
import { Header } from '~/components/Header';
import Loading from '~/components/Loading';
import ProBanner from '~/components/ProBanner';
import { Toaster } from '~/components/Sonner';
import CommandDialog from '~/components/command/CommandDialog';
import { SidebarInset, SidebarProvider } from '~/components/Sidebar';

import { LoadingStatus, Product } from '~/types/Meta';

import { useSession } from '~/data/hooks/session';
import { useAppDispatch } from '~/data/hooks/store';

import {
  selectCurrentEnvironment,
  useListEnvironmentsQuery
} from './environments/environmentsApi';
import { selectDismissedProBanner } from './events/eventSlice';
import { fetchInfoAsync, selectInfo } from './meta/metaSlice';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  useListNamespacesQuery
} from './namespaces/namespacesApi';
import { selectSidebar, sidebarChanged } from './preferences/preferencesSlice';

function InnerLayout() {
  const { session } = useSession();

  const dismissedProBanner = useSelector(selectDismissedProBanner);
  const dispatch = useAppDispatch();

  const { namespaceKey } = useParams();
  const location = useLocation();
  const navigate = useNavigate();

  const environments = useListEnvironmentsQuery();

  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const currentNamespace = useSelector(selectCurrentNamespace);
  const info = useSelector(selectInfo);

  const namespaces = useListNamespacesQuery(
    {
      environmentKey: currentEnvironment.key
    },
    { skip: environments.isLoading }
  );

  const sidebarOpen = useSelector(selectSidebar);
  const setSidebarOpen = () => {
    dispatch(sidebarChanged(!sidebarOpen));
  };

  useEffect(() => {
    if (!namespaceKey) {
      return;
    }

    // if the namespaceKey in the url is not the same as the currentNamespace, then
    // dispatch the currentNamespaceChanged action to update the currentNamespace in the store
    // this allows the namespace to be changed by the url and not just the namespace dropdown,
    // which is required for 'deep' linking
    if (currentNamespace?.key !== namespaceKey) {
      dispatch(currentNamespaceChanged(namespaceKey));
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
    <SidebarProvider open={sidebarOpen} onOpenChange={setSidebarOpen}>
      <AppSidebar variant="inset" ns={currentNamespace.key} />
      <SidebarInset>
        <div className="min-h-screen flex flex-col">
          <Header
            ns={currentNamespace.key}
            env={currentEnvironment.key}
            sidebarOpen={sidebarOpen}
          />
          <div className="sticky top-0 z-10">
            {!dismissedProBanner && info.product !== Product.PRO && (
              <div className="z-10">
                <ProBanner />
              </div>
            )}
          </div>
          <main className="flex-1 relative pt-8 flex">
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
