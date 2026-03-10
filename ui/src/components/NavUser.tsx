import { MoreVerticalIcon, LogOutIcon, User2 } from 'lucide-react';
import { useNavigate } from 'react-router';

import { Avatar, AvatarFallback, AvatarImage } from '~/components/Avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '~/components/ui/dropdown-menu';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar
} from '~/components/Sidebar';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { expireAuthSelf } from '~/data/api';

type User = {
  name?: string;
  email?: string;
  issuer?: string;
  imgURL?: string;
};

export function NavUser({ user }: { user: User }) {
  const { isMobile } = useSidebar();
  const { setError } = useError();
  const { clearSession } = useSession();
  const navigate = useNavigate();

  const logout = async () => {
    try {
      await expireAuthSelf();
      clearSession();
      if (user?.issuer) {
        window.location.href = `//${user.issuer}`;
      } else {
        navigate('/login');
      }
    } catch (err) {
      setError(err);
    }
  };

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="h-8 w-8 rounded-lg">
                <AvatarImage src={user.imgURL} alt={user.name} />
                <AvatarFallback className="rounded-lg">
                  <User2 className="h-4 w-4" />
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {user.name || 'User'}
                </span>
                <span className="truncate text-xs">{user.issuer || ''}</span>
              </div>
              <MoreVerticalIcon className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
            side={isMobile ? 'bottom' : 'right'}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="h-8 w-8 rounded-lg">
                  <AvatarImage src={user.imgURL} alt={user.name} />
                  <AvatarFallback className="rounded-lg">
                    <User2 className="h-4 w-4" />
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">{user.name}</span>
                  <span className="truncate text-xs">{user.email}</span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={logout}>
              <LogOutIcon />
              Log out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
