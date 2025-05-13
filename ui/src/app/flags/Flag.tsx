import {
  FilesIcon,
  ToggleLeftIcon,
  Trash2Icon,
  VariableIcon,
  LineChartIcon
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import { useParams } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';
import { selectInfo } from '~/app/meta/metaSlice';

import Dropdown from '~/components/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import { PageHeader } from '~/components/Page';
import FlagForm from '~/components/flags/FlagForm';
import CopyToNamespacePanel from '~/components/panels/CopyToNamespacePanel';
import DeletePanel from '~/components/panels/DeletePanel';

import { FlagType, flagTypeToLabel } from '~/types/Flag';

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
  const skipRefetch = useRef<boolean>(false);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const environment = useSelector(selectCurrentEnvironment);
  const namespaces = useSelector(selectNamespaces);
  const namespace = useSelector(selectCurrentNamespace);

  const revision = getRevision();

  const info = useSelector(selectInfo);

  const {
    data: flag,
    error,
    isLoading,
    isError
  } = useGetFlagQuery(
    {
      environmentKey: environment.key,
      namespaceKey: namespace.key,
      flagKey: flagKey || ''
    },
    { skip: skipRefetch.current }
  );

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
              <span className="font-medium text-violet-500 dark:text-violet-400">
                {flag.key}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Flag"
          setOpen={setShowDeleteFlagModal}
          handleDelete={() => {
            skipRefetch.current = true;
            return deleteFlag({
              environmentKey: environment.key,
              namespaceKey: namespace.key,
              flagKey: flag.key,
              revision
            }).unwrap();
          }}
          onSuccess={() => {
            navigate(`/namespaces/${namespace.key}/flags`);
            setSuccess('Successfully deleted flag');
          }}
          onError={() => {
            skipRefetch.current = false;
          }}
        />
      </Modal>

      {/* flag copy modal */}
      <Modal open={showCopyFlagModal} setOpen={setShowCopyFlagModal}>
        <CopyToNamespacePanel
          panelMessage={
            <>
              Copy the flag{' '}
              <span className="font-medium text-violet-500 dark:text-violet-400">
                {flag.key}
              </span>{' '}
              to the namespace:
            </>
          }
          panelType="Flag"
          setOpen={setShowCopyFlagModal}
          handleCopy={(namespaceKey: string) =>
            copyFlag({
              environmentKey: environment.key,
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
      <PageHeader
        title={
          <div className="flex items-center">
            {flag.name}
            <div className="ml-4 inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium bg-secondary/50 text-secondary-foreground">
              {flag.type === FlagType.BOOLEAN ? (
                <ToggleLeftIcon className="h-3.5 w-3.5" />
              ) : (
                <VariableIcon className="h-3.5 w-3.5" />
              )}
              {flagTypeToLabel(flag.type)}
            </div>
            {info.analytics?.enabled && (
              <button
                className="ml-2 p-1 rounded hover:bg-accent"
                title="View Analytics"
                onClick={() => navigate(`/namespaces/${namespace.key}/analytics?flag=${flag.key}`)}
              >
                <LineChartIcon className="h-5 w-5 text-muted-foreground" />
              </button>
            )}
          </div>
        }
      >
        <Dropdown
          label="Actions"
          actions={[
            ...(info.analytics?.enabled
              ? [
                  {
                    id: 'flag-analytics',
                    label: 'View Analytics',
                    onClick: () =>
                      navigate(
                        `/namespaces/${namespace.key}/analytics?flag=${flag.key}`
                      ),
                    icon: LineChartIcon
                  }
                ]
              : []),
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
      <div className="mb-8">
        {flag.key && (
          <div className="my-2 inline-flex items-center rounded-md bg-secondary/30 px-3 py-1.5">
            <code className="text-sm font-mono text-muted-foreground">
              {flag.key}
            </code>
          </div>
        )}
      </div>

      {/* Form Section - Full Width */}
      <div className="mt-5">
        <FlagForm flag={flag} />
      </div>
    </>
  );
}
