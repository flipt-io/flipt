import { GitBranchIcon, GitPullRequestArrow, GitPullRequestCreate } from 'lucide-react';
import { useState } from 'react';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '~/components/ui/dropdown-menu';

import { IEnvironment, ProposalState, SCM } from '~/types/Environment';

import { getRepoUrlFromConfig } from '~/utils/helpers';

import DeleteBranchModal from './DeleteBranchModal';
import { CreateMergeProposalModal } from './CreateMergeProposalModal';
import { useListBranchEnvironmentsQuery } from '~/app/environments/environmentsApi';

export default function BranchActionsDropdown({
  environment
}: {
  environment: IEnvironment;
}) {
  const baseBranch = environment.configuration?.base ?? '';
  const hasRemote = environment.configuration?.remote !== undefined;
  const repoUrl = getRepoUrlFromConfig(environment.configuration!);

  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [mergeModalOpen, setMergeModalOpen] = useState(false);

  const { data: baseBranches } = useListBranchEnvironmentsQuery({
    baseEnvironmentKey: environment.configuration?.base ?? ''
  });

  const proposal = baseBranches?.branches.find(
    (branch) => branch.environmentKey === environment.key
  )?.proposal;

  let prNumber: string | undefined;

  // TODO: support other SCMs
  if (proposal && proposal.scm === SCM.GITHUB) {
    prNumber = proposal.url.split('/').pop();
  }

  const isProposalOpen = proposal && proposal.state === ProposalState.OPEN;

  const handleViewBranch = () => {
    window.open(repoUrl, '_blank');
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <span className="ml-1 px-2 cursor-pointer">
            <GitBranchIcon className="w-4 h-4" />
          </span>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuLabel className="text-sm font-normal text-gray-500">
            Base branch: <span className="font-mono">{baseBranch}</span>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          {hasRemote && (
            <>
              <DropdownMenuItem onClick={handleViewBranch} className="flex items-center gap-1">
                <GitBranchIcon className="w-4 h-4 mr-2" />
                View branch
              </DropdownMenuItem>
              {!isProposalOpen ? (
                <DropdownMenuItem onClick={() => setMergeModalOpen(true)} className="flex items-center gap-1">
                  <GitPullRequestCreate className="w-4 h-4 mr-2" />
                  Propose changes
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem onClick={() => window.open(proposal?.url ?? '', '_blank')} className="flex items-center gap-1">
                  <GitPullRequestArrow className="w-4 h-4 mr-2" />
                  {prNumber ? (
                    <div className="flex items-center gap-1">
                      <span>View open merge proposal</span>
                      <span className="text-xs font-mono">#{prNumber}</span>
                    </div>
                  ) : (
                    <span>View open merge proposal</span>
                  )}
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
            </>
          )}
          <DropdownMenuItem
            variant="destructive"
            onClick={() => setDeleteModalOpen(true)}
          >
            Delete branch
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <DeleteBranchModal
        open={deleteModalOpen}
        setOpen={setDeleteModalOpen}
        environment={environment}
      />
      <CreateMergeProposalModal
        open={mergeModalOpen}
        setOpen={setMergeModalOpen}
        environment={environment}
      />
    </>
  );
}
