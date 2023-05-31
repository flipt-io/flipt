import { PlusIcon } from '@heroicons/react/24/outline';
import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import DeletePanel from '~/components/DeletePanel';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/Button';
import Modal from '~/components/Modal';
import NamespaceForm from '~/components/namespaces/NamespaceForm';
import NamespaceTable from '~/components/namespaces/NamespaceTable';
import Slideover from '~/components/Slideover';
import { useAppDispatch } from '~/data/hooks/store';
import { INamespace } from '~/types/Namespace';
import { deleteNamespaceAsync, selectNamespaces } from './namespacesSlice';

export default function Namespaces() {
  const [showNamespaceForm, setShowNamespaceForm] = useState<boolean>(false);

  const [editingNamespace, setEditingNamespace] = useState<INamespace | null>(
    null
  );

  const [showDeleteNamespaceModal, setShowDeleteNamespaceModal] =
    useState<boolean>(false);
  const [deletingNamespace, setDeletingNamespace] = useState<INamespace | null>(
    null
  );

  const dispatch = useAppDispatch();

  const namespaces = useSelector(selectNamespaces);
  const namespaceFormRef = useRef(null);

  return (
    <>
      {/* namespace edit form */}
      <Slideover
        open={showNamespaceForm}
        setOpen={setShowNamespaceForm}
        ref={namespaceFormRef}
      >
        <NamespaceForm
          ref={namespaceFormRef}
          namespace={editingNamespace || undefined}
          setOpen={setShowNamespaceForm}
          onSuccess={() => {
            setShowNamespaceForm(false);
            setEditingNamespace(null);
          }}
        />
      </Slideover>

      {/* namespace delete modal */}
      <Modal
        open={showDeleteNamespaceModal}
        setOpen={setShowDeleteNamespaceModal}
      >
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the namespace{' '}
              <span className="font-medium text-violet-500">
                {deletingNamespace?.key}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Namespace"
          setOpen={setShowDeleteNamespaceModal}
          handleDelete={
            () => dispatch(deleteNamespaceAsync(deletingNamespace?.key ?? '')) // TODO: Determine impact of blank ID param
          }
        />
      </Modal>

      <div className="my-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-xl font-semibold text-gray-700">Namespaces</h1>
            <p className="mt-2 text-sm text-gray-500">
              Namespaces allow you to group your flags, segments and rules under
              a single name
            </p>
          </div>
          <div className="mt-4">
            <Button
              primary
              onClick={() => {
                setEditingNamespace(null);
                setShowNamespaceForm(true);
              }}
            >
              <PlusIcon
                className="-ml-1.5 mr-1 h-5 w-5 text-white"
                aria-hidden="true"
              />
              <span>New Namespace</span>
            </Button>
          </div>
        </div>

        <div className="mt-8 flex flex-col">
          {namespaces && namespaces.length > 0 ? (
            <NamespaceTable
              namespaces={namespaces}
              setEditingNamespace={setEditingNamespace}
              setShowEditNamespaceModal={setShowNamespaceForm}
              setDeletingNamespace={setDeletingNamespace}
              setShowDeleteNamespaceModal={setShowDeleteNamespaceModal}
            />
          ) : (
            <EmptyState
              text="New Namespace"
              onClick={() => {
                setEditingNamespace(null);
                setShowNamespaceForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
