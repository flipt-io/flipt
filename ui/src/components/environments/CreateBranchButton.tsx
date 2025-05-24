import { GitBranchPlusIcon } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import * as Yup from 'yup';

import {
  currentEnvironmentChanged,
  useCreateBranchEnvironmentMutation
} from '~/app/environments/environmentsApi';

import MoreInfo from '~/components/MoreInfo';
import { Popover, PopoverContent, PopoverTrigger } from '~/components/Popover';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { useError } from '~/data/hooks/error';
import { useAppDispatch } from '~/data/hooks/store';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation } from '~/data/validations';

export function CreateBranchButton({
  baseEnvironment
}: {
  baseEnvironment: any;
}) {
  const [showBranchPopover, setShowBranchPopover] = useState(false);
  const [tooltipOpen, setTooltipOpen] = useState(false);

  const [branchInput, setBranchInput] = useState('');

  const [createBranch, { isLoading: isCreatingBranch }] =
    useCreateBranchEnvironmentMutation();
  const branchInputRef = useRef<HTMLInputElement>(null);

  const { setSuccess } = useSuccess();
  const { setError, clearError } = useError();
  const dispatch = useAppDispatch();

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

  useEffect(() => {
    if (showBranchPopover && branchInputRef.current) {
      branchInputRef.current.focus();
    }
  }, [showBranchPopover]);

  if (!baseEnvironment) return null;

  const handleCreateBranch = async () => {
    const branchName = branchInput.trim();
    if (!branchName) return;
    try {
      await branchValidationSchema.validate({ branchName });
      await createBranch({
        baseEnvironmentKey: baseEnvironment.key,
        environmentKey: branchName
      }).unwrap();
      setBranchInput('');
      setShowBranchPopover(false);
      clearError();
      setSuccess('Branch created successfully');
      dispatch(currentEnvironmentChanged({ key: branchName }));
    } catch (e) {
      setError(e);
    }
  };

  return (
    <Popover open={showBranchPopover} onOpenChange={setShowBranchPopover}>
      <Tooltip
        open={!showBranchPopover && tooltipOpen}
        onOpenChange={setTooltipOpen}
      >
        <TooltipTrigger asChild>
          <PopoverTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="ml-1 p-1 rounded focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              tabIndex={-1}
              type="button"
              onMouseEnter={() => setTooltipOpen(true)}
              onMouseLeave={() => setTooltipOpen(false)}
              onFocus={() => setTooltipOpen(true)}
              onBlur={() => setTooltipOpen(false)}
              onClick={() => setTooltipOpen(false)}
              data-testid="create-branch-button"
            >
              <GitBranchPlusIcon className="w-4 h-4" />
            </Button>
          </PopoverTrigger>
        </TooltipTrigger>
        <TooltipContent>
          Create a new branch from this environment
        </TooltipContent>
      </Tooltip>
      <PopoverContent
        className="rounded-xl shadow-2xl border border-gray-200 bg-white dark:bg-gray-900 p-4 w-96"
        align="start"
        sideOffset={4}
        onOpenAutoFocus={(e) => e.preventDefault()}
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
            ref={branchInputRef}
            type="text"
            placeholder="New branch name"
            value={branchInput}
            onChange={(e) => setBranchInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleCreateBranch();
              if (e.key === 'Escape') setShowBranchPopover(false);
            }}
            disabled={isCreatingBranch}
            className="mb-4 px-4 py-2 border-2 border-gray-200 rounded-lg focus:border-primary focus:ring-2 focus:ring-primary/20 text-base"
          />
          <div className="flex gap-2 justify-end">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setShowBranchPopover(false)}
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
