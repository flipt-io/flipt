import { GitBranchIcon } from 'lucide-react';
import { useState } from 'react';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '~/components/ui/dropdown-menu';

import { IEnvironment } from '~/types/Environment';

import { getRepoUrlFromConfig } from '~/utils/helpers';

import DeleteBranchModal from './DeleteBranchModal';

export default function BranchActionsDropdown({
  environment
}: {
  environment: IEnvironment;
}) {
  const baseBranch = environment.configuration?.base ?? '';
  const hasRemote = environment.configuration?.remote !== undefined;
  const repoUrl = getRepoUrlFromConfig(environment.configuration!);

  const [openDeleteModal, setOpenDeleteModal] = useState(false);

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
              <DropdownMenuItem onClick={handleViewBranch}>
                View branch
              </DropdownMenuItem>
              <DropdownMenuSeparator />
            </>
          )}
          <DropdownMenuItem
            variant="destructive"
            onClick={() => setOpenDeleteModal(true)}
          >
            Delete branch
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <DeleteBranchModal
        open={openDeleteModal}
        setOpen={setOpenDeleteModal}
        environment={environment}
      />
    </>
  );
}
