import { FilterIcon } from 'lucide-react';
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
    return <>{inTimezone(constraint.value)}</>;
  }

  if (isArrayValue) {
    return <ConstraintArrayValue value={constraint.value} />;
  }

  return <div className="whitespace-normal">{constraint.value}</div>;
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
              <span className="font-medium text-violet-500">
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
            <h3 className="font-medium leading-6 text-gray-900">Constraints</h3>
            <p className="mt-1 text-sm text-gray-500">
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
            <table className="min-w-full divide-y divide-gray-300">
              <thead>
                <tr>
                  <th
                    scope="col"
                    className="pb-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6"
                  >
                    Property
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 sm:table-cell"
                  >
                    Type
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 lg:table-cell"
                  >
                    Operator
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 lg:table-cell"
                  >
                    Value
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 lg:table-cell"
                  >
                    Description
                  </th>
                  <th scope="col" className="relative pb-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {constraints.map((constraint: IConstraint, index: number) => (
                  <tr key={index}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm text-gray-600 sm:pl-6">
                      {constraint.property}
                    </td>
                    <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 sm:table-cell">
                      {constraintTypeToLabel(constraint.type)}
                    </td>
                    <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                      {ConstraintOperators[constraint.operator]}
                    </td>
                    <td className="hidden whitespace-normal px-3 py-4 text-sm text-gray-500 lg:table-cell">
                      <ConstraintValue constraint={constraint} />
                    </td>
                    <td className="hidden truncate whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                      {constraint.description}
                    </td>
                    <td className="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <>
                        <a
                          href="#"
                          className="pr-2 text-violet-600 hover:text-violet-900"
                          onClick={(e) => {
                            e.preventDefault();
                            setEditingConstraint({
                              ...constraint,
                              index
                            });
                            setShowConstraintForm(true);
                          }}
                        >
                          Edit
                          <span className="sr-only">
                            , {constraint.property}
                          </span>
                        </a>
                        <span aria-hidden="true"> | </span>
                        <a
                          href="#"
                          className="pl-2 text-violet-600 hover:text-violet-900"
                          onClick={(e) => {
                            e.preventDefault();
                            setDeletingConstraint({
                              ...constraint,
                              index
                            });
                            setShowDeleteConstraintModal(true);
                          }}
                        >
                          Delete
                          <span className="sr-only">
                            , {constraint.property}
                          </span>
                        </a>
                      </>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <Well>
              <FilterIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-4">
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
