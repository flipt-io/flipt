import { FolderIcon } from 'lucide-react';
import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';

import { Button, ButtonWithPlus } from '~/components/Button';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import NamespaceForm from '~/components/namespaces/NamespaceForm';
import NamespaceTable from '~/components/namespaces/NamespaceTable';
import DeletePanel from '~/components/panels/DeletePanel';

import { INamespace } from '~/types/Namespace';

import { getRevision } from '~/utils/helpers';

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

  const environment = useSelector(selectCurrentEnvironment);
  const namespaces = useSelector(selectNamespaces);
  const revision = useSelector(getRevision);

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
            deleteNamespace({
              environmentKey: environment.key,
              namespaceKey: deletingNamespace?.key!,
              revision: revision
            }).unwrap()
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
            <Well>
              <FolderIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                No Namespaces Yet
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                Namespaces allow you to group your flags, segments and rules
                under a single name
              </p>
              <Button
                variant="primary"
                onClick={() => {
                  setEditingNamespace(null);
                  setShowNamespaceForm(true);
                }}
              >
                Create Namespace
              </Button>
            </Well>
          )}
        </div>
      </div>
    </>
  );
}
