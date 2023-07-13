import { useSortable } from '@dnd-kit/sortable';
import { IEvaluatable } from '~/types/Evaluatable';
import { IFlag } from '~/types/Flag';
import { ISegment } from '~/types/Segment';
import Rule from './Rule';

type SortableRuleProps = {
  flag: IFlag;
  rule: IEvaluatable;
  segments: ISegment[];
  onSuccess: () => void;
  onDelete: () => void;
  readOnly?: boolean;
};

export default function SortableRule(props: SortableRuleProps) {
  const { flag, rule, segments, onSuccess, onDelete, readOnly } = props;
  const {
    isDragging,
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition
  } = useSortable({
    id: rule.id,
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
    <Rule
      key={rule.id}
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      style={style}
      className={className}
      flag={flag}
      rule={rule}
      segments={segments}
      onSuccess={onSuccess}
      onDelete={onDelete}
      readOnly={readOnly}
    />
  );
}
