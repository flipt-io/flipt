import * as React from 'react';

import { NavMain } from '~/components/NavMain';
import { NavSecondary } from '~/components/NavSecondary';
import { NavUser } from '~/components/NavUser';
import { NamespaceSwitcher } from '~/components/namespaces/NamespaceSwitcher';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader
} from '~/components/Sidebar';
import { getUser } from '~/data/user';
import { useSession } from '~/data/hooks/session';

export function AppSidebar({
  ns,
  ...props
}: {
  ns: string;
} & React.ComponentProps<typeof Sidebar>) {
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
