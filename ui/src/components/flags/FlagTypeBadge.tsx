import { VariableIcon } from 'lucide-react';
import { ToggleLeftIcon } from 'lucide-react';

import { FlagType } from '~/types/Flag';
import { flagTypeToLabel } from '~/types/Flag';

import { cn } from '~/lib/utils';

type FlagTypeBadgeProps = {
  type: FlagType;
  className?: string;
};

export function FlagTypeBadge({ type, className }: FlagTypeBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md bg-violet-50 px-2 py-1 text-xs font-medium text-violet-700 ring-1 ring-inset ring-violet-700/10 dark:bg-violet-400/10 dark:text-violet-400 gap-1',
        className
      )}
    >
      {type === FlagType.BOOLEAN ? (
        <ToggleLeftIcon className="h-4 w-4" />
      ) : (
        <VariableIcon className="h-4 w-4" />
      )}
      {flagTypeToLabel(type)}
    </span>
  );
}
