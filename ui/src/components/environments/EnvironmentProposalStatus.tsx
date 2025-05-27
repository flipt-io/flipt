import {
  GitPullRequestArrow,
  GitPullRequestClosed,
  GitPullRequestCreate
} from 'lucide-react';
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

import { CreateMergeProposalModal } from './CreateMergeProposalModal';

function MergeProposalStatus({ proposal }: { proposal: IEnvironmentProposal }) {
  let prNumber: string | undefined;

  // TODO: support other SCMs
  if (proposal.scm === SCM.GITHUB) {
    prNumber = proposal.url.split('/').pop();
  }

  const statusColor =
    proposal.state === ProposalState.OPEN ? 'text-green-500' : '';
  const statusIcon =
    proposal.state === ProposalState.OPEN ? (
      <GitPullRequestArrow className={`w-4 h-4 ${statusColor}`} />
    ) : (
      <GitPullRequestClosed className={`w-4 h-4 ${statusColor}`} />
    );
  const statusText =
    proposal.state === ProposalState.OPEN
      ? 'View open merge proposal'
      : 'View closed merge proposal';

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
            {statusIcon}
            <span className="sr-only">Merge proposal</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" align="center">
          {prNumber ? (
            <div className="flex items-center gap-1">
              <span>{statusText}</span>
              <span className="text-xs font-mono">#{prNumber}</span>
            </div>
          ) : (
            <span>{statusText}</span>
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

  if (proposal) {
    return <MergeProposalStatus proposal={proposal} />;
  }

  return <NoMergeProposalStatus environment={environment} />;
}
