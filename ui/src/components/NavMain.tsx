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

export function NavMain({ ns }: { ns: string }) {
  const items = [
    {
      title: 'Flags',
      url: `#/namespaces/${ns}/flags`,
      icon: FlagIcon
    },
    {
      title: 'Segments',
      url: `#/namespaces/${ns}/segments`,
      icon: UsersIcon
    },
    {
      title: 'Playground',
      url: `#/namespaces/${ns}/playground`,
      icon: BugPlayIcon
    },
    {
      title: 'Analytics',
      url: `#/namespaces/${ns}/analytics`,
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
