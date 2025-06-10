import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import * as Yup from 'yup';

import {
  currentEnvironmentChanged,
  selectAllEnvironments,
  useCreateBranchEnvironmentMutation
} from '~/app/environments/environmentsApi';

import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '~/components/Dialog';
import MoreInfo from '~/components/MoreInfo';
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
    <Dialog onOpenChange={setOpen} open={open}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Branch</DialogTitle>
          <DialogDescription>
            <MoreInfo href="https://docs.flipt.io/v2/concepts#branches">
              Learn more about branches
            </MoreInfo>
          </DialogDescription>
        </DialogHeader>

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
          className="mt-2"
        />
        <DialogFooter>
          <DialogClose>Cancel</DialogClose>
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
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
