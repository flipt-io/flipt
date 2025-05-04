import { 
  BracesIcon, 
  BracketsIcon,
  CalendarIcon,
  FilterIcon, 
  HashIcon, 
  IdCardIcon,
  Text,
  ToggleLeftIcon,
  XIcon 
} from 'lucide-react';
import { useContext, useMemo, useRef, useState } from 'react';

import { Button, ButtonWithPlus } from '~/components/Button';
import Chips from '~/components/Chips';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import DeletePanel from '~/components/panels/DeletePanel';
import ConstraintForm from '~/components/segments/ConstraintForm';
import { SegmentFormContext } from '~/components/segments/SegmentFormContext';

import {
  ConstraintOperators,
  ConstraintType,
  IConstraint,
  NoValueOperators,
  constraintTypeToLabel
} from '~/types/Constraint';

import { useTimezone } from '~/data/hooks/timezone';

function ConstraintArrayValue({ value }: { value: string | undefined }) {
  const items: string[] | number[] = useMemo(() => {
    try {
      return JSON.parse(value || '[]');
    } catch (err) {
      return [];
    }
  }, [value]);

  return <Chips values={items} maxItemCount={5} />;
}

function ConstraintValue({ constraint }: { constraint: IConstraint }) {
  const { inTimezone } = useTimezone();
  const isArrayValue = ['isoneof', 'isnotoneof'].includes(constraint.operator);

  if (
    constraint.type === ConstraintType.DATETIME &&
    constraint.value !== undefined
  ) {
    try {
      // Attempt to format the date - if it fails, show a fallback
      const formattedDate = inTimezone(constraint.value);
      return (
        <span className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-gray-900 dark:text-white">
          {formattedDate}
        </span>
      );
    } catch (err) {
      // Show the raw value with an error indication
      return (
        <span className="bg-red-100 dark:bg-red-900 px-2 py-1 rounded text-red-900 dark:text-red-100">
          {constraint.value || "(invalid date)"}
        </span>
      );
    }
  }

  if (isArrayValue) {
    return <ConstraintArrayValue value={constraint.value} />;
  }

  if (constraint.type === ConstraintType.BOOLEAN) {
    const boolValue = constraint.value?.toLowerCase() === 'true';
    return (
      <span className={`px-2 py-1 rounded-md text-xs font-medium ${
        boolValue 
          ? 'bg-emerald-100 text-emerald-900 dark:bg-emerald-900 dark:text-emerald-100' 
          : 'bg-red-100 text-red-900 dark:bg-red-900 dark:text-red-100'
      }`}>
        {boolValue ? 'TRUE' : 'FALSE'}
      </span>
    );
  }

  return (
    <span className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-gray-900 dark:text-white break-words max-w-full">
      {constraint.value}
    </span>
  );
}

type ConstraintsProps = {
  constraints: IConstraint[];
};

export default function Constraints({ constraints }: ConstraintsProps) {
  const [showConstraintForm, setShowConstraintForm] = useState<boolean>(false);
  const [editingConstraint, setEditingConstraint] = useState<
    (IConstraint & { index: number }) | null
  >(null);
  const [showDeleteConstraintModal, setShowDeleteConstraintModal] =
    useState<boolean>(false);
  const [deletingConstraint, setDeletingConstraint] = useState<
    (IConstraint & { index: number }) | null
  >(null);

  const constraintFormRef = useRef(null);

  const { deleteConstraint } = useContext(SegmentFormContext);

  return (
    <>
      {/* constraint edit form */}
      <Slideover
        open={showConstraintForm}
        setOpen={setShowConstraintForm}
        ref={constraintFormRef}
      >
        <ConstraintForm
          ref={constraintFormRef}
          constraint={editingConstraint || null}
          setOpen={setShowConstraintForm}
          onSuccess={() => {
            setShowConstraintForm(false);
          }}
        />
      </Slideover>

      {/* constraint delete modal */}
      <Modal
        open={showDeleteConstraintModal}
        setOpen={setShowDeleteConstraintModal}
      >
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the constraint for{' '}
              <span className="font-medium text-violet-500 dark:text-violet-400">
                {deletingConstraint?.property}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Constraint"
          setOpen={setShowDeleteConstraintModal}
          handleDelete={() => {
            try {
              deleteConstraint(deletingConstraint!);
            } catch (e) {
              return Promise.reject(e);
            }
            setDeletingConstraint(null);
            return Promise.resolve();
          }}
        />
      </Modal>

      {/* constraints */}
      <div className="mt-2 min-w-full">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
          <h3 className="font-medium leading-6 text-gray-900 dark:text-white">Constraints</h3>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              Determine if a request matches a segment.
            </p>
          </div>
          {constraints && constraints.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <ButtonWithPlus
                variant="primary"
                type="button"
                onClick={() => {
                  setEditingConstraint(null);
                  setShowConstraintForm(true);
                }}
              >
                New Constraint
              </ButtonWithPlus>
            </div>
          )}
        </div>
        <div className="mt-10">
          {constraints && constraints.length > 0 ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {constraints.map((constraint: IConstraint, index: number) => (
                <ConstraintCard 
                  key={index}
                  constraint={constraint}
                  index={index}
                  onEdit={() => {
                    setEditingConstraint({
                      ...constraint,
                      index
                    });
                    setShowConstraintForm(true);
                  }}
                  onDelete={() => {
                    setDeletingConstraint({
                      ...constraint,
                      index
                    });
                    setShowDeleteConstraintModal(true);
                  }}
                />
              ))}
            </div>
          ) : (
            <Well>
              <FilterIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground dark:text-white mb-4">
                No Constraints Yet
              </h3>
              <Button
                variant="primary"
                aria-label="New Constraint"
                onClick={(e) => {
                  e.preventDefault();
                  setEditingConstraint(null);
                  setShowConstraintForm(true);
                }}
              >
                Create Constraint
              </Button>
            </Well>
          )}
        </div>
      </div>
    </>
  );
}

function ConstraintCard({ constraint, index, onEdit, onDelete }: { 
  constraint: IConstraint, 
  index: number,
  onEdit: () => void, 
  onDelete: () => void 
}) {
  const isArrayValue = ['isoneof', 'isnotoneof'].includes(constraint.operator);
  
  const getTypeIcon = (type: ConstraintType) => {
    switch(type) {
      case ConstraintType.STRING:
        return <Text className="h-4 w-4" />;
      case ConstraintType.NUMBER:
        return <HashIcon className="h-4 w-4" />;
      case ConstraintType.BOOLEAN:
        return <ToggleLeftIcon className="h-4 w-4" />;
      case ConstraintType.DATETIME:
        return <CalendarIcon className="h-4 w-4" />;
      case ConstraintType.ENTITY_ID:
        return <IdCardIcon className="h-4 w-4" />;
      default:
        return <BracketsIcon className="h-4 w-4" />;
    }
  };
  
  const getTypeColor = (type: ConstraintType) => {
    switch(type) {
      case ConstraintType.STRING:
        return 'bg-blue-100 text-blue-900 dark:bg-blue-900 dark:text-blue-100';
      case ConstraintType.NUMBER:
        return 'bg-emerald-100 text-emerald-900 dark:bg-emerald-900 dark:text-emerald-100';
      case ConstraintType.BOOLEAN:
        return 'bg-purple-100 text-purple-900 dark:bg-purple-900 dark:text-purple-100';
      case ConstraintType.DATETIME:
        return 'bg-amber-100 text-amber-900 dark:bg-amber-900 dark:text-amber-100';
      case ConstraintType.ENTITY_ID:
        return 'bg-pink-100 text-pink-900 dark:bg-pink-900 dark:text-pink-100';
      default:
        return 'bg-gray-100 text-gray-900 dark:bg-gray-700 dark:text-gray-100';
    }
  };
  
  const typeColor = getTypeColor(constraint.type);
  const typeIcon = getTypeIcon(constraint.type);
  const typeLabel = constraintTypeToLabel(constraint.type);
  
  return (
    <div 
      className="group rounded-lg border border-gray-200 dark:border-gray-800 text-left text-sm transition-all hover:bg-gray-50 dark:hover:bg-gray-800 h-full flex flex-col cursor-pointer relative overflow-hidden shadow-sm hover:shadow"
      onClick={(e) => {
        e.preventDefault();
        onEdit();
      }}
    >
      <div className="flex flex-col p-4 h-full">
        {/* Header with type icon and delete button */}
        <div className="flex items-center justify-between mb-5">
          <span className={`rounded-md p-1.5 flex items-center space-x-2 justify-center ${typeColor}`} title={typeLabel}>
            {typeIcon}
            <span className="text-xs">{typeLabel}</span>
          </span>
          
          <div className="flex items-center gap-2">
            <button
              onClick={(e) => {
                e.stopPropagation(); // Prevent card click from triggering
                onDelete();
              }}
              className="p-1.5 rounded-full text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100"
              aria-label="Delete constraint"
            >
              <XIcon className="h-4 w-4" />
            </button>
          </div>
        </div>
        
        {/* Property with type indication */}
        <div className="flex flex-col min-w-0 flex-1">
          <div className="flex flex-col gap-4">
            {/* Property row */}
            <div className="flex items-baseline">
              <span className="text-gray-500 dark:text-gray-400 text-sm font-medium uppercase w-24">
                PROPERTY:
              </span>
              <code className="text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded font-mono text-sm">
                {constraint.property}
              </code>
            </div>
            
            {/* Operator row */}
            <div className="flex items-baseline">
              <span className="text-gray-500 dark:text-gray-400 text-sm font-medium uppercase w-24">
                OPERATOR:
              </span>
              <span className="text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm">
                {ConstraintOperators[constraint.operator]}
              </span>
            </div>
            
            {/* Only show value if the operator requires a value */}
            {!NoValueOperators.includes(constraint.operator) && (
              <div className="flex items-baseline">
                <span className="text-gray-500 dark:text-gray-400 text-sm font-medium uppercase w-24">
                  VALUE:
                </span>
                <div className="text-gray-900 dark:text-white">
                  <ConstraintValue constraint={constraint} />
                </div>
              </div>
            )}
          </div>
          
          {/* Description (if present) */}
          {constraint.description && (
            <div className="mt-4 pt-3 border-t border-gray-100 dark:border-gray-800">
              <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                {constraint.description}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
