import {
  BellIcon,
  HelpCircleIcon,
  SettingsIcon,
  TerminalIcon
} from 'lucide-react';
import * as React from 'react';
import { useSelector } from 'react-redux';

import { selectInfo } from '~/app/meta/metaSlice';

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar
} from '~/components/ui/sidebar';

const items = [
  {
    title: 'Settings',
    url: '#/settings',
    icon: SettingsIcon
  },
  {
    title: 'Get Help',
    url: '#/support',
    icon: HelpCircleIcon
  }
];

function emitCtrlK() {
  const event = new KeyboardEvent('keydown', {
    key: 'k',
    code: 'KeyK',
    ctrlKey: true,
    bubbles: true
  });
  document.dispatchEvent(event);
}

export function NavSecondary({
  ...props
}: React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { isMobile } = useSidebar();
  const info = useSelector(selectInfo);
  const updateAvailable = info && info.build.updateAvailable;

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {updateAvailable && (
            <SidebarMenuItem key="update-available">
              <SidebarMenuButton
                asChild
                tooltip="A new version of Flipt is available!"
              >
                <a href={info.build.latestVersionURL}>
                  <BellIcon />
                  <span>Update Flipt</span>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          )}

          {!isMobile && (
            <SidebarMenuItem
              key="run-command"
              onClick={(e) => {
                e.preventDefault();
                emitCtrlK();
              }}
            >
              <SidebarMenuButton asChild tooltip="Run (ctrl + k)">
                <a>
                  <TerminalIcon />
                  <span>Run (ctrl + k)</span>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          )}
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton asChild tooltip={item.title}>
                <a href={item.url}>
                  <item.icon />
                  <span>{item.title}</span>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
