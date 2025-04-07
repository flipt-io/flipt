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
import { useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import {
  useDeleteRolloutMutation,
  useListRolloutsQuery,
  useOrderRolloutsMutation
} from '~/app/flags/rolloutsApi';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import { ButtonWithPlus } from '~/components/Button';
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
import { IRollout } from '~/types/Rollout';
import { SegmentOperatorType } from '~/types/Segment';
import EmptyState from '~/components/EmptyState';

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
              <span className="font-medium text-violet-500">
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
            <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
              <ButtonWithPlus
                variant="primary"
                type="button"
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
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
              <div className="dark:pattern-bg-solidwhite pattern-boxes pattern-bg-gray-solid50 pattern-gray-solid100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-gray-solid w-full border border-gray-200 p-4 lg:p-6">
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
              </div>
            </div>
          ) : (
            <EmptyState
              text="New Rollout"
              disabled={readOnly}
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
