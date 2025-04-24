import { AsteriskIcon, SigmaIcon } from 'lucide-react';

import { Badge } from '~/components/Badge';

import { SegmentMatchType, segmentMatchTypeToLabel } from '~/types/Segment';

import { cls } from '~/utils/helpers';

export function SegmentMatchTypeBadge({
  type,
  className
}: {
  type: SegmentMatchType;
  className?: string;
}) {
  return (
    <Badge
      variant="outlinemuted"
      className={cls('flex items-center gap-1', className)}
    >
      {type === SegmentMatchType.ALL ? (
        <SigmaIcon className="h-4 w-4" />
      ) : (
        <AsteriskIcon className="h-4 w-4" />
      )}
      {segmentMatchTypeToLabel(type)}
    </Badge>
  );
}
