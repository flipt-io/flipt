import { FilesIcon, Trash2Icon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';

import Dropdown from '~/components/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import MoreInfo from '~/components/MoreInfo';
import { PageHeader } from '~/components/Page';
import FlagForm from '~/components/flags/FlagForm';
import CopyToNamespacePanel from '~/components/panels/CopyToNamespacePanel';
import DeletePanel from '~/components/panels/DeletePanel';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { getRevision } from '~/utils/helpers';

import {
  useCopyFlagMutation,
  useDeleteFlagMutation,
  useGetFlagQuery
} from './flagsApi';

export default function Flag() {
  let { flagKey } = useParams();

  const [showDeleteFlagModal, setShowDeleteFlagModal] = useState(false);
  const [showCopyFlagModal, setShowCopyFlagModal] = useState(false);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const environment = useSelector(selectCurrentEnvironment);
  const namespaces = useSelector(selectNamespaces);
  const namespace = useSelector(selectCurrentNamespace);

  const revision = getRevision();

  const {
    data: flag,
    error,
    isLoading,
    isError
  } = useGetFlagQuery({
    environmentKey: environment.name,
    namespaceKey: namespace.key,
    flagKey: flagKey || ''
  });

  const [deleteFlag] = useDeleteFlagMutation();
  const [copyFlag] = useCopyFlagMutation();

  useEffect(() => {
    if (isError) {
      setError(error);
    }
  }, [error, isError, setError]);

  if (isLoading || !flag) {
    return <Loading />;
  }

  return (
    <>
      {/* flag delete modal */}
      <Modal open={showDeleteFlagModal} setOpen={setShowDeleteFlagModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the flag{' '}
              <span className="font-medium text-violet-500">{flag.key}</span>?
              This action cannot be undone.
            </>
          }
          panelType="Flag"
          setOpen={setShowDeleteFlagModal}
          handleDelete={() =>
            deleteFlag({
              environmentKey: environment.name,
              namespaceKey: namespace.key,
              flagKey: flag.key,
              revision
            }).unwrap()
          }
          onSuccess={() => {
            navigate(`/namespaces/${namespace.key}/flags`);
            setSuccess('Successfully deleted flag');
          }}
        />
      </Modal>

      {/* flag copy modal */}
      <Modal open={showCopyFlagModal} setOpen={setShowCopyFlagModal}>
        <CopyToNamespacePanel
          panelMessage={
            <>
              Copy the flag{' '}
              <span className="font-medium text-violet-500">{flag.key}</span> to
              the namespace:
            </>
          }
          panelType="Flag"
          setOpen={setShowCopyFlagModal}
          handleCopy={(namespaceKey: string) =>
            copyFlag({
              environmentKey: environment.name,
              from: { namespaceKey: namespace.key, flagKey: flag.key },
              to: { namespaceKey: namespaceKey, flagKey: flag.key }
            }).unwrap()
          }
          onSuccess={() => {
            clearError();
            setShowCopyFlagModal(false);
            setSuccess('Successfully copied flag');
          }}
        />
      </Modal>

      {/* flag header / actions */}
      <PageHeader title={flag.name}>
        <Dropdown
          label="Actions"
          actions={[
            {
              id: 'flag-copy',
              label: 'Copy to Namespace',
              disabled: namespaces.length < 2,
              onClick: () => {
                setShowCopyFlagModal(true);
              },
              icon: FilesIcon
            },
            {
              id: 'flag-delete',
              label: 'Delete',
              onClick: () => setShowDeleteFlagModal(true),
              icon: Trash2Icon,
              variant: 'destructive'
            }
          ]}
        />
      </PageHeader>

      {/* Info Section */}
      <div className="mb-8 space-y-4">
        <MoreInfo href="https://www.flipt.io/docs/concepts#flags">
          Learn more about flags
        </MoreInfo>
      </div>

      {/* Form Section - Full Width */}
      <div className="mt-5">
        <FlagForm flag={flag} />
      </div>
    </>
  );
}
