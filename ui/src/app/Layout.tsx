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
import { Loading } from '~/components/Loading';
import { SidebarInset, SidebarProvider } from '~/components/Sidebar';
import { Toaster } from '~/components/Sonner';
import { TooltipProvider } from '~/components/Tooltip';
import CommandDialog from '~/components/command/CommandDialog';

import { useSession } from '~/data/hooks/session';
import { useAppDispatch } from '~/data/hooks/store';
import { LoadingStatus } from '~/types/Meta';
import { fetchInfoAsync, selectConfig } from './meta/metaSlice';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  useListNamespacesQuery
} from './namespaces/namespacesSlice';
import Banner from '~/components/Banner';
import {
  selectDismissedV2Banner,
  v2BannerDismissed
} from './events/eventSlice';
import { selectSidebar, sidebarChanged } from './preferences/preferencesSlice';

function InnerLayout() {
  const { session } = useSession();

  const dismissedV2Banner = useSelector(selectDismissedV2Banner);
  const bannerEnabled = import.meta.env.FLIPT_UI_BANNER_ENABLED !== 'false';
  const dispatch = useAppDispatch();
  const sidebarOpen = useSelector(selectSidebar);
  const setSidebarOpen = (open: boolean) => {
    dispatch(sidebarChanged(open));
  };

  const { namespaceKey } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const currentNamespace = useSelector(selectCurrentNamespace);
  const config = useSelector(selectConfig);

  useEffect(() => {
    if (!namespaceKey) {
      return;
    }

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
    return <Loading variant="fullscreen" />;
  }

  const ns = currentNamespace?.key || 'default';

  return (
    <TooltipProvider>
      <SidebarProvider open={sidebarOpen} onOpenChange={setSidebarOpen}>
        <AppSidebar variant="inset" ns={ns} />
        <SidebarInset>
          <Header ns={ns} />
          {!dismissedV2Banner && bannerEnabled && (
            <Banner
              emoji="🎉"
              title="Flipt v2 is now available!"
              description="Git-native feature flags with multi-environment support, branching, and real-time updates"
              href="https://docs.flipt.io/v2/introduction"
              onDismiss={() => dispatch(v2BannerDismissed())}
            />
          )}
          <main className="flex pt-1 sm:pt-4">
            <div className="mx-auto w-full max-w-6xl overflow-x-auto px-4 sm:px-6 lg:px-8">
              <Outlet />
            </div>
          </main>
          <Footer />
          <CommandDialog />
          <Toaster />
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}

export default function Layout() {
  return <InnerLayout />;
}
