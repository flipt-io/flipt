import { FilterIcon } from 'lucide-react';
import { useContext, useRef, useState } from 'react';

import { Button, ButtonWithPlus } from '~/components/Button';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import ConstraintTable from '~/components/constraints/ConstraintTable';
import DeletePanel from '~/components/panels/DeletePanel';
import ConstraintForm from '~/components/segments/ConstraintForm';
import { SegmentFormContext } from '~/components/segments/SegmentFormContext';

import { IConstraint } from '~/types/Constraint';

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
