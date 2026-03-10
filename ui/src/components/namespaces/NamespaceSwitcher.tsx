import { Folder, ChevronDown, ChevronRight } from 'lucide-react';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import { Popover, PopoverContent, PopoverTrigger } from '~/components/Popover';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar
} from '~/components/Sidebar';
import { useAppDispatch } from '~/data/hooks/store';

import logoFlag from '~/assets/logo-flag.png';

export function NamespaceSwitcher() {
  const dispatch = useAppDispatch();
  const namespaces = useSelector(selectNamespaces);
  const currentNamespace = useSelector(selectCurrentNamespace);
  const { isMobile } = useSidebar();

  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  const handleSelectNamespace = (key: string) => {
    const ns = namespaces.find((n) => n.key === key);
    if (ns) {
      dispatch(currentNamespaceChanged(ns));
      navigate(`/namespaces/${key}/flags`);
      setOpen(false);
    }
  };

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <SidebarMenuButton size="lg" aria-label="namespace switcher">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg">
                <img
                  src={logoFlag}
                  alt="logo"
                  className="m-auto h-8 w-8"
                  width={512}
                  height={512}
                />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {currentNamespace?.name || currentNamespace?.key || 'default'}
                </span>
              </div>
              {isMobile && <ChevronDown className="ml-auto" />}
              {!isMobile && <ChevronRight className="ml-auto" />}
            </SidebarMenuButton>
          </PopoverTrigger>
          <PopoverContent
            className="bg-sidebar-accent text-sidebar-accent-foreground border-sidebar h-[350px] max-w-[40vw] min-w-[250px] p-1"
            align="start"
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            <div className="p-2 text-xs font-semibold uppercase">
              Namespaces
            </div>
            <div className="max-h-[300px] overflow-y-auto">
              {namespaces.map((ns) => (
                <button
                  key={ns.key}
                  className="hover:bg-sidebar hover:text-sidebar-foreground flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none"
                  onClick={() => handleSelectNamespace(ns.key)}
                >
                  <Folder className="h-4 w-4" />
                  <span className="truncate">{ns.name || ns.key}</span>
                </button>
              ))}
            </div>
          </PopoverContent>
        </Popover>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
