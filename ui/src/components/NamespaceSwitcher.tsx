import { ChevronsUpDown } from 'lucide-react';
import * as React from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger
} from '~/components/ui/dropdown-menu';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar
} from '~/components/ui/sidebar';

import { INamespace } from '~/types/Namespace';

import logoFlag from '~/assets/logo-flag.png';
import { useAppDispatch } from '~/data/hooks/store';

export interface IEnv {
  key: string;
  namespaces: INamespace[];
}
const emptyNamespaces: INamespace[] = [];

function Environments({
  items,
  onSelect
}: {
  items: IEnv[];
  onSelect: (env: string, ns: string) => void;
}) {
  if (items.length == 1) {
    const environment = items[0];
    return environment.namespaces.map((ns) => (
      <DropdownMenuItem
        key={ns.name}
        onClick={() => onSelect(environment.key, ns.key)}
        className="gap-2 p-2"
      >
        <span className="truncate">{ns.name}</span>
      </DropdownMenuItem>
    ));
  }
  return items.map((environment) => {
    return (
      <DropdownMenuSub key={environment.key}>
        <DropdownMenuSubTrigger>{environment.key}</DropdownMenuSubTrigger>
        <DropdownMenuPortal>
          <DropdownMenuSubContent>
            {environment.namespaces.map((ns) => (
              <DropdownMenuItem
                key={environment.key + ns.key}
                onClick={() => onSelect(environment.key, ns.key)}
                className="gap-2 p-2"
              >
                <span className="truncate">{ns.name}</span>
              </DropdownMenuItem>
            ))}
          </DropdownMenuSubContent>
        </DropdownMenuPortal>
      </DropdownMenuSub>
    );
  });
}

export function NamespaceSwitcher() {
  const { isMobile } = useSidebar();

  const environment = useSelector(selectCurrentEnvironment);
  const namespaces = useSelector(selectNamespaces);
  const items = React.useMemo(() => {
    // FIXME: use real environments with their namespace
    return [
      {
        key: environment.key,
        namespaces: namespaces || emptyNamespaces
      }
    ];
  }, [environment, namespaces]);

  const namespace = useSelector(selectCurrentNamespace);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const setCurrentNamespace = (namespace: INamespace) => {
    dispatch(currentNamespaceChanged(namespace));
    navigate(`/namespaces/${namespace.key}`);
  };

  const changeNamespace = (_env: string, ns: string) => {
    const value = namespaces?.find((el) => el.key == ns) as INamespace;
    value && setCurrentNamespace(value);
  };

  const activeNamespace = React.useMemo(() => {
    return {
      key: namespace.key,
      name: namespace.name,
      environment: environment.key
    };
  }, [namespace, environment]);

  if (!activeNamespace) {
    return null;
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg ">
                <img
                  src={logoFlag}
                  alt="logo"
                  width={512}
                  height={512}
                  className="m-auto h-8 w-8"
                />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {activeNamespace.name}
                </span>
                <span className="truncate text-xs">
                  {activeNamespace.environment}
                </span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
            align="start"
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-xs text-muted-foreground">
              Namespaces
            </DropdownMenuLabel>
            <Environments items={items} onSelect={changeNamespace} />
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
