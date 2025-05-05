import {
  CalendarIcon,
  FilterIcon,
  HashIcon,
  IdCardIcon,
  Text,
  ToggleLeftIcon,
  XIcon
} from 'lucide-react';
import { useContext, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { selectViewMode } from '~/app/preferences/preferencesSlice';

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
import { ViewMode } from '~/types/Preferences';

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
        <span className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm text-gray-900 dark:text-white">
          {formattedDate}
        </span>
      );
    } catch (err) {
      // Show the raw value with an error indication
      return (
        <span className="bg-red-100 dark:bg-red-900 px-2 py-1 rounded text-sm text-red-900 dark:text-red-100">
          {constraint.value || '(invalid date)'}
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
      <span
        className={`px-2 py-1 rounded-md text-xs font-medium ${
          boolValue
            ? 'bg-emerald-100 text-emerald-900 dark:bg-emerald-900 dark:text-emerald-100'
            : 'bg-red-100 text-red-900 dark:bg-red-900 dark:text-red-100'
        }`}
      >
        {boolValue ? 'TRUE' : 'FALSE'}
      </span>
    );
  }

  return (
    <span className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm text-gray-900 dark:text-white break-words max-w-full">
      {constraint.value}
    </span>
  );
}

function getTypeIcon(type: ConstraintType) {
  switch (type) {
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
      return <FilterIcon className="h-4 w-4" />;
  }
}

function getTypeColor(type: ConstraintType) {
  switch (type) {
    case ConstraintType.STRING:
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100';
    case ConstraintType.NUMBER:
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100';
    case ConstraintType.BOOLEAN:
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100';
    case ConstraintType.DATETIME:
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-100';
    case ConstraintType.ENTITY_ID:
      return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-100';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-100';
  }
}

function ConstraintCard({
  constraint,
  onEdit,
  onDelete
}: {
  constraint: IConstraint;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="relative flex flex-col rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-950 overflow-hidden shadow-sm hover:shadow-md">
      <div className="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 p-3">
        <div className="flex items-center space-x-2">
          <span className={`p-1.5 rounded-md ${getTypeColor(constraint.type)}`}>
            {getTypeIcon(constraint.type)}
          </span>
          <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {constraintTypeToLabel(constraint.type)}
          </h3>
        </div>
        <button
          onClick={(e) => {
            e.preventDefault();
            onDelete();
          }}
          className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
        >
          <XIcon className="h-4 w-4" />
        </button>
      </div>

      <div className="flex-1 p-4 space-y-3" onClick={onEdit}>
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
            Property
          </span>
          <code className="text-sm font-mono text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {constraint.property}
          </code>
        </div>

        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
            Operator
          </span>
          <span className="text-sm font-medium text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {ConstraintOperators[constraint.operator] || constraint.operator}
          </span>
        </div>

        {!NoValueOperators.includes(constraint.operator) && (
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
              Value
            </span>
            <div className="flex">
              <ConstraintValue constraint={constraint} />
            </div>
          </div>
        )}

        {constraint.description && (
          <div className="pt-2 mt-2 border-t border-gray-100 dark:border-gray-800">
            <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400 block mb-1">
              Description
            </span>
            <p className="text-sm text-gray-700 dark:text-gray-300 line-clamp-2">
              {constraint.description}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

function ConstraintTable({
  constraints,
  onEdit,
  onDelete
}: {
  constraints: IConstraint[];
  onEdit: (constraint: IConstraint, index: number) => void;
  onDelete: (constraint: IConstraint, index: number) => void;
}) {
  return (
    <div className="mt-4 overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-800">
          <tr>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Type
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Property
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Operator
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Value
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Description
            </th>
            <th scope="col" className="relative px-3 py-3">
              <span className="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
          {constraints.map((constraint, index) => (
            <tr
              key={index}
              className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
              onClick={() => onEdit(constraint, index)}
            >
              <td className="px-3 py-4 whitespace-nowrap text-sm font-medium">
                <div className="flex items-center">
                  <span
                    className={`p-1 rounded-md mr-2 ${getTypeColor(constraint.type)}`}
                  >
                    {getTypeIcon(constraint.type)}
                  </span>
                  {constraintTypeToLabel(constraint.type)}
                </div>
              </td>
              <td className="px-3 py-4 whitespace-nowrap text-sm">
                <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                  {constraint.property}
                </code>
              </td>
              <td className="px-3 py-4 whitespace-nowrap text-sm">
                {ConstraintOperators[constraint.operator] ||
                  constraint.operator}
              </td>
              <td className="px-3 py-4 whitespace-nowrap text-sm">
                {!NoValueOperators.includes(constraint.operator) ? (
                  <ConstraintValue constraint={constraint} />
                ) : (
                  <span className="text-gray-400 dark:text-gray-500">—</span>
                )}
              </td>
              <td className="px-3 py-4 text-sm text-gray-700 dark:text-gray-300">
                {constraint.description ? (
                  <p className="line-clamp-1">{constraint.description}</p>
                ) : (
                  <span className="text-gray-400 dark:text-gray-500">—</span>
                )}
              </td>
              <td className="px-3 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete(constraint, index);
                  }}
                  className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400"
                >
                  <XIcon className="h-4 w-4" />
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
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

  // Get user's view mode preference
  const viewMode = useSelector(selectViewMode);

  // Determine if we should use table view based on preference or count
  const useTableView = useMemo(() => {
    if (viewMode === ViewMode.TABLE) return true;
    if (viewMode === ViewMode.CARDS) return false;
    // Default auto behavior - table for more than 6 items
    return constraints.length > 6;
  }, [constraints.length, viewMode]);

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
            <h3 className="font-medium leading-6 text-gray-900 dark:text-gray-100">
              Constraints
            </h3>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-300">
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
            useTableView ? (
              <ConstraintTable
                constraints={constraints}
                onEdit={(constraint, index) => {
                  setEditingConstraint({
                    ...constraint,
                    index
                  });
                  setShowConstraintForm(true);
                }}
                onDelete={(constraint, index) => {
                  setDeletingConstraint({
                    ...constraint,
                    index
                  });
                  setShowDeleteConstraintModal(true);
                }}
              />
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {constraints.map((constraint: IConstraint, index: number) => (
                  <ConstraintCard
                    key={index}
                    constraint={constraint}
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
            )
          ) : (
            <Well>
              <FilterIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground dark:text-gray-200 mb-4">
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
