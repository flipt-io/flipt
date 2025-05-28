import { useSortable } from '@dnd-kit/sortable';

import { IRule } from '~/types/Rule';
import { ISegment } from '~/types/Segment';
import { IVariant } from '~/types/Variant';

import { cls } from '~/utils/helpers';

import Rule from './Rule';

type SortableRuleProps = {
  rule: IRule;
  segments: ISegment[];
  index?: number;
  variants: IVariant[];
  onSuccess: () => void;
  onDelete: () => void;
};

export default function SortableRule(props: SortableRuleProps) {
  const { variants, rule, segments, index, onSuccess, onDelete } = props;
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
        'ring-ring ring-1': isDragging
      })}
      variants={variants}
      rule={rule}
      segments={segments}
      index={index}
      onSuccess={onSuccess}
      onDelete={onDelete}
    />
  );
}
