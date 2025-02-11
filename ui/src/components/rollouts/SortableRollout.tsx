import { useSortable } from '@dnd-kit/sortable';

import { IFlag } from '~/types/Flag';
import { IRollout } from '~/types/Rollout';
import { ISegment } from '~/types/Segment';

import { cls } from '~/utils/helpers';

import Rollout from './Rollout';

type SortableRolloutProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
};

export default function SortableRollout(props: SortableRolloutProps) {
  const { flag, rollout, segments, onSuccess, onEdit, onDelete } = props;
  const {
    isDragging,
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition
  } = useSortable({
    id: rollout.id!
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
        transition
      }
    : undefined;

  return (
    <Rollout
      key={rollout.id}
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      style={style}
      className={cls({
        'border-violet-500': isDragging
      })}
      flag={flag}
      rollout={rollout}
      segments={segments}
      onSuccess={onSuccess}
      onEdit={onEdit}
      onDelete={onDelete}
    />
  );
}
