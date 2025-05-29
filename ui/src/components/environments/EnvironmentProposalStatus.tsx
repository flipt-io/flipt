import { GitPullRequestArrow, GitPullRequestCreate } from 'lucide-react';
import { useState } from 'react';

import { useListBranchEnvironmentsQuery } from '~/app/environments/environmentsApi';

import { Button } from '~/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import {
  IEnvironment,
  IEnvironmentProposal,
  ProposalState,
  SCM
} from '~/types/Environment';

import { CreateMergeProposalModal } from './branches/CreateMergeProposalModal';

function OpenMergeProposalStatus({
  proposal
}: {
  proposal: IEnvironmentProposal;
}) {
  let prNumber: string | undefined;

  // TODO: support other SCMs
  if (proposal.scm === SCM.GITHUB) {
    prNumber = proposal.url.split('/').pop();
  }

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="mr-1 cursor-pointer"
            onClick={() => {
              window.open(proposal.url, '_blank');
            }}
          >
            <GitPullRequestArrow className={`w-4 h-4 text-green-500`} />
            <span className="sr-only">Merge proposal</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" align="center">
          {prNumber ? (
            <div className="flex items-center gap-1">
              <span>View open merge proposal</span>
              <span className="text-xs font-mono">#{prNumber}</span>
            </div>
          ) : (
            <span>View open merge proposal</span>
          )}
        </TooltipContent>
      </Tooltip>
    </>
  );
}

function NoMergeProposalStatus({ environment }: { environment: IEnvironment }) {
  const [mergeModalOpen, setMergeModalOpen] = useState(false);
  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="mr-1 cursor-pointer"
            onClick={() => setMergeModalOpen(true)}
          >
            <GitPullRequestCreate className="size-4" />
            <span className="sr-only">Create merge proposal</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" align="center">
          Create merge proposal
        </TooltipContent>
      </Tooltip>
      <CreateMergeProposalModal
        open={mergeModalOpen}
        setOpen={setMergeModalOpen}
        environment={environment}
      />
    </>
  );
}

export default function EnvironmentProposalStatus({
  environment
}: {
  environment: IEnvironment;
}) {
  const { data: baseBranches } = useListBranchEnvironmentsQuery({
    baseEnvironmentKey: environment.configuration?.base ?? ''
  });

  const proposal = baseBranches?.branches.find(
    (branch) => branch.environmentKey === environment.key
  )?.proposal;

  if (proposal && proposal.state === ProposalState.OPEN) {
    return <OpenMergeProposalStatus proposal={proposal} />;
  }

  return <NoMergeProposalStatus environment={environment} />;
}
