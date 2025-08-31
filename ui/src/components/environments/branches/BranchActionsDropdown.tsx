import {
  FolderGit,
  GitPullRequestArrow,
  GitPullRequestCreate,
  Github,
  Gitlab,
  Server,
  Trash2Icon
} from 'lucide-react';
import { useState } from 'react';

import { useListBranchEnvironmentsQuery } from '~/app/environments/environmentsApi';

import { Badge } from '~/components/Badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '~/components/DropdownMenu';

import { IEnvironment, ProposalState, SCM } from '~/types/Environment';
import { Product } from '~/types/Meta';

import { useAppSelector } from '~/data/hooks/store';
import { getRepoUrlFromConfig } from '~/utils/helpers';

import { CreateMergeProposalModal } from './CreateMergeProposalModal';
import DeleteBranchModal from './DeleteBranchModal';

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

  const { info } = useAppSelector((state) => state.meta);
  const { data: baseBranches } = useListBranchEnvironmentsQuery({
    environmentKey: environment.configuration?.base ?? ''
  });
  const branch = baseBranches?.branches.find(
    (branch) => branch.key === environment.key
  );

  const proposal = branch?.proposal;
  const scm = environment.configuration?.scm;

  const handleViewBranch = () => {
    window.open(repoUrl, '_blank');
  };

  let ProviderIcon = FolderGit;
  if (scm === SCM.GITHUB) ProviderIcon = Github;
  if (scm === SCM.GITLAB) ProviderIcon = Gitlab;

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Badge
            variant="secondary"
            className="flex items-center gap-2 px-2 py-1 bg-background font-semibold text-xs cursor-pointer"
          >
            <Server className="w-4 h-4" />
            Branched from: <span className="font-mono">{baseBranch}</span>
          </Badge>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          {hasRemote && (
            <>
              <DropdownMenuItem
                onClick={handleViewBranch}
                className="flex items-center gap-1"
              >
                <ProviderIcon className="w-4 h-4 mr-2" />
                View remote
              </DropdownMenuItem>
              {info.product === Product.PRO && (
                <>
                  {proposal?.state !== ProposalState.OPEN ? (
                    <DropdownMenuItem
                      onClick={() => setMergeModalOpen(true)}
                      className="flex items-center gap-1"
                    >
                      <GitPullRequestCreate className="w-4 h-4 mr-2" />
                      Propose changes
                    </DropdownMenuItem>
                  ) : (
                    <DropdownMenuItem
                      onClick={() => window.open(proposal?.url ?? '', '_blank')}
                      className="flex items-center gap-1"
                    >
                      <GitPullRequestArrow className="w-4 h-4 mr-2" />
                      <div className="flex items-center gap-1">
                        <span>View open merge proposal</span>
                      </div>
                    </DropdownMenuItem>
                  )}
                </>
              )}
              <DropdownMenuSeparator />
            </>
          )}
          <DropdownMenuItem
            variant="destructive"
            onClick={() => setDeleteModalOpen(true)}
            className="flex items-center gap-1"
          >
            <Trash2Icon className="w-4 h-4 mr-2" />
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
