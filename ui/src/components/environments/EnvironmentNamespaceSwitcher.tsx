import { ChevronsUpDown } from 'lucide-react';
import * as React from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import {
  currentEnvironmentChanged,
  selectCurrentEnvironment,
  selectEnvironments
} from '~/app/environments/environmentsApi';
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
  DropdownMenuSeparator,
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

import { IEnvironment } from '~/types/Environment';
import { INamespace } from '~/types/Namespace';

import logoFlag from '~/assets/logo-flag.png';
import { useAppDispatch } from '~/data/hooks/store';

export interface IEnv {
  key: string;
  namespaces: INamespace[];
}
const emptyNamespaces: INamespace[] = [];

function Namespaces({
  env,
  onSelect
}: {
  env: IEnv;
  onSelect: (env: string, ns: string) => void;
}) {
  return (
    <>
      <DropdownMenuLabel className="text-xs text-muted-foreground">
        Namespaces
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      {env.namespaces.map((ns) => (
        <DropdownMenuItem
          key={ns.name}
          onClick={() => onSelect(env.key, ns.key)}
          className="gap-2 p-2"
        >
          <span className="truncate">{ns.name}</span>
        </DropdownMenuItem>
      ))}
    </>
  );
}

function Environments({
  items,
  activeEnvironment,
  onSelect
}: {
  items: IEnv[];
  activeEnvironment: string;
  onSelect: (env: string, ns: string) => void;
}) {
  if (items.length == 1) {
    const environment = items[0];

    return <Namespaces env={environment} onSelect={onSelect} />;
  }

  return (
    <>
      <DropdownMenuLabel className="text-xs text-muted-foreground">
        Environments
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      {items.map((environment) => {
        if (environment.key == activeEnvironment) {
          return (
            <DropdownMenuSub key={environment.key}>
              <DropdownMenuSubTrigger>{environment.key}</DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent sideOffset={4}>
                  <Namespaces env={environment} onSelect={onSelect} />
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
          );
        }
        return (
          <DropdownMenuItem
            key={environment.key}
            onClick={() => onSelect(environment.key, 'default')}
            className="gap-2 p-2"
          >
            <span className="truncate">{environment.key}</span>
          </DropdownMenuItem>
        );
      })}
    </>
  );
}

export function EnvironmentNamespaceSwitcher() {
  const { isMobile } = useSidebar();

  const environment = useSelector(selectCurrentEnvironment);
  const environments = useSelector(selectEnvironments);
  const namespaces = useSelector(selectNamespaces);
  const items = React.useMemo(() => {
    return environments.map((env) => {
      const item = {
        key: env.key,
        namespaces: emptyNamespaces
      };
      if (env.key == environment.key) {
        item.namespaces = namespaces || emptyNamespaces;
      }
      return item;
    });
  }, [environment, namespaces, environments]);

  const namespace = useSelector(selectCurrentNamespace);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const setCurrentNamespace = (namespace: INamespace) => {
    dispatch(currentNamespaceChanged(namespace));
  };

  const changeEnvironment = (key: string) => {
    const env = environments?.find((el) => el.key == key) as IEnvironment;
    if (env) {
      dispatch(currentEnvironmentChanged(env));
    }
  };

  const changeNamespace = (ns: string) => {
    const value = namespaces?.find((el) => el.key == ns) as INamespace;
    value && setCurrentNamespace(value);
  };

  const activeNamespace = React.useMemo(() => {
    return {
      key: namespace.key,
      name: namespace.name,
      environment: environment?.configuration?.base || environment.key
    };
  }, [namespace, environment]);

  const onSelect = (env: string, ns: string) => {
    if (activeNamespace.environment != env) {
      changeEnvironment(env);
    }
    if (activeNamespace.key != ns) {
      changeNamespace(ns);
    }
    navigate(`/namespaces/${ns}/flags`);
  };

  if (!activeNamespace) {
    return null;
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem data-testid="namespace-listbox">
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
            data-testid="namespace-listbox-options"
          >
            <Environments
              activeEnvironment={environment.key}
              items={items}
              onSelect={onSelect}
            />
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
