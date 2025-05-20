import * as React from 'react';

import { NavMain } from '~/components/NavMain';
import { NavUser } from '~/components/NavUser';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader
} from '~/components/ui/sidebar';

import { ISelectable } from '~/types/Selectable';

import { useSession } from '~/data/hooks/session';
import { getUser } from '~/data/user';

import { EnvironmentBranchSelector } from './EnvironmentBranchSelector';
import { EnvironmentNamespaceSwitcher } from './EnvironmentNamespaceSwitcher';
import { NavSecondary } from './NavSecondary';

export function AppSidebar({
  ns,
  ...props
}: { ns: string } & React.ComponentProps<typeof Sidebar>) {
  const { session } = useSession();
  const user = getUser(session);
  const [selectedBranch, setSelectedBranch] =
    React.useState<ISelectable | null>(null);

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <EnvironmentNamespaceSwitcher />
        <EnvironmentBranchSelector
          selectedBranch={selectedBranch}
          setSelectedBranch={setSelectedBranch}
          className="mt-2"
        />
      </SidebarHeader>
      <SidebarContent>
        <NavMain ns={ns} />
        <NavSecondary className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>{user && <NavUser user={user} />}</SidebarFooter>
    </Sidebar>
  );
}
