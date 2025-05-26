import { ChevronsUpDown, Folder, GitBranch, Server } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import {
  currentEnvironmentChanged,
  selectAllEnvironments,
  selectCurrentEnvironment
} from '~/app/environments/environmentsApi';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  useListNamespacesQuery
} from '~/app/namespaces/namespacesApi';

import { Popover, PopoverContent, PopoverTrigger } from '~/components/Popover';
import { Button } from '~/components/ui/button';
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

function EnvironmentBranchList({
  environments,
  currentEnvironment,
  setSelectedEnvironment,
  selectedEnvironment
}: {
  environments: IEnvironment[];
  currentEnvironment: string;
  setSelectedEnvironment: (key: string) => void;
  selectedEnvironment: string;
}) {
  const dispatch = useAppDispatch();

  // Handlers
  const handleSelectEnv = (env: IEnvironment) => {
    const key = env.key;
    setSelectedEnvironment(key);
    if (key !== currentEnvironment) {
      dispatch(currentEnvironmentChanged(key));
      dispatch(currentNamespaceChanged(null));
    }
  };
  const handleSelectBranch = (branch: any) => {
    const key = branch.key || branch.environmentKey || '';
    setSelectedEnvironment(key);
    if (key !== currentEnvironment) {
      dispatch(currentEnvironmentChanged(key));
      dispatch(currentNamespaceChanged(null));
    }
  };

  // When current env changes, select it in the left panel
  useEffect(() => {
    setSelectedEnvironment(currentEnvironment);
  }, [setSelectedEnvironment, currentEnvironment]);

  // Group environments: base envs as top-level, branches nested under their base
  const grouped = {} as Record<string, { base: any | null; branches: any[] }>;
  environments.forEach((env) => {
    if (env.configuration?.base) {
      // It's a branch
      const base = env.configuration.base;
      if (!grouped[base]) grouped[base] = { base: null, branches: [] };
      grouped[base].branches.push(env);
    } else {
      // It's a base env
      if (!grouped[env.key]) grouped[env.key] = { base: env, branches: [] };
      else grouped[env.key].base = env;
    }
  });

  const baseEnvKeys = Object.keys(grouped);

  // Render left panel: environments and branches
  return (
    <div
      className="w-1/2  border-r overflow-y-auto"
      data-testid="environment-listbox"
    >
      <div className="p-4 text-xs text-muted-foreground font-semibold uppercase">
        Environments
      </div>
      {baseEnvKeys.map((baseKey) => {
        const group = grouped[baseKey];
        if (!group.base) return null; // skip if no base env
        const env = group.base;
        const branches = group.branches;
        const isSelected = selectedEnvironment === env.key;
        return (
          <div key={env.key}>
            <div className="flex items-center px-2">
              <Button
                variant={isSelected ? 'soft' : 'ghost'}
                size="sm"
                className={`flex-1 gap-2 justify-start px-3 py-1.5 rounded-md ${isSelected ? 'font-semibold' : 'font-normal'}`}
                onClick={() => handleSelectEnv(env)}
              >
                <Server className="w-4 h-4" />
                <span className="truncate">{env.name || env.key}</span>
              </Button>
            </div>
            {branches.length > 0 && (
              <div className="ml-6 pr-2">
                {branches.map((branch: any) => (
                  <Button
                    key={branch.key || branch.environmentKey}
                    variant={
                      selectedEnvironment ===
                      (branch.key || branch.environmentKey)
                        ? 'soft'
                        : 'ghost'
                    }
                    size="sm"
                    className={`w-full justify-start px-3 py-1.5 rounded-md ${selectedEnvironment === (branch.key || branch.environmentKey) ? 'font-semibold' : 'font-normal'}`}
                    onClick={() => handleSelectBranch(branch)}
                  >
                    <GitBranch className="w-4 h-4" />
                    <span className="truncate">
                      {branch.name || branch.environmentKey}
                    </span>
                  </Button>
                ))}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

function NamespaceList({
  namespaces,
  currentNamespace,
  setOpen
}: {
  namespaces: INamespace[];
  currentNamespace: INamespace;
  setOpen: (open: boolean) => void;
}) {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const handleSelectNamespace = (key: string) => {
    if (key !== currentNamespace?.key) {
      dispatch(currentNamespaceChanged(key));
    }
    setOpen(false);
    navigate(`/namespaces/${key}/flags`);
  };

  return (
    <div className="w-1/2 overflow-y-auto" data-testid="namespace-listbox">
      <div className="p-4 text-xs text-muted-foreground font-semibold uppercase">
        Namespaces
      </div>
      {namespaces.length === 0 && (
        <div className="px-4 py-2 text-muted-foreground text-sm">
          No namespaces found
        </div>
      )}
      {namespaces.map((ns: INamespace) => {
        const isSelected = currentNamespace?.key === ns.key;
        return (
          <div key={ns.key} className="px-2">
            <Button
              key={ns.key}
              variant={isSelected ? 'soft' : 'ghost'}
              size="sm"
              className={`w-full gap-2 justify-start px-3 py-1.5 rounded-md ${isSelected ? 'font-semibold' : 'font-normal'}`}
              onClick={() => handleSelectNamespace(ns.key)}
            >
              <Folder className="w-4 h-4" />
              <span className="truncate">{ns.name}</span>
            </Button>
          </div>
        );
      })}
    </div>
  );
}

export function EnvironmentNamespaceSwitcher() {
  const { isMobile } = useSidebar();

  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const currentNamespace = useSelector(selectCurrentNamespace);

  // Get all environments (base + branched) from Redux store
  const environments = useSelector(selectAllEnvironments);

  const [open, setOpen] = useState(false);

  // Track selected env/branch in the left panel
  const [selectedEnvironment, setSelectedEnvironment] = useState<string>('');

  // For the selected environment (base or branch), fetch namespaces
  const { data: namespacesData } = useListNamespacesQuery(
    { environmentKey: selectedEnvironment },
    { skip: !selectedEnvironment, refetchOnMountOrArgChange: true }
  );
  const namespaces = useMemo(
    () => namespacesData?.items ?? [],
    [namespacesData]
  );

  return (
    <SidebarMenu>
      <SidebarMenuItem data-testid="environment-namespace-switcher">
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              onClick={() => setOpen(true)}
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
                  {currentNamespace?.name}
                </span>
                <span className="truncate text-xs">
                  {currentEnvironment?.key}
                </span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </PopoverTrigger>
          <PopoverContent
            className="min-w-[500px] max-w-[90vw] p-0 flex h-[350px]"
            align="start"
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            <EnvironmentBranchList
              environments={environments}
              currentEnvironment={currentEnvironment.key}
              setSelectedEnvironment={setSelectedEnvironment}
              selectedEnvironment={selectedEnvironment}
            />
            <NamespaceList
              namespaces={namespaces}
              currentNamespace={currentNamespace}
              setOpen={setOpen}
            />
          </PopoverContent>
        </Popover>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
