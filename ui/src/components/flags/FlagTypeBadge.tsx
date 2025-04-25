import { VariableIcon } from 'lucide-react';
import { ToggleLeftIcon } from 'lucide-react';

import { Badge } from '~/components/Badge';

import { FlagType } from '~/types/Flag';
import { flagTypeToLabel } from '~/types/Flag';

import { cls } from '~/utils/helpers';

type FlagTypeBadgeProps = {
  type: FlagType;
  className?: string;
};

export function FlagTypeBadge({ type, className }: FlagTypeBadgeProps) {
  return (
    <Badge
      variant="outlinemuted"
      className={cls(className, 'flex items-center gap-1')}
    >
      {type === FlagType.BOOLEAN ? (
        <ToggleLeftIcon className="h-4 w-4" />
      ) : (
        <VariableIcon className="h-4 w-4" />
      )}
      {flagTypeToLabel(type)}
    </Badge>
  );
}
