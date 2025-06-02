import {
  FilesIcon,
  LineChartIcon,
  SquareTerminalIcon,
  ToggleLeftIcon,
  Trash2Icon,
  VariableIcon
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import { useParams } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';

import { Badge } from '~/components/Badge';
import Dropdown from '~/components/Dropdown';
import Loading from '~/components/Loading';
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
      <DeletePanel
        open={showDeleteFlagModal}
        panelMessage={
          <>
            Are you sure you want to delete the flag{' '}
            <span className="font-medium text-brand">{flag.key}</span>?
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

      {/* flag copy modal */}
      <CopyToNamespacePanel
        open={showCopyFlagModal}
        panelMessage={
          <>
            Copy the flag{' '}
            <span className="font-medium text-brand">{flag.key}</span> to the
            namespace:
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

      {/* flag header / actions */}
      <PageHeader
        title={
          <div className="flex items-center gap-2">
            {flag.name}
            <Badge variant="outlinemuted" className="hidden sm:block">
              {flag.key}
            </Badge>
            {info.analytics?.enabled && (
              <button
                className=" p-1 rounded hover:bg-accent hidden sm:block"
                title="View Analytics"
                onClick={() =>
                  navigate(`/namespaces/${namespace.key}/analytics/${flag.key}`)
                }
              >
                <LineChartIcon className="h-5 w-5 text-muted-foreground" />
              </button>
            )}
            <button
              className="p-1 rounded hover:bg-accent hidden sm:block"
              title="View in Playground"
              onClick={() =>
                navigate(`/namespaces/${namespace.key}/playground/${flag.key}`)
              }
            >
              <SquareTerminalIcon className="h-5 w-5 text-muted-foreground" />
            </button>
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
                        `/namespaces/${namespace.key}/analytics/${flag.key}`
                      ),
                    icon: LineChartIcon
                  }
                ]
              : []),
            {
              id: 'flag-playground',
              label: 'View in Playground',
              onClick: () =>
                navigate(`/namespaces/${namespace.key}/playground/${flag.key}`),
              icon: SquareTerminalIcon
            },
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
      <div className="flex mb-8 gap-3 mt-2">
        <Badge variant="outlinemuted" className="sm:hidden">
          {flag.key}
        </Badge>
        <div className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium bg-secondary/50 text-muted-foreground">
          {flag.type === FlagType.BOOLEAN ? (
            <ToggleLeftIcon className="h-3.5 w-3.5" />
          ) : (
            <VariableIcon className="h-3.5 w-3.5" />
          )}
          {flagTypeToLabel(flag.type)}
        </div>
      </div>

      {/* Form Section - Full Width */}
      <div className="mt-5">
        <FlagForm flag={flag} />
      </div>
    </>
  );
}
