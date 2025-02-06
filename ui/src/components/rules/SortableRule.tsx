import { useSortable } from '@dnd-kit/sortable';

import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { ISegment } from '~/types/Segment';
import { IVariant } from '~/types/Variant';

import Rule from './Rule';

type SortableRuleProps = {
  flag: IFlag;
  rule: IRule;
  segments: ISegment[];
  variants: IVariant[];
  onSuccess: () => void;
  onDelete: () => void;
};

export default function SortableRule(props: SortableRuleProps) {
  const { flag, rule, segments, variants, onSuccess, onDelete } = props;
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
      variants={variants}
      onSuccess={onSuccess}
      onDelete={onDelete}
    />
  );
}
