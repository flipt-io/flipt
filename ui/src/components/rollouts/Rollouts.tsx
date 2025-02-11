import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy
} from '@dnd-kit/sortable';
import { StarIcon } from '@heroicons/react/24/outline';
import { useFormikContext } from 'formik';
import { useContext, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';

import { ButtonWithPlus, TextButton } from '~/components/Button';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import Select from '~/components/forms/Select';
import DeletePanel from '~/components/panels/DeletePanel';
import EditRolloutForm from '~/components/rollouts/EditRolloutForm';
import Rollout from '~/components/rollouts/Rollout';
import RolloutForm from '~/components/rollouts/RolloutForm';
import SortableRollout from '~/components/rollouts/SortableRollout';

import { IFlag } from '~/types/Flag';
import { IRollout } from '~/types/Rollout';
import { SegmentOperatorType } from '~/types/Segment';

import { useError } from '~/data/hooks/error';
import { cls } from '~/utils/helpers';

type RolloutsProps = {
  flag: IFlag;
  rollouts?: IRollout[];
};

export function DefaultRollout() {
  const formik = useFormikContext<IFlag>();

  return (
    <div className="flex flex-col p-2">
      <div className="w-full items-center space-y-2 rounded-md border border-violet-300 bg-background shadow-md shadow-violet-100 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-6 lg:py-2">
        <div className="w-full rounded-t-lg border-b border-gray-200 p-2">
          <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
            <StarIcon className="hidden h-4 w-4 justify-start text-gray-400 hover:text-violet-300 sm:flex" />
            <h3 className="text-sm font-normal leading-6 text-gray-700">
              Default Rollout
            </h3>
            <span className="hidden h-4 w-4 justify-end sm:flex" />
          </div>
        </div>

        <div className="flex grow flex-col items-center justify-center sm:ml-2">
          <p className="text-center text-sm font-light text-gray-600">
            This is the default value that will be returned if no other rules
            match. It is directly tied to the flag enabled state.
          </p>
        </div>

        <div className="flex w-full flex-1 items-center p-2 text-xs lg:p-0">
          <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
            <div className="flex w-full flex-col overflow-y-scroll bg-background">
              {' '}
              <div className="w-full flex-1">
                <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
                    <div>
                      <label
                        htmlFor="defaultValue"
                        className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                      >
                        Value
                      </label>
                    </div>
                    <div>
                      <Select
                        id="defaultValue"
                        name="defaultValue"
                        value={formik.values.enabled ? 'true' : 'false'}
                        onChange={(e) => {
                          formik.setFieldValue(
                            'enabled',
                            e.target.value === 'true'
                          );
                        }}
                        options={[
                          { label: 'True', value: 'true' },
                          { label: 'False', value: 'false' }
                        ]}
                        className={cls(
                          'w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle'
                        )}
                      />
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex-shrink-0 py-1">
                <div className="flex justify-end space-x-2">
                  <TextButton
                    disabled={formik.isSubmitting}
                    onClick={() => {
                      formik.resetForm();
                    }}
                  >
                    Reset
                  </TextButton>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
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

  const { clearError } = useError();

  const rolloutFormRef = useRef(null);

  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);

  const segmentsList = useListSegmentsQuery({
    environmentKey: environment.name,
    namespaceKey: namespace.key
  });
  const segments = useMemo(
    () => segmentsList.data?.segments || [],
    [segmentsList]
  );

  const { setRollouts, createRollout, updateRollout, deleteRollout } =
    useContext(FlagFormContext);

  const rollouts = useMemo(() => {
    return props.rollouts!.map((rollout) => {
      if (rollout.segment) {
        let segments: string[] = [];
        if (rollout.segment.segments && rollout.segment.segments.length > 0) {
          segments = rollout.segment.segments;
        }

        return {
          ...rollout,
          segment: {
            segmentOperator:
              rollout.segment.segmentOperator || SegmentOperatorType.OR,
            segments,
            value: rollout.segment.value
          }
        };
      }

      return {
        ...rollout
      };
    });
  }, [props.rollouts]);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

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

      setRollouts(reordered);
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

  if (segmentsList.isLoading) {
    return <Loading />;
  }

  return (
    <>
      {/* rollout delete modal */}
      <Modal open={showDeleteRolloutModal} setOpen={setShowDeleteRolloutModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete this rule at
              <span className="font-medium text-violet-500">
                {' '}
                position {deletingRollout?.rank}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Rollout"
          setOpen={setShowDeleteRolloutModal}
          handleDelete={() => {
            if (!deletingRollout) {
              return Promise.resolve();
            }
            deleteRollout(deletingRollout);
            return Promise.resolve();
          }}
        />
      </Modal>

      {/* rollout create form */}
      <Slideover
        open={showRolloutForm}
        setOpen={setShowRolloutForm}
        ref={rolloutFormRef}
      >
        <RolloutForm
          rank={(rollouts?.length || 0) + 1}
          segments={segments}
          setOpen={setShowRolloutForm}
          createRollout={createRollout}
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
            segments={segments}
            rollout={editingRollout}
            setOpen={setShowEditRolloutForm}
            updateRollout={updateRollout}
            onSuccess={() => {
              setShowEditRolloutForm(false);
            }}
          />
        </Slideover>
      )}

      {/* rollouts */}
      <div className="mt-2">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <p className="mt-1 text-sm text-gray-500">
              Return boolean values based on rules you define. Rules are
              evaluated in order from top to bottom.
            </p>
            <p className="mt-1 text-sm text-gray-500">
              Rules can be rearranged by clicking on the header and dragging and
              dropping it into place.
            </p>
          </div>
          <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
            <ButtonWithPlus
              variant="primary"
              type="button"
              onClick={() => {
                setEditingRollout(null);
                setShowRolloutForm(true);
              }}
            >
              New Rollout
            </ButtonWithPlus>
          </div>
        </div>
        <div className="mt-10">
          <div className="flex">
            <div className="dark:pattern-bg-solidwhite pattern-boxes w-full border border-gray-200 p-4 pattern-bg-gray-solid50 pattern-gray-solid100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-gray-solid lg:p-6">
              {rollouts && rollouts.length > 0 && (
                <DndContext
                  sensors={sensors}
                  collisionDetection={closestCenter}
                  onDragStart={onDragStart}
                  onDragEnd={onDragEnd}
                >
                  <SortableContext
                    items={rollouts.map((rollout) => rollout.id!)}
                    strategy={verticalListSortingStrategy}
                  >
                    <ul role="list" className="flex-col space-y-6 p-2 md:flex">
                      {rollouts &&
                        rollouts.map((rollout) => (
                          <SortableRollout
                            key={rollout.id}
                            flag={flag}
                            rollout={rollout}
                            segments={segments}
                            onEdit={() => {
                              setEditingRollout(rollout);
                              setShowEditRolloutForm(true);
                            }}
                            onDelete={() => {
                              setActiveRollout(null);
                              setDeletingRollout(rollout);
                              setShowDeleteRolloutModal(true);
                            }}
                            onSuccess={clearError}
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
              <DefaultRollout />
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
