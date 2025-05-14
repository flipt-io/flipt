import * as React from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';

import { NavMain } from '~/components/NavMain';
import { NavUser } from '~/components/NavUser';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  useSidebar
} from '~/components/ui/sidebar';

import { useSession } from '~/data/hooks/session';
import { getUser } from '~/data/user';

import { EnvironmentNamespaceSwitcher } from './EnvironmentNamespaceSwitcher';
import { NavSecondary } from './NavSecondary';
import { EnvironmentRemoteInfo } from './environments/EnvironmentRemoteInfo';

export function AppSidebar({
  ns,
  ...props
}: { ns: string } & React.ComponentProps<typeof Sidebar>) {
  const { session } = useSession();
  const user = getUser(session);

  const { isMobile, state } = useSidebar();
  const currentEnvironment = useSelector(selectCurrentEnvironment);

  return (
    <Sidebar collapsible="icon" {...props} isMobile={isMobile}>
      <SidebarHeader>
        <EnvironmentNamespaceSwitcher isMobile={isMobile} />
        {state === 'expanded' && (
          <EnvironmentRemoteInfo environment={currentEnvironment} />
        )}
      </SidebarHeader>
      <SidebarContent>
        <NavMain ns={ns} />
        <NavSecondary isMobile={isMobile} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>{user && <NavUser user={user} />}</SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
