import { PlusIcon } from '@heroicons/react/24/outline';
import { useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { selectReadonly } from '~/app/meta/metaSlice';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import Modal from '~/components/Modal';
import NamespaceForm from '~/components/namespaces/NamespaceForm';
import NamespaceTable from '~/components/namespaces/NamespaceTable';
import DeletePanel from '~/components/panels/DeletePanel';
import Slideover from '~/components/Slideover';
import { INamespace } from '~/types/Namespace';
import {
  useDeleteNamespaceMutation,
  useListNamespacesQuery
} from './namespacesSlice';

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
  const listNamespaces = useListNamespacesQuery();

  const namespaces = useMemo(() => {
    return listNamespaces.data?.namespaces || [];
  }, [listNamespaces]);
  const readOnly = useSelector(selectReadonly);
  const [deleteNamespace] = useDeleteNamespaceMutation();
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
              <span className="text-violet-500 font-medium">
                {deletingNamespace?.key}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Namespace"
          setOpen={setShowDeleteNamespaceModal}
          handleDelete={() =>
            deleteNamespace(deletingNamespace?.key ?? '').unwrap()
          }
        />
      </Modal>

      <div className="my-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h3 className="text-gray-700 text-xl font-semibold">Namespaces</h3>
            <p className="text-gray-500 mt-2 text-sm">
              Namespaces allow you to group your flags, segments and rules under
              a single name
            </p>
          </div>
          <div className="mt-4">
            <Button
              variant="primary"
              disabled={readOnly}
              title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
              onClick={() => {
                setEditingNamespace(null);
                setShowNamespaceForm(true);
              }}
            >
              <PlusIcon
                className="text-white -ml-1.5 mr-1 h-5 w-5"
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
              readOnly={readOnly}
            />
          ) : (
            <EmptyState
              text="New Namespace"
              disabled={readOnly}
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
