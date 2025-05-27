import { GitBranchIcon, GitBranchPlusIcon, GitPullRequest } from 'lucide-react';
import { useState } from 'react';
import { useSelector } from 'react-redux';

import {
  selectCurrentEnvironment,
  useListBranchEnvironmentsQuery
} from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';

import { Button } from '~/components/ui/button';
import { SidebarTrigger } from '~/components/ui/sidebar';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { Badge } from './Badge';
import { CreateBranchPopover } from './environments/CreateBranchPopover';
import EnvironmentProposalStatus from './environments/EnvironmentProposalStatus';
import { EnvironmentRemoteInfo } from './environments/EnvironmentRemoteInfo';

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
  const hasRemote = currentEnvironment?.configuration?.remote !== undefined;

  const topbarStyle = {
    backgroundColor: info?.ui?.topbarColor,
    borderRadius: '1rem 1rem 0 0'
  };
  return (
    <header
      className="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
      style={topbarStyle}
    >
      <div className="flex w-full items-center justify-between px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-2" />
        <div className="flex items-center gap-2">
          {!sidebarOpen && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge
                  variant="secondary"
                  className="px-3 py-1 bg-background font-semibold text-xs"
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
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="ml-1 px-2 cursor-pointer">
                  <GitBranchIcon className="w-4 h-4 text-gray-400" />
                </span>
              </TooltipTrigger>
              <TooltipContent side="bottom" align="center">
                Branched from {currentEnvironment?.configuration?.base}
              </TooltipContent>
            </Tooltip>
          ) : (
            <CreateBranchPopover
              open={createBranchOpen}
              setOpen={setCreateBranchOpen}
              environment={currentEnvironment}
            >
              <Button
                variant="ghost"
                size="icon"
                className="ml-1 cursor-pointer"
                type="button"
                data-testid="create-branch-button"
              >
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span>
                      <GitBranchPlusIcon className="w-4 h-4" />
                    </span>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" align="center">
                    Create branch
                  </TooltipContent>
                </Tooltip>
              </Button>
            </CreateBranchPopover>
          )}
          {isBranch && hasRemote && (
            <EnvironmentProposalStatus environment={currentEnvironment} />
          )}
          {hasRemote && (
            <EnvironmentRemoteInfo environment={currentEnvironment} />
          )}
        </div>
      </div>
    </header>
  );
}
