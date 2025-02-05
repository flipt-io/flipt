import { useSortable } from '@dnd-kit/sortable';
import { IEvaluatable } from '~/types/Evaluatable';
import { IFlag } from '~/types/Flag';
import { ISegment } from '~/types/Segment';
import Rule from './Rule';
import { IRule } from '~/types/Rule';

type SortableRuleProps = {
  flag: IFlag;
  rule: IRule;
  segments: ISegment[];
  onSuccess: () => void;
  onDelete: () => void;
};

export default function SortableRule(props: SortableRuleProps) {
  const { flag, rule, segments, onSuccess, onDelete } = props;
  const {
    isDragging,
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition
  } = useSortable({
    id: rule.id
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
    />
  );
}
