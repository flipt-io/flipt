import { GitBranchIcon } from 'lucide-react';
import React, { useMemo } from 'react';
import { useSelector } from 'react-redux';

import {
  selectCurrentEnvironment,
  useListBranchEnvironmentsQuery
} from '~/app/environments/environmentsApi';

import Listbox from '~/components/forms/Listbox';

import { IBranchEnvironment } from '~/types/Environment';
import { ISelectable } from '~/types/Selectable';

export function EnvironmentBranchSelector({
  selectedBranch,
  setSelectedBranch,
  className
}: {
  selectedBranch: ISelectable | null;
  setSelectedBranch: (branch: ISelectable) => void;
  className?: string;
}) {
  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const baseEnvironmentKey = currentEnvironment?.key;

  const { data, isLoading, error } = useListBranchEnvironmentsQuery(
    { baseEnvironmentKey },
    { skip: !baseEnvironmentKey }
  );

  const branchOptions: ISelectable[] = useMemo(() => {
    if (!currentEnvironment) return [];
    const baseOption: ISelectable = {
      key: currentEnvironment.configuration?.branch || currentEnvironment.key,
      displayValue: `${currentEnvironment.configuration?.branch || currentEnvironment.key}`
    };
    const branchList =
      data?.branches?.map((branch: IBranchEnvironment) => ({
        key: branch.environmentKey,
        displayValue: branch.branch
      })) || [];
    return [baseOption, ...branchList];
  }, [currentEnvironment, data]);

  if (!baseEnvironmentKey) return null;
  if (isLoading)
    return (
      <div className="text-xs text-muted-foreground px-2 py-1">
        Loading branches...
      </div>
    );
  if (error)
    return (
      <div className="text-xs text-red-500 px-2 py-1">
        Error loading branches
      </div>
    );

  return (
    <div className="flex items-center">
      <GitBranchIcon className="h-4 w-4 text-muted-foreground" />
      <Listbox<ISelectable>
        id="environment-branch-selector"
        name="environmentBranch"
        values={branchOptions}
        selected={selectedBranch || branchOptions[0]}
        setSelected={setSelectedBranch}
        className="border-none font-semibold text-xs"
      />
    </div>
  );
}
