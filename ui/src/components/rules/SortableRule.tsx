import { useSortable } from '@dnd-kit/sortable';
import { IEvaluatable } from '~/types/Evaluatable';
import { INamespace } from '~/types/Namespace';
import Rule from './Rule';

type SortableRuleProps = {
  namespace: INamespace;
  rule: IEvaluatable;
  onEdit: () => void;
  onDelete: () => void;
  readOnly?: boolean;
};

export default function SortableRule(props: SortableRuleProps) {
  const { namespace, rule, onEdit, onDelete, readOnly } = props;
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

  const className = isDragging ? 'border-violet-200 cursor-move' : '';

  return (
    <Rule
      key={rule.id}
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      style={style}
      className={className}
      namespace={namespace}
      rule={rule}
      onEdit={onEdit}
      onDelete={onDelete}
      readOnly={readOnly}
    />
  );
}
