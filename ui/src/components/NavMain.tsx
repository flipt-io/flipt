import {
  BugPlayIcon,
  ChartNoAxesCombinedIcon,
  FlagIcon,
  UsersIcon
} from 'lucide-react';

import {
  SidebarGroup,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem
} from '~/components/ui/sidebar';

export function NavMain() {
  const items = [
    {
      title: 'Flags',
      url: '#/namespaces/default/flags',
      icon: FlagIcon
    },
    {
      title: 'Segments',
      url: '#/namespaces/default/segments',
      icon: UsersIcon
    },
    {
      title: 'Playground',
      url: '#/namespaces/default/playground',
      icon: BugPlayIcon
    },
    {
      title: 'Analytics',
      url: '#',
      icon: ChartNoAxesCombinedIcon
    }
  ];
  return (
    <SidebarGroup>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            <SidebarMenuButton tooltip={item.title} asChild>
              <a href={item.url}>
                {item.icon && <item.icon />}
                <span>{item.title}</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
