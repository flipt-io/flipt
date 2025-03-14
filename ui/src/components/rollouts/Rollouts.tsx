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
import { useContext, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';

import { ButtonWithPlus } from '~/components/Button';
import EmptyState from '~/components/EmptyState';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import DeletePanel from '~/components/panels/DeletePanel';
import EditRolloutForm from '~/components/rollouts/EditRolloutForm';
import Rollout from '~/components/rollouts/Rollout';
import RolloutForm from '~/components/rollouts/RolloutForm';
import SortableRollout from '~/components/rollouts/SortableRollout';

import { IFlag } from '~/types/Flag';
import { IRollout } from '~/types/Rollout';
import { SegmentOperatorType } from '~/types/Segment';

import { useError } from '~/data/hooks/error';

type RolloutsProps = {
  flag: IFlag;
};

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
    return flag.rollouts!.map((rollout) => {
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
  }, [flag.rollouts]);

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
          {rollouts && rollouts.length > 0 && (
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
          )}
        </div>
        <div className="mt-10">
          {rollouts && rollouts.length > 0 ? (
            <div className="flex">
              <div className="dark:pattern-bg-solidwhite pattern-boxes w-full border border-gray-200 p-4 pattern-bg-gray-solid50 pattern-gray-solid100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-gray-solid lg:p-6">
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
              </div>
            </div>
          ) : (
            <EmptyState
              text="New Rollout"
              onClick={() => {
                setEditingRollout(null);
                setShowRolloutForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
