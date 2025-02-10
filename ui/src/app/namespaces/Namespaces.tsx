import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { ButtonWithPlus } from '~/components/Button';
import EmptyState from '~/components/EmptyState';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import NamespaceForm from '~/components/namespaces/NamespaceForm';
import NamespaceTable from '~/components/namespaces/NamespaceTable';
import DeletePanel from '~/components/panels/DeletePanel';

import { INamespace } from '~/types/Namespace';

import { selectNamespaces } from './namespacesApi';
import { useDeleteNamespaceMutation } from './namespacesApi';

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

  const namespaces = useSelector(selectNamespaces);
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
          namespace={editingNamespace || null}
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
          handleDelete={() =>
            deleteNamespace(deletingNamespace?.key ?? '').unwrap()
          }
        />
      </Modal>

      <div className="my-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h3 className="text-xl font-semibold text-gray-700">Namespaces</h3>
            <p className="mt-2 text-sm text-gray-500">
              Namespaces allow you to group your flags, segments and rules under
              a single name
            </p>
          </div>
          <div className="mt-4">
            <ButtonWithPlus
              variant="primary"
              onClick={() => {
                setEditingNamespace(null);
                setShowNamespaceForm(true);
              }}
            >
              New Namespace
            </ButtonWithPlus>
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
