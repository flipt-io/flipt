import { GitPullRequest } from 'lucide-react';
import { useState } from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';

import { Button } from '~/components/ui/button';
import { SidebarTrigger } from '~/components/ui/sidebar';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { Badge } from './Badge';
import { CreateBranchButton } from './environments/CreateBranchButton';
import { CreateMergeProposalModal } from './environments/CreateMergeProposalModal';
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
  const [mergeModalOpen, setMergeModalOpen] = useState(false);

  // Determine if this is a branch
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
                Current Namespace and Environment
              </TooltipContent>
            </Tooltip>
          )}
          {currentEnvironment && !currentEnvironment?.configuration?.base && (
            <CreateBranchButton baseEnvironment={currentEnvironment} />
          )}
        </div>
        <div className="flex items-center gap-2">
          {isBranch && (
            <>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="mr-1"
                    onClick={() => setMergeModalOpen(true)}
                  >
                    <GitPullRequest className="size-4" />
                    <span className="sr-only">Create Merge Proposal</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom" align="center">
                  Create Merge Proposal
                </TooltipContent>
              </Tooltip>
              <CreateMergeProposalModal
                open={mergeModalOpen}
                setOpen={setMergeModalOpen}
                environment={currentEnvironment}
              />
            </>
          )}
          {currentEnvironment?.configuration?.remote && (
            <EnvironmentRemoteInfo environment={currentEnvironment} />
          )}
        </div>
      </div>
    </header>
  );
}
