import {
  closestCenter,
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy
} from '@dnd-kit/sortable';
import { PlusIcon, StarIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useUpdateFlagMutation } from '~/app/flags/flagsApi';
import {
  useDeleteRolloutMutation,
  useListRolloutsQuery,
  useOrderRolloutsMutation
} from '~/app/flags/rolloutsApi';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import Button from '~/components/forms/buttons/Button';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import EditRolloutForm from '~/components/rollouts/forms/EditRolloutForm';
import RolloutForm from '~/components/rollouts/forms/RolloutForm';
import Rollout from '~/components/rollouts/Rollout';
import SortableRollout from '~/components/rollouts/SortableRollout';
import Slideover from '~/components/Slideover';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IFlag } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';
import { IRollout } from '~/types/Rollout';
import { SegmentOperatorType } from '~/types/Segment';
import Select from '~/components/forms/Select';
import TextButton from '~/components/forms/buttons/TextButton';
import Loading from '~/components/Loading';
import { cls } from '~/utils/helpers';

type RolloutsProps = {
  flag: IFlag;
};

interface DefaultRolloutFormValues {
  defaultValue: string;
}

export function DefaultRollout(props: RolloutsProps) {
  const { flag } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace) as INamespace;
  const readOnly = useSelector(selectReadonly);

  const [updateFlag] = useUpdateFlagMutation();

  const handleSubmit = async (values: DefaultRolloutFormValues) => {
    await updateFlag({
      namespaceKey: namespace.key,
      flagKey: flag.key,
      values: {
        ...flag,
        enabled: values.defaultValue === 'true'
      }
    });
  };

  return (
    <Formik
      enableReinitialize
      initialValues={{
        defaultValue: flag.enabled ? 'true' : 'false'
      }}
      onSubmit={(values: DefaultRolloutFormValues, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess('Successfully updated default rollout');
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
    >
      {(formik) => {
        return (
          <div className="flex flex-col p-2">
            <div className="bg-white border-violet-300 w-full items-center space-y-2 rounded-md border shadow-md shadow-violet-100 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-6 lg:py-2">
              <div className="bg-white border-gray-200 w-full border-b p-2">
                <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
                  <StarIcon className="text-gray-400 hidden h-4 w-4 justify-start hover:text-violet-300 sm:flex" />
                  <h3 className="text-gray-700 text-sm font-normal leading-6">
                    Default Rollout
                  </h3>
                  <span className="hidden h-4 w-4 justify-end sm:flex" />
                </div>
              </div>

              <div className="flex grow flex-col items-center justify-center sm:ml-2">
                <p className="text-gray-600 text-center text-sm font-light">
                  This is the default value that will be returned if no other
                  rules match. It is directly tied to the flag enabled state.
                </p>
              </div>

              <div className="flex w-full flex-1 items-center p-2 text-xs lg:p-0">
                <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
                  <Form className="bg-white flex w-full flex-col overflow-y-scroll">
                    <div className="w-full flex-1">
                      <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                        <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
                          <div>
                            <label
                              htmlFor="defaultValue"
                              className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                            >
                              Value
                            </label>
                          </div>
                          <div>
                            <Select
                              id="defaultValue"
                              name="defaultValue"
                              value={formik.values.defaultValue}
                              options={[
                                { label: 'True', value: 'true' },
                                { label: 'False', value: 'false' }
                              ]}
                              className={cls(
                                'w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle',
                                {
                                  'text-gray-500 bg-gray-100 cursor-not-allowed':
                                    readOnly
                                }
                              )}
                              disabled={readOnly}
                            />
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="flex-shrink-0 py-1">
                      <div className="flex justify-end space-x-3">
                        <TextButton
                          disabled={formik.isSubmitting || readOnly}
                          onClick={() => {
                            formik.resetForm();
                          }}
                        >
                          Reset
                        </TextButton>
                        <TextButton
                          type="submit"
                          className="min-w-[80px]"
                          disabled={
                            !formik.isValid || formik.isSubmitting || readOnly
                          }
                        >
                          {formik.isSubmitting ? (
                            <Loading isPrimary />
                          ) : (
                            'Update'
                          )}
                        </TextButton>
                      </div>
                    </div>
                  </Form>
                </div>
              </div>
            </div>
          </div>
        );
      }}
    </Formik>
  );
}

export default function Rollouts(props: RolloutsProps) {
  const { flag } = props;

  const [activeRollout, setActiveRollout] = useState<IRollout | null>(null);

  const [showRolloutForm, setShowRolloutForm] = useState<boolean>(false);

  const [showEditRolloutForm, setShowEditRolloutForm] =
    useState<boolean>(false);
  const [editingRollout, setEditingRollout] = useState<IRollout | null>(null);

  const [showDeleteRolloutModal, setShowDeleteRolloutModal] =
    useState<boolean>(false);
  const [deletingRollout, setDeletingRollout] = useState<IRollout | null>(null);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const rolloutFormRef = useRef(null);

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);
  const segmentsList = useListSegmentsQuery(namespace.key);
  const segments = useMemo(
    () => segmentsList.data?.segments || [],
    [segmentsList]
  );

  const [deleteRollout] = useDeleteRolloutMutation();

  const rolloutsList = useListRolloutsQuery({
    namespaceKey: namespace.key,
    flagKey: flag.key
  });
  const rolloutsRules = useMemo(
    () => rolloutsList.data?.rules || [],
    [rolloutsList]
  );

  const rollouts = useMemo(() => {
    // Combine both segmentKey and segmentKeys for legacy purposes.
    // TODO(yquansah): Should be removed once there are no more references to `segmentKey`.
    return rolloutsRules.map((rollout) => {
      if (rollout.segment) {
        let segmentKeys: string[] = [];
        if (
          rollout.segment.segmentKeys &&
          rollout.segment.segmentKeys.length > 0
        ) {
          segmentKeys = rollout.segment.segmentKeys;
        } else if (rollout.segment.segmentKey) {
          segmentKeys = [rollout.segment.segmentKey];
        }

        return {
          ...rollout,
          segment: {
            segmentOperator:
              rollout.segment.segmentOperator || SegmentOperatorType.OR,
            segmentKeys,
            value: rollout.segment.value
          }
        };
      }

      return {
        ...rollout
      };
    });
  }, [rolloutsRules]);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

  const [orderRollouts] = useOrderRolloutsMutation();

  const reorderRollouts = (rollouts: IRollout[]) => {
    orderRollouts({
      namespaceKey: namespace.key,
      flagKey: flag.key,
      rolloutIds: rollouts.map((rollout) => rollout.id)
    })
      .then(() => {
        clearError();
        setSuccess('Successfully reordered rollouts');
      })
      .catch((err) => {
        setError(err);
      });
  };

  // disabling eslint due to this being a third-party event type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onDragEnd = (event: { active: any; over: any }) => {
    const { active, over } = event;

    if (active.id !== over.id) {
      const reordered = (function (rollouts: IRollout[]) {
        const oldIndex = rollouts.findIndex(
          (rollout) => rollout.id === active.id
        );
        const newIndex = rollouts.findIndex(
          (rollout) => rollout.id === over.id
        );

        return arrayMove(rollouts, oldIndex, newIndex);
      })(rollouts);

      reorderRollouts(reordered);
    }

    setActiveRollout(null);
  };

  // disabling eslint due to this being a third-party event type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onDragStart = (event: { active: any }) => {
    const { active } = event;
    const rollout = rollouts.find((rollout) => rollout.id === active.id);
    if (rollout) {
      setActiveRollout(rollout);
    }
  };

  return (
    <>
      {/* rollout delete modal */}
      <Modal open={showDeleteRolloutModal} setOpen={setShowDeleteRolloutModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete this rule at
              <span className="text-violet-500 font-medium">
                {' '}
                position {deletingRollout?.rank}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Rollout"
          setOpen={setShowDeleteRolloutModal}
          handleDelete={() =>
            deleteRollout({
              namespaceKey: namespace.key,
              flagKey: flag.key,
              rolloutId: deletingRollout?.id ?? ''
            }).unwrap()
          }
        />
      </Modal>

      {/* rollout create form */}
      <Slideover
        open={showRolloutForm}
        setOpen={setShowRolloutForm}
        ref={rolloutFormRef}
      >
        <RolloutForm
          flagKey={flag.key}
          rank={rollouts.length + 1}
          segments={segments}
          setOpen={setShowRolloutForm}
          onSuccess={() => {
            setShowRolloutForm(false);
          }}
        />
      </Slideover>

      {/* rollout edit form */}
      {editingRollout && (
        <Slideover
          open={showEditRolloutForm}
          setOpen={setShowEditRolloutForm}
          ref={rolloutFormRef}
        >
          <EditRolloutForm
            flagKey={flag.key}
            segments={segments}
            rollout={editingRollout}
            setOpen={setShowEditRolloutForm}
            onSuccess={() => {
              setShowEditRolloutForm(false);
            }}
          />
        </Slideover>
      )}

      {/* rollouts */}
      <div className="mt-10 w-full">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h3 className="text-gray-900 font-medium leading-6">Rollouts</h3>
            <p className="text-gray-500 mt-1 text-sm">
              Return boolean values based on rules you define
            </p>
          </div>
          <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
            <Button
              variant="primary"
              type="button"
              disabled={readOnly}
              title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
              onClick={() => {
                setEditingRollout(null);
                setShowRolloutForm(true);
              }}
            >
              <PlusIcon
                className="text-white -ml-1.5 mr-1 h-5 w-5"
                aria-hidden="true"
              />
              <span>New Rollout</span>
            </Button>
          </div>
        </div>
        <div className="mt-10">
          <div className="flex lg:space-x-5">
            <div className="hidden w-1/4 flex-col space-y-7 pr-3 lg:flex">
              <p className="text-gray-700 text-sm font-light">
                Rollout rules are evaluated in order from{' '}
                <span className="font-semibold">top to bottom</span>. The first
                rule that matches will be applied.
              </p>
              <p className="text-gray-700 text-sm font-light">
                Rollouts can be rearranged by clicking on a rollout header and{' '}
                <span className="font-semibold">dragging and dropping</span> it
                into place.
              </p>
            </div>
            <div className="border-gray-200 pattern-boxes w-full border p-4 pattern-bg-gray-50 pattern-gray-100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-black dark:pattern-gray-900 lg:p-6">
              {rollouts && rollouts.length > 0 && (
                <DndContext
                  sensors={sensors}
                  collisionDetection={closestCenter}
                  onDragStart={onDragStart}
                  onDragEnd={onDragEnd}
                >
                  <SortableContext
                    items={rollouts.map((rollout) => rollout.id)}
                    strategy={verticalListSortingStrategy}
                  >
                    <ul role="list" className="flex-col space-y-6 p-2 md:flex">
                      {rollouts.map((rollout) => (
                        <SortableRollout
                          key={`${rollout.id}-${rollout.updatedAt}`}
                          flag={flag}
                          rollout={rollout}
                          segments={segments}
                          onEdit={() => {
                            setEditingRollout(rollout);
                            setShowEditRolloutForm(true);
                          }}
                          onDelete={() => {
                            setDeletingRollout(rollout);
                            setShowDeleteRolloutModal(true);
                          }}
                          readOnly={readOnly}
                        />
                      ))}
                    </ul>
                  </SortableContext>
                  <DragOverlay>
                    {activeRollout ? (
                      <Rollout
                        flag={flag}
                        rollout={activeRollout}
                        segments={segments}
                      />
                    ) : null}
                  </DragOverlay>
                </DndContext>
              )}
              <DefaultRollout flag={flag} />
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
