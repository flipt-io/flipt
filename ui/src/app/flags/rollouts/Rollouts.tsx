import { PlusIcon } from '@heroicons/react/24/outline';
import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import DeletePanel from '~/components/DeletePanel';
import EmptyState from '~/components/EmptyState';
import RolloutForm from '~/components/flags/rollouts/RolloutForm';
import Button from '~/components/forms/Button';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import { deleteRollout } from '~/data/api';
import { IFlag } from '~/types/Flag';
import { IRollout } from '~/types/Rollout';
import { selectCurrentNamespace } from '../../namespaces/namespacesSlice';

type RolloutsProps = {
  flag: IFlag;
  onFlagChange: () => void;
};

export default function Rollouts(props: RolloutsProps) {
  const { flag, onFlagChange } = props;

  const [showRolloutForm, setShowRolloutForm] = useState<boolean>(false);
  const [editingRollout, setEditingRollout] = useState<IRollout | null>(null);
  const [showDeleteRolloutModal, setShowDeleteRolloutModal] =
    useState<boolean>(false);
  const [deletingRollout, setDeletingRollout] = useState<IRollout | null>(null);

  const rolloutFormRef = useRef(null);

  const namespace = useSelector(selectCurrentNamespace);

  return (
    <>
      {/* rollout edit form */}
      <Slideover
        open={showRolloutForm}
        setOpen={setShowRolloutForm}
        ref={rolloutFormRef}
      >
        <RolloutForm
          ref={rolloutFormRef}
          flagKey={flag.key}
          rollout={editingRollout || undefined}
          setOpen={setShowRolloutForm}
          onSuccess={() => {
            setShowRolloutForm(false);
            onFlagChange();
          }}
        />
      </Slideover>

      {/* rollout delete modal */}
      <Modal open={showDeleteRolloutModal} setOpen={setShowDeleteRolloutModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the rollout at position{' '}
              <span className="font-medium text-violet-500">
                {deletingRollout?.rank}
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
          onSuccess={() => {
            onFlagChange();
          }}
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
              Return different values based on rules you define
            </p>
          </div>
          {flag.rollouts && flag.rollouts.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                primary
                type="button"
                onClick={() => {
                  setEditingRollout(null);
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

        <div className="my-10">
          {flag.rollouts && flag.rollouts.length > 0 ? (
            <table className="min-w-full divide-y divide-gray-300">
              <thead>
                <tr>
                  <th
                    scope="col"
                    className="pb-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6"
                  >
                    Key
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 sm:table-cell"
                  >
                    Name
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
                {flag.rollouts.map((rollout) => (
                  <tr key={rollout.key}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm text-gray-600 sm:pl-6">
                      {rollout.key}
                    </td>
                    <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 sm:table-cell">
                      {rollout.name}
                    </td>
                    <td className="hidden truncate whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                      {rollout.description}
                    </td>
                    <td className="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <a
                        href="#"
                        className="pr-2 text-violet-600 hover:text-violet-900"
                        onClick={(e) => {
                          e.preventDefault();
                          setEditingRollout(rollout);
                          setShowRolloutForm(true);
                        }}
                      >
                        Edit
                        <span className="sr-only">,{rollout.key}</span>
                      </a>
                      |
                      <a
                        href="#"
                        className="pl-2 text-violet-600 hover:text-violet-900"
                        onClick={(e) => {
                          e.preventDefault();
                          setDeletingRollout(rollout);
                          setShowDeleteRolloutModal(true);
                        }}
                      >
                        Delete
                        <span className="sr-only">,{rollout.key}</span>
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
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
