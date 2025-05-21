import { GitBranchIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import * as Yup from 'yup';

import {
  currentEnvironmentChanged,
  useCreateBranchEnvironmentMutation,
  useListBranchEnvironmentsQuery,
  useListEnvironmentsQuery
} from '~/app/environments/environmentsApi';

import Combobox from '~/components/Combobox';

import { IBranchEnvironment, IEnvironment } from '~/types/Environment';
import { ISelectable } from '~/types/Selectable';

import { useError } from '~/data/hooks/error';
import { useAppDispatch } from '~/data/hooks/store';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation } from '~/data/validations';
import { cn } from '~/lib/utils';

export function EnvironmentBranchSelector({
  environment,
  className
}: {
  environment: IEnvironment;
  className?: string;
}) {
  const dispatch = useAppDispatch();
  const {
    data: environmentsData,
    isLoading: isEnvironmentsLoading,
    error: environmentsError
  } = useListEnvironmentsQuery();
  const environments = useMemo(
    () => environmentsData?.environments ?? [],
    [environmentsData]
  );

  // Determine the base environment (if branched, use its base)
  const isBranched = !!environment.configuration?.base;
  const baseEnvKey = isBranched
    ? environment.configuration?.base || ''
    : environment.key;
  const baseEnvironment = environments.find(
    (env: IEnvironment) => env.key === baseEnvKey
  );

  // Query branches for the base environment
  const {
    data: branchesData,
    isLoading: isBranchesLoading,
    error: branchesError
  } = useListBranchEnvironmentsQuery(
    { baseEnvironmentKey: baseEnvKey },
    { skip: !baseEnvKey }
  );
  const branches = useMemo(() => branchesData?.branches ?? [], [branchesData]);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const branchValidationSchema = Yup.object({
    branchName: keyValidation.test(
      'is-valid-branch-name',
      'Invalid branch name',
      (value) => {
        if (!value) return false;
        return !/^flipt\/[\w-]+$/.test(value);
      }
    )
  });

  const [createBranchEnvironment] = useCreateBranchEnvironmentMutation();

  const [inputValue, setInputValue] = useState('');
  const [pendingEnvKey, setPendingEnvKey] = useState<string | null>(null);

  // Handler for creating a new branch
  const handleCreateBranch = async (branchName: string) => {
    try {
      await branchValidationSchema.validate({ branchName });
      await createBranchEnvironment({
        baseEnvironmentKey: baseEnvKey,
        environmentKey: branchName
      }).unwrap();
      setPendingEnvKey(branchName);
      setInputValue('');
      clearError();
      setSuccess(`Successfully created branch "${branchName}"`);
    } catch (error) {
      setError(error);
    }
  };

  const changeEnvironment = (option: ISelectable | null) => {
    if (!option) return;
    if (option.key.startsWith('__create__')) {
      handleCreateBranch(option.key.replace('__create__', ''));
      return;
    }
    const env = environments.find(
      (el: IEnvironment) => el.key == option.key
    ) as IEnvironment;
    if (env) {
      dispatch(currentEnvironmentChanged(env));
    }
    setInputValue('');
  };

  useEffect(() => {
    if (
      pendingEnvKey &&
      environments.some((env) => env.key === pendingEnvKey)
    ) {
      dispatch(currentEnvironmentChanged({ key: pendingEnvKey }));
      setPendingEnvKey(null);
    }
  }, [environments, pendingEnvKey, dispatch]);

  // Build branch options for the base environment
  const branchOptions: ISelectable[] = useMemo(() => {
    if (!baseEnvironment) return [];
    const baseOption: ISelectable = {
      key: baseEnvironment.key,
      displayValue: `${baseEnvironment.configuration?.branch || baseEnvironment.key}`
    };
    const branchList =
      branches.map((branch: IBranchEnvironment) => ({
        key: branch.environmentKey,
        displayValue: branch.branch.replace(/^flipt\/[\w-]+\//, '')
      })) || [];
    return [baseOption, ...branchList];
  }, [baseEnvironment, branches]);

  // Set the selected value to the current environment
  const selectedOption =
    branchOptions.find((opt) => opt.key === environment.key) ||
    branchOptions[0];

  // Filter options based on input value
  const filteredOptions = branchOptions.filter((opt) =>
    opt.displayValue.toLowerCase().includes(inputValue.toLowerCase())
  );

  const branchExists = branchOptions.some(
    (opt) => opt.displayValue.toLowerCase() === inputValue.toLowerCase()
  );

  // Add a "Create branch" option if the input doesn't match any existing branch
  const comboboxOptions =
    inputValue && !branchExists
      ? [
          ...filteredOptions,
          {
            key: `__create__${inputValue}`,
            displayValue: `Create branch "${inputValue}"`
          }
        ]
      : filteredOptions;

  // Handle loading and error states
  if (!baseEnvKey) return null;
  if (isEnvironmentsLoading || isBranchesLoading) {
    return (
      <div className="text-xs text-muted-foreground px-2 py-1">
        Loading environments or branches...
      </div>
    );
  }
  if (environmentsError || branchesError) {
    return (
      <div className="text-xs text-red-500 px-2 py-1">
        Error loading environments or branches
      </div>
    );
  }
  if (comboboxOptions.length === 0) return null;

  return (
    <div className={cn('flex items-center', className)}>
      <GitBranchIcon className="h-4 w-4 text-muted-foreground" />
      <Combobox<ISelectable>
        id="environment-branch-combobox"
        name="environmentBranch"
        values={comboboxOptions}
        selected={selectedOption}
        setSelected={changeEnvironment}
        placeholder="Switch or create branch"
        disabled={isBranchesLoading}
        className="border-none font-semibold text-xs focus:outline-none focus:ring-0 ring-0"
        onInputChange={setInputValue}
      />
    </div>
  );
}
