import { CalendarIcon, FilesIcon, Trash2Icon } from 'lucide-react';
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
import { PageHeader } from '~/components/ui/page';

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
              <span className="font-medium text-violet-500">
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
              <span className="font-medium text-violet-500">{segment.key}</span>
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
              <span className="font-medium text-violet-500">{segment.key}</span>{' '}
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

      <PageHeader title={segment.name}>
        <Dropdown
          label="Actions"
          actions={[
            {
              id: 'segment-copy',
              label: 'Copy to Namespace',
              disabled: readOnly || namespaces.length < 2,
              onClick: () => setShowCopySegmentModal(true),
              icon: FilesIcon
            },
            {
              id: 'segement-delete',
              label: 'Delete',
              disabled: readOnly,
              onClick: () => setShowDeleteSegmentModal(true),
              icon: Trash2Icon,
              variant: 'destructive'
            }
          ]}
        />
      </PageHeader>

      <div className="mb-8 space-y-4">
        <div className="flex items-center text-sm text-gray-500">
          <CalendarIcon className="mr-1.5 h-5 w-5 text-gray-400" />
          Created{' '}
          {formatDistanceToNowStrict(parseISO(segment.createdAt), {
            addSuffix: true
          })}
        </div>

        <MoreInfo href="https://www.flipt.io/docs/concepts#segments">
          Learn more about segments
        </MoreInfo>
      </div>

      <div className="mb-8 max-w-screen-md">
        <SegmentForm segment={segment} />
      </div>

      <div>
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h3 className="font-medium leading-6 text-gray-900">Constraints</h3>
            <p className="mt-1 text-sm text-gray-500">
              Determine if a request matches a segment
            </p>
          </div>
          {segment.constraints && segment.constraints.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                variant="primary"
                type="button"
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
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
                {segment.constraints.map((constraint) => (
                  <tr key={constraint.id}>
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
                      {!readOnly && (
                        <>
                          <a
                            href="#"
                            className="pr-2 text-violet-600 hover:text-violet-900"
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
                            className="pl-2 text-violet-600 hover:text-violet-900"
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
    </>
  );
}
