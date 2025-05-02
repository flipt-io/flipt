import { useSortable } from '@dnd-kit/sortable';

import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { ISegment } from '~/types/Segment';

import { cls } from '~/utils/helpers';

import Rule from './Rule';

type SortableRuleProps = {
  flag: IFlag;
  rule: IRule;
  segments: ISegment[];
  index?: number;
  onSuccess: () => void;
  onDelete: () => void;
};

export default function SortableRule(props: SortableRuleProps) {
  const { flag, rule, segments, index, onSuccess, onDelete } = props;
  const {
    isDragging,
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition
  } = useSortable({
    id: rule.id!
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
        transition
      }
    : undefined;

  return (
    <Rule
      key={rule.id}
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      style={style}
      className={cls({
        'border-violet-500': isDragging
      })}
      flag={flag}
      rule={rule}
      segments={segments}
      index={index}
      onSuccess={onSuccess}
      onDelete={onDelete}
    />
  );
}
