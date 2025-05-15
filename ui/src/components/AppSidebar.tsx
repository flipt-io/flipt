import * as React from 'react';

import { NavMain } from '~/components/NavMain';
import { NavUser } from '~/components/NavUser';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader
} from '~/components/ui/sidebar';

import { useSession } from '~/data/hooks/session';
import { getUser } from '~/data/user';

import { NamespaceSwitcher } from './NamespaceSwitcher';
import { NavSecondary } from './NavSecondary';

export function AppSidebar({
  ns,
  ...props
}: { ns: string } & React.ComponentProps<typeof Sidebar>) {
  const { session } = useSession();
  const user = getUser(session);
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <NamespaceSwitcher />
      </SidebarHeader>
      <SidebarContent>
        <NavMain ns={ns} />
        <NavSecondary className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>{user && <NavUser user={user} />}</SidebarFooter>
    </Sidebar>
  );
}
