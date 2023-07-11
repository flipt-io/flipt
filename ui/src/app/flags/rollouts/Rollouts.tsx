import { InformationCircleIcon, PlusIcon } from '@heroicons/react/24/outline';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import RolloutForm from '~/components/rollouts/RolloutForm';
import SortableRollout from '~/components/rollouts/SortableRollout';
import Slideover from '~/components/Slideover';
import { deleteRollout, listRollouts, listSegments } from '~/data/api';
import { IFlag } from '~/types/Flag';
import { IRollout, IRolloutList } from '~/types/Rollout';
import { ISegment, ISegmentList } from '~/types/Segment';

type RolloutsProps = {
  flag: IFlag;
};

export default function Rollouts(props: RolloutsProps) {
  const { flag } = props;

  const [segments, setSegments] = useState<ISegment[]>([]);
  const [rollouts, setRollouts] = useState<IRollout[]>([]);

  const [rolloutsVersion, setRolloutsVersion] = useState(0);
  const [showRolloutForm, setShowRolloutForm] = useState<boolean>(false);

  const [showDeleteRolloutModal, setShowDeleteRolloutModal] =
    useState<boolean>(false);
  const [deletingRollout, setDeletingRollout] = useState<IRollout | null>(null);

  const rolloutFormRef = useRef(null);

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const loadData = useCallback(async () => {
    // TODO: move to redux
    const segmentList = (await listSegments(namespace.key)) as ISegmentList;
    const { segments } = segmentList;
    setSegments(segments);

    const rolloutList = (await listRollouts(
      namespace.key,
      flag.key
    )) as IRolloutList;

    setRollouts(rolloutList.rules);
  }, [namespace.key, flag.key]);

  const incrementRolloutsVersion = () => {
    setRolloutsVersion(rolloutsVersion + 1);
  };

  useEffect(() => {
    loadData();
  }, [loadData, rolloutsVersion]);

  return (
    <>
      {/* rollout edit form */}
      <Slideover
        open={showRolloutForm}
        setOpen={setShowRolloutForm}
        ref={rolloutFormRef}
      >
        <RolloutForm
          flagKey={flag.key}
          segments={segments}
          // rollout={editingRollout || undefined}
          rank={(rollouts?.length || 0) + 1}
          setOpen={setShowRolloutForm}
          onSuccess={() => {
            setShowRolloutForm(false);
            incrementRolloutsVersion();
          }}
        />
      </Slideover>

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
          handleDelete={
            () =>
              deleteRollout(namespace.key, flag.key, deletingRollout?.id ?? '') // TODO: Determine impact of blank ID param
          }
          onSuccess={incrementRolloutsVersion}
        />
      </Modal>

      {/* rollouts */}
      <div className="mt-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-lg font-medium leading-6 text-gray-900">
              Rollouts
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Return boolean values based on rules you define
            </p>
          </div>
          {rollouts && rollouts.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                primary
                type="button"
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
                onClick={() => {
                  // setEditingRollout(null);
                  setShowRolloutForm(true);
                }}
              >
                <PlusIcon
                  className="-ml-1.5 mr-1 h-5 w-5 text-white"
                  aria-hidden="true"
                />
                <span>New Rollout</span>
              </Button>
            </div>
          )}
        </div>
        <div className="mt-10">
          {rollouts && rollouts.length > 0 ? (
            <div className="flex lg:space-x-5">
              <div className="hidden w-1/4 flex-col space-y-7 pr-3 lg:flex">
                <p className="text-sm text-gray-500">
                  Rules are evaluated in order from{' '}
                  <span className="font-semibold">top to bottom</span>. The
                  first rule that matches will be applied.
                </p>
                <p className="text-sm text-gray-500">
                  <InformationCircleIcon className="mr-1 inline-block h-4 w-4 text-violet-300" />
                  You can re-arrange rules by{' '}
                  <span className="font-semibold">dragging and dropping</span>{' '}
                  them into place.
                </p>
              </div>
              <div
                className="pattern-boxes w-full border p-4 pattern-bg-gray-50 pattern-gray-100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-black dark:pattern-gray-900  
  lg:w-3/4 lg:p-6"
              >
                {/* <DndContext
                  sensors={sensors}
                  collisionDetection={closestCenter}
                  onDragStart={onDragStart}
                  onDragEnd={onDragEnd}
                >
                  <SortableContext
                    items={rules.map((rule) => rule.id)}
                    strategy={verticalListSortingStrategy}
                  > */}
                <ul role="list" className="flex-col space-y-6 p-2 md:flex">
                  {rollouts &&
                    rollouts.length > 0 &&
                    rollouts.map((rollout) => (
                      <SortableRollout
                        key={rollout.id}
                        flagKey={flag.key}
                        rollout={rollout}
                        segments={segments}
                        onQuickEditSuccess={incrementRolloutsVersion}
                        onEdit={() => {
                          // setEditingRollout(rollout);
                          //setShowEditRolloutForm(true);
                        }}
                        onDelete={() => {
                          setDeletingRollout(rollout);
                          setShowDeleteRolloutModal(true);
                        }}
                        readOnly={readOnly}
                      />
                    ))}
                </ul>
                {/* </SortableContext>
                  <DragOverlay>
                    {activeRule ? (
                      <NewRule
                        namespace={namespace}
                        totalRules={rules.length}
                        rule={activeRule}
                      />
                    ) : null}
                  </DragOverlay>
                </DndContext> */}
              </div>
            </div>
          ) : (
            <EmptyState
              text="New Rollout"
              disabled={readOnly}
              onClick={() => {
                //setEditingRollout(null);
                setShowRolloutForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
