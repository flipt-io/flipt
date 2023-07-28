import { useSortable } from '@dnd-kit/sortable';
import { IFlag } from '~/types/Flag';
import { IRollout } from '~/types/Rollout';
import { ISegment } from '~/types/Segment';
import Rollout from './Rollout';

type SortableRolloutProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  readOnly?: boolean;
};

export default function SortableRollout(props: SortableRolloutProps) {
  const { flag, rollout, segments, onSuccess, onEdit, onDelete, readOnly } =
    props;
  const {
    isDragging,
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition
  } = useSortable({
    id: rollout.id,
    disabled: readOnly
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
        transition
      }
    : undefined;

  const className = isDragging ? 'border-violet-500 cursor-move' : '';

  return (
    <Rollout
      key={rollout.id}
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      style={style}
      className={className}
      flag={flag}
      rollout={rollout}
      segments={segments}
      onSuccess={onSuccess}
      onEdit={onEdit}
      onDelete={onDelete}
      readOnly={readOnly}
    />
  );
}
