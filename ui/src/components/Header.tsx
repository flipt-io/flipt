import { GitBranchPlusIcon } from 'lucide-react';
import { useState } from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';

import { SidebarTrigger } from '~/components/ui/sidebar';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { Badge } from './Badge';
import BranchActionsDropdown from './environments/branches/BranchActionsDropdown';
import { CreateBranchPopover } from './environments/branches/CreateBranchPopover';

export function Header({
  ns,
  env,
  sidebarOpen
}: {
  ns: string;
  env: string;
  sidebarOpen: boolean;
}) {
  const info = useSelector(selectInfo);
  const currentEnvironment = useSelector(selectCurrentEnvironment);

  const [createBranchOpen, setCreateBranchOpen] = useState(false);

  const isBranch = currentEnvironment?.configuration?.base !== undefined;

  const topbarStyle = {
    backgroundColor: info?.ui?.topbarColor,
    borderRadius: '1rem 1rem 0 0'
  };

  return (
    <header
      className="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
      style={topbarStyle}
    >
      <div className="flex w-full items-center px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-2 flex-shrink-0" />
        <div className="flex-1" />
        <div className="flex items-center gap-2 flex-shrink-0">
          {!sidebarOpen && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge
                  variant="secondary"
                  className="px-3 py-1 bg-background text-xs flex-shrink-0 cursor-pointer"
                >
                  {ns} <span className="mx-1 text-muted-foreground">•</span>{' '}
                  {env}
                </Badge>
              </TooltipTrigger>
              <TooltipContent side="bottom" align="end">
                Current namespace and environment
              </TooltipContent>
            </Tooltip>
          )}
          {isBranch ? (
            <BranchActionsDropdown environment={currentEnvironment} />
          ) : (
            <CreateBranchPopover
              open={createBranchOpen}
              setOpen={setCreateBranchOpen}
              environment={currentEnvironment}
            >
              <Badge
                variant="secondary"
                className="px-3 py-1 bg-background text-xs flex-shrink-0 cursor-pointer"
                data-testid="create-branch-button"
              >
                <span className="flex items-center gap-2 text-xs">
                  <GitBranchPlusIcon className="w-4 h-4" />
                  Branch environment
                </span>
              </Badge>
            </CreateBranchPopover>
          )}
        </div>
      </div>
    </header>
  );
}
