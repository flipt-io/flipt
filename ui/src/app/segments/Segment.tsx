import { CalendarIcon } from '@heroicons/react/20/solid';
import { formatDistanceToNowStrict, parseISO } from 'date-fns';
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import DeletePanel from '~/components/DeletePanel';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import { DeleteButton } from '~/components/forms/buttons/DeleteButton';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import MoreInfo from '~/components/MoreInfo';
import ConstraintForm from '~/components/segments/ConstraintForm';
import SegmentForm from '~/components/segments/SegmentForm';
import Slideover from '~/components/Slideover';
import { deleteConstraint, deleteSegment, getSegment } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useTimezone } from '~/data/hooks/timezone';
import {
  ComparisonType,
  ConstraintOperators,
  IConstraint
} from '~/types/Constraint';
import { ISegment } from '~/types/Segment';

export default function Segment() {
  let { segmentKey } = useParams();
  const { inTimezone } = useTimezone();

  const [segment, setSegment] = useState<ISegment | null>(null);
  const [segmentVersion, setSegmentVersion] = useState(0);

  const [showConstraintForm, setShowConstraintForm] = useState<boolean>(false);
  const [editingConstraint, setEditingConstraint] =
    useState<IConstraint | null>(null);
  const [showDeleteConstraintModal, setShowDeleteConstraintModal] =
    useState<boolean>(false);
  const [deletingConstraint, setDeletingConstraint] =
    useState<IConstraint | null>(null);
  const [showDeleteSegmentModal, setShowDeleteSegmentModal] =
    useState<boolean>(false);

  const { setError, clearError } = useError();
  const navigate = useNavigate();

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const incrementSegmentVersion = () => {
    setSegmentVersion(segmentVersion + 1);
  };

  useEffect(() => {
    if (!segmentKey) return;

    getSegment(namespace.key, segmentKey)
      .then((segment: ISegment) => {
        setSegment(segment);
        clearError();
      })
      .catch((err) => {
        setError(err);
      });
  }, [segmentVersion, namespace.key, segmentKey, clearError, setError]);

  const constraintTypeToLabel = (t: string) =>
    ComparisonType[t as keyof typeof ComparisonType];

  const constraintOperatorToLabel = (o: string) => ConstraintOperators[o];

  const constraintFormRef = useRef(null);

  if (!segment) return <Loading />;

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
            incrementSegmentVersion();
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
          handleDelete={
            () =>
              deleteConstraint(
                namespace.key,
                segment.key,
                deletingConstraint?.id ?? ''
              ) // TODO: Determine impact of blank ID param
          }
          onSuccess={() => {
            incrementSegmentVersion();
          }}
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
          handleDelete={() => deleteSegment(namespace.key, segment.key)}
          onSuccess={() => {
            navigate(`/namespaces/${namespace.key}/segments`);
          }}
        />
      </Modal>

      {/* segment header / delete button */}
      <div className="flex items-center justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
            {segment.name}
          </h2>
          <div className="mt-1 flex flex-col sm:mt-0 sm:flex-row sm:flex-wrap sm:space-x-6">
            <div
              title={inTimezone(segment.createdAt)}
              className="mt-2 flex items-center text-sm text-gray-500"
            >
              <CalendarIcon
                className="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400"
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
          <DeleteButton
            disabled={readOnly}
            title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
            onClick={() => setShowDeleteSegmentModal(true)}
          />
        </div>
      </div>

      <div className="flex flex-col">
        {/* segment details */}
        <div className="my-10">
          <div className="md:grid md:grid-cols-3 md:gap-6">
            <div className="md:col-span-1">
              <h3 className="text-lg font-medium leading-6 text-gray-900">
                Details
              </h3>
              <p className="mt-1 text-sm text-gray-500">
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
              <SegmentForm
                segment={segment}
                segmentChanged={incrementSegmentVersion}
              />
            </div>
          </div>
        </div>

        {/* constraints */}
        <div className="mt-10">
          <div>
            <div className="sm:flex sm:items-center">
              <div className="sm:flex-auto">
                <h1 className="text-lg font-medium leading-6 text-gray-900">
                  Constraints
                </h1>
                <p className="mt-1 text-sm text-gray-500">
                  Determine if a request matches a segment
                </p>
              </div>
              {segment.constraints && segment.constraints.length > 0 && (
                <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
                  <Button
                    primary
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
                <table className=" min-w-full divide-y divide-gray-300">
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
                        <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm text-gray-600 sm:pl-6">
                          {constraint.property}
                        </td>
                        <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 sm:table-cell">
                          {constraintTypeToLabel(constraint.type)}
                        </td>
                        <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                          {constraintOperatorToLabel(constraint.operator)}
                        </td>
                        <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                          {constraintTypeToLabel(constraint.type) ===
                            ComparisonType.DATETIME_COMPARISON_TYPE &&
                          constraint.value !== undefined
                            ? inTimezone(constraint.value)
                            : constraint.value}
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
        </div>
      </div>
    </>
  );
}
