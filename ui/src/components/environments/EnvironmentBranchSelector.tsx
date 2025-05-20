import { GitBranchIcon } from 'lucide-react';
import { useMemo } from 'react';
import { useSelector } from 'react-redux';

import {
  currentEnvironmentChanged,
  selectAllEnvironments,
  useListBranchEnvironmentsQuery
} from '~/app/environments/environmentsApi';

import Listbox from '~/components/forms/Listbox';

import { IBranchEnvironment, IEnvironment } from '~/types/Environment';
import { ISelectable } from '~/types/Selectable';

import { useAppDispatch } from '~/data/hooks/store';
import { cn } from '~/lib/utils';

export function EnvironmentBranchSelector({
  environment,
  className
}: {
  environment: IEnvironment;
  className?: string;
}) {
  const dispatch = useAppDispatch();
  const environments = useSelector(selectAllEnvironments);

  // Determine the base environment (if branched, use its base)
  const isBranched = !!environment.configuration?.base;
  const baseEnvKey = isBranched
    ? environment.configuration?.base || ''
    : environment.key;
  const baseEnvironment = environments.find((env) => env.key === baseEnvKey);

  // Query branches for the base environment
  const { data, isLoading, error } = useListBranchEnvironmentsQuery(
    { baseEnvironmentKey: baseEnvKey },
    { skip: !baseEnvKey }
  );

  const changeEnvironment = (e: ISelectable) => {
    const env = environments?.find((el) => el.key == e.key) as IEnvironment;
    if (env) {
      dispatch(currentEnvironmentChanged(env));
    }
  };

  // Build branch options for the base environment
  const branchOptions: ISelectable[] = useMemo(() => {
    if (!baseEnvironment) return [];
    const baseOption: ISelectable = {
      key: baseEnvironment.key,
      displayValue: `${baseEnvironment.configuration?.branch || baseEnvironment.key}`
    };
    const branchList =
      data?.branches?.map((branch: IBranchEnvironment) => ({
        key: branch.environmentKey,
        displayValue: branch.branch.replace(/^flipt\/[\w-]+\//, '')
      })) || [];
    return [baseOption, ...branchList];
  }, [baseEnvironment, data]);

  // Set the selected value to the current environment
  const selectedOption =
    branchOptions.find((opt) => opt.key === environment.key) ||
    branchOptions[0];

  if (!baseEnvKey) return null;
  if (branchOptions.length <= 1) return null;
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
    <div className={cn('flex items-center', className)}>
      <GitBranchIcon className="h-4 w-4 text-muted-foreground" />
      <Listbox<ISelectable>
        id="environment-branch-selector"
        name="environmentBranch"
        values={branchOptions}
        selected={selectedOption}
        setSelected={changeEnvironment}
        className="border-none font-semibold text-xs focus:outline-none focus:ring-0 ring-0"
      />
    </div>
  );
}
