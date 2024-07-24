import {
  CalendarIcon,
  DocumentDuplicateIcon,
  TrashIcon
} from '@heroicons/react/24/outline';
import { formatDistanceToNowStrict, parseISO } from 'date-fns';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import {
  useCopySegmentMutation,
  useDeleteConstraintMutation,
  useDeleteSegmentMutation,
  useGetSegmentQuery
} from '~/app/segments/segmentsApi';
import Chips from '~/components/Chips';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import Dropdown from '~/components/forms/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import MoreInfo from '~/components/MoreInfo';
import CopyToNamespacePanel from '~/components/panels/CopyToNamespacePanel';
import DeletePanel from '~/components/panels/DeletePanel';
import ConstraintForm from '~/components/segments/ConstraintForm';
import SegmentForm from '~/components/segments/SegmentForm';
import Slideover from '~/components/Slideover';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { useTimezone } from '~/data/hooks/timezone';
import {
  ConstraintOperators,
  ConstraintType,
  constraintTypeToLabel,
  IConstraint
} from '~/types/Constraint';

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

export default function Segment() {
  let { segmentKey } = useParams();
  const { inTimezone } = useTimezone();

  const [showConstraintForm, setShowConstraintForm] = useState<boolean>(false);
  const [editingConstraint, setEditingConstraint] =
    useState<IConstraint | null>(null);
  const [showDeleteConstraintModal, setShowDeleteConstraintModal] =
    useState<boolean>(false);
  const [deletingConstraint, setDeletingConstraint] =
    useState<IConstraint | null>(null);
  const [showDeleteSegmentModal, setShowDeleteSegmentModal] =
    useState<boolean>(false);
  const [showCopySegmentModal, setShowCopySegmentModal] =
    useState<boolean>(false);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const namespaces = useSelector(selectNamespaces);
  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const {
    data: segment,
    error,
    isLoading,
    isError
  } = useGetSegmentQuery({
    namespaceKey: namespace.key,
    segmentKey: segmentKey || ''
  });

  const [deleteSegment] = useDeleteSegmentMutation();
  const [deleteSegmentConstraint] = useDeleteConstraintMutation();
  const [copySegment] = useCopySegmentMutation();

  const constraintFormRef = useRef(null);

  useEffect(() => {
    if (isError) {
      setError(error);
    }
  }, [error, isError, setError]);

  if (isLoading || !segment) {
    return <Loading />;
  }

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
          segmentKey={segment.key}
          constraint={editingConstraint || undefined}
          setOpen={setShowConstraintForm}
          onSuccess={() => {
            clearError();
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
              <span className="text-violet-500 font-medium">
                {deletingConstraint?.property}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Constraint"
          setOpen={setShowDeleteConstraintModal}
          handleDelete={() =>
            deleteSegmentConstraint({
              namespaceKey: namespace.key,
              segmentKey: segment.key,
              constraintId: deletingConstraint?.id ?? ''
            }).unwrap()
          }
        />
      </Modal>

      {/* segment delete modal */}
      <Modal open={showDeleteSegmentModal} setOpen={setShowDeleteSegmentModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the segment{' '}
              <span className="text-violet-500 font-medium">{segment.key}</span>
              ? This action cannot be undone.
            </>
          }
          panelType="Segment"
          setOpen={setShowDeleteSegmentModal}
          handleDelete={() =>
            deleteSegment({
              namespaceKey: namespace.key,
              segmentKey: segment.key
            }).unwrap()
          }
          onSuccess={() => {
            navigate(`/namespaces/${namespace.key}/segments`);
          }}
        />
      </Modal>

      {/* segment copy modal */}
      <Modal open={showCopySegmentModal} setOpen={setShowCopySegmentModal}>
        <CopyToNamespacePanel
          panelMessage={
            <>
              Copy the segment{' '}
              <span className="text-violet-500 font-medium">{segment.key}</span>{' '}
              to the namespace:
            </>
          }
          panelType="Segment"
          setOpen={setShowCopySegmentModal}
          handleCopy={(namespaceKey: string) =>
            copySegment({
              from: { namespaceKey: namespace.key, segmentKey: segment.key },
              to: { namespaceKey: namespaceKey, segmentKey: segment.key }
            }).unwrap()
          }
          onSuccess={() => {
            clearError();
            setShowCopySegmentModal(false);
            setSuccess('Successfully copied segment');
          }}
        />
      </Modal>

      {/* segment header / delete button */}
      <div className="flex items-center justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-gray-900 text-2xl font-bold leading-7 sm:truncate sm:text-3xl sm:tracking-tight">
            {segment.name}
          </h2>
          <div className="mt-1 flex flex-col sm:mt-0 sm:flex-row sm:flex-wrap sm:space-x-6">
            <div
              title={inTimezone(segment.createdAt)}
              className="text-gray-500 mt-2 flex items-center text-sm"
            >
              <CalendarIcon
                className="text-gray-400 mr-1.5 h-5 w-5 flex-shrink-0"
                aria-hidden="true"
              />
              Created{' '}
              {formatDistanceToNowStrict(parseISO(segment.createdAt), {
                addSuffix: true
              })}
            </div>
          </div>
        </div>
        <div className="flex flex-none">
          <Dropdown
            label="Actions"
            actions={[
              {
                id: 'copy',
                label: 'Copy to Namespace',
                disabled: readOnly || namespaces.length < 2,
                onClick: () => setShowCopySegmentModal(true),
                icon: DocumentDuplicateIcon
              },
              {
                id: 'delete',
                label: 'Delete',
                disabled: readOnly,
                onClick: () => setShowDeleteSegmentModal(true),
                icon: TrashIcon,
                activeClassName: readOnly ? 'text-red-500' : 'text-red-700',
                inActiveClassName: readOnly ? 'text-red-400' : 'text-red-600'
              }
            ]}
          />
        </div>
      </div>

      <div className="flex flex-col">
        {/* segment details */}
        <div className="mb-5 mt-10">
          <div className="md:grid md:grid-cols-3 md:gap-6">
            <div className="md:col-span-1">
              <p className="text-gray-500 mt-1 text-sm">
                Basic information about the segment
              </p>
              <MoreInfo
                className="mt-5"
                href="https://www.flipt.io/docs/concepts#segments"
              >
                Learn more about segments
              </MoreInfo>
            </div>
            <div className="mt-5 md:col-span-2 md:mt-0">
              <SegmentForm segment={segment} />
            </div>
          </div>
        </div>

        {/* constraints */}
        <div className="mt-10">
          <div>
            <div className="sm:flex sm:items-center">
              <div className="sm:flex-auto">
                <h3 className="text-gray-900 font-medium leading-6">
                  Constraints
                </h3>
                <p className="text-gray-500 mt-1 text-sm">
                  Determine if a request matches a segment
                </p>
              </div>
              {segment.constraints && segment.constraints.length > 0 && (
                <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
                  <Button
                    variant="primary"
                    type="button"
                    disabled={readOnly}
                    title={
                      readOnly ? 'Not allowed in Read-Only mode' : undefined
                    }
                    onClick={() => {
                      setEditingConstraint(null);
                      setShowConstraintForm(true);
                    }}
                  >
                    New Constraint
                  </Button>
                </div>
              )}
            </div>
            <div className="my-10">
              {segment.constraints && segment.constraints.length > 0 ? (
                <table className="min-w-full divide-y divide-gray-300">
                  <thead>
                    <tr>
                      <th
                        scope="col"
                        className="text-gray-900 pb-3.5 pl-4 pr-3 text-left text-sm font-semibold sm:pl-6"
                      >
                        Property
                      </th>
                      <th
                        scope="col"
                        className="text-gray-900 hidden px-3 pb-3.5 text-left text-sm font-semibold sm:table-cell"
                      >
                        Type
                      </th>
                      <th
                        scope="col"
                        className="text-gray-900 hidden px-3 pb-3.5 text-left text-sm font-semibold lg:table-cell"
                      >
                        Operator
                      </th>
                      <th
                        scope="col"
                        className="text-gray-900 hidden px-3 pb-3.5 text-left text-sm font-semibold lg:table-cell"
                      >
                        Value
                      </th>
                      <th
                        scope="col"
                        className="text-gray-900 hidden px-3 pb-3.5 text-left text-sm font-semibold lg:table-cell"
                      >
                        Description
                      </th>
                      <th
                        scope="col"
                        className="relative pb-3.5 pl-3 pr-4 sm:pr-6"
                      >
                        <span className="sr-only">Edit</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {segment.constraints.map((constraint) => (
                      <tr key={constraint.id}>
                        <td className="text-gray-600 whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                          {constraint.property}
                        </td>
                        <td className="text-gray-500 hidden whitespace-nowrap px-3 py-4 text-sm sm:table-cell">
                          {constraintTypeToLabel(constraint.type)}
                        </td>
                        <td className="text-gray-500 hidden whitespace-nowrap px-3 py-4 text-sm lg:table-cell">
                          {ConstraintOperators[constraint.operator]}
                        </td>
                        <td className="text-gray-500 hidden whitespace-normal px-3 py-4 text-sm lg:table-cell">
                          <ConstraintValue constraint={constraint} />
                        </td>
                        <td className="text-gray-500 hidden truncate whitespace-nowrap px-3 py-4 text-sm lg:table-cell">
                          {constraint.description}
                        </td>
                        <td className="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                          {!readOnly && (
                            <>
                              <a
                                href="#"
                                className="text-violet-600 pr-2 hover:text-violet-900"
                                onClick={(e) => {
                                  e.preventDefault();
                                  setEditingConstraint(constraint);
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
                                className="text-violet-600 pl-2 hover:text-violet-900"
                                onClick={(e) => {
                                  e.preventDefault();
                                  setDeletingConstraint(constraint);
                                  setShowDeleteConstraintModal(true);
                                }}
                              >
                                Delete
                                <span className="sr-only">
                                  , {constraint.property}
                                </span>
                              </a>
                            </>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <EmptyState
                  text="New Constraint"
                  disabled={readOnly}
                  onClick={() => {
                    setEditingConstraint(null);
                    setShowConstraintForm(true);
                  }}
                />
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
