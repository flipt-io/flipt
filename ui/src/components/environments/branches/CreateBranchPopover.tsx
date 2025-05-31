import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import * as Yup from 'yup';

import {
  currentEnvironmentChanged,
  selectAllEnvironments,
  useCreateBranchEnvironmentMutation
} from '~/app/environments/environmentsApi';

import MoreInfo from '~/components/MoreInfo';
import { Popover, PopoverContent, PopoverTrigger } from '~/components/Popover';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';

import { IEnvironment } from '~/types/Environment';

import { useError } from '~/data/hooks/error';
import { useAppDispatch } from '~/data/hooks/store';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation } from '~/data/validations';

export function CreateBranchPopover({
  open,
  setOpen,
  environment,
  children
}: {
  environment: IEnvironment;
  open: boolean;
  setOpen: (open: boolean) => void;
  children: React.ReactNode;
}) {
  const [branchInput, setBranchInput] = useState('');

  const [createBranch, { isLoading: isCreatingBranch }] =
    useCreateBranchEnvironmentMutation();

  // Get all environments (base + branched) from Redux store
  const environments = useSelector(selectAllEnvironments);

  const environmentNames = useMemo(
    () => environments.map((env) => env.name),
    [environments]
  );

  const { setSuccess } = useSuccess();
  const { setError, clearError } = useError();
  const dispatch = useAppDispatch();

  const branchValidationSchema = Yup.object({
    branchName: keyValidation
      .test('is-valid-branch-name', 'Invalid branch name', (value) => {
        if (!value) return false;
        return !/^flipt\/[\w-]+$/.test(value);
      })
      .test(
        'not-duplicate-environment',
        'Branch name cannot match an existing environment (case-insensitive)',
        (value) => {
          if (!value) return true;
          return !environmentNames.some(
            (envName) => envName?.toLowerCase() === value.toLowerCase()
          );
        }
      )
  });

  if (!environment) return null;

  const handleCreateBranch = async () => {
    const branchName = branchInput.trim();
    if (!branchName) return;
    try {
      await branchValidationSchema.validate({ branchName });
      await createBranch({
        environmentKey: environment.key,
        key: branchName
      }).unwrap();
      setBranchInput('');
      setOpen(false);
      clearError();
      setSuccess('Branch created successfully');
      dispatch(currentEnvironmentChanged(branchName));
    } catch (e) {
      setError(e);
    }
  };

  const handleOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      setBranchInput('');
    }
  };

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent
        className="rounded-xl shadow-2xl border border-gray-200 bg-white dark:bg-gray-900 p-4 w-96"
        align="start"
        sideOffset={4}
      >
        <div className="space-y-1 mb-1">
          <div className="text-lg font-medium text-gray-900 dark:text-gray-100">
            Create Branch
          </div>
          <MoreInfo href="https://www.flipt.io/docs/v2/concepts#branches">
            Learn more about branches
          </MoreInfo>
        </div>
        <div className="mt-6">
          <Input
            type="text"
            placeholder="New branch name"
            value={branchInput}
            onChange={(e) => setBranchInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleCreateBranch();
              if (e.key === 'Escape') handleOpenChange(false);
            }}
            disabled={isCreatingBranch}
            className="mb-4 px-4 py-2 border-2 border-gray-200 rounded-lg focus:border-primary focus:ring-2 focus:ring-primary/20 text-base"
          />
          <div className="flex gap-2 justify-end">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => handleOpenChange(false)}
              type="button"
              className="font-semibold"
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={handleCreateBranch}
              disabled={isCreatingBranch || !branchInput.trim()}
              type="button"
              className="font-semibold"
            >
              Create
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
