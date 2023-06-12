import {
  CalendarIcon,
  DocumentDuplicateIcon,
  TrashIcon
} from '@heroicons/react/24/outline';
import { formatDistanceToNowStrict, parseISO } from 'date-fns';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Outlet, useNavigate, useParams } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import Dropdown from '~/components/forms/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import CopyToNamespacePanel from '~/components/panels/CopyToNamespacePanel';
import DeletePanel from '~/components/panels/DeletePanel';
import TabBar from '~/components/TabBar';
import { copyFlag, deleteFlag, getFlag } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { useTimezone } from '~/data/hooks/timezone';
import { IFlag } from '~/types/Flag';

export default function Flag() {
  let { flagKey } = useParams();
  const { inTimezone } = useTimezone();

  const [flag, setFlag] = useState<IFlag | null>(null);
  const [flagVersion, setFlagVersion] = useState(0);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const namespaces = useSelector(selectNamespaces);
  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const [showDeleteFlagModal, setShowDeleteFlagModal] = useState(false);
  const [showCopyFlagModal, setShowCopyFlagModal] = useState(false);

  const incrementFlagVersion = () => {
    setFlagVersion(flagVersion + 1);
  };

  const tabs = [
    {
      name: 'Details',
      to: ''
    },
    {
      name: 'Evaluation',
      to: 'evaluation'
    }
  ];

  useEffect(() => {
    if (!flagKey) return;

    getFlag(namespace.key, flagKey)
      .then((flag: IFlag) => {
        setFlag(flag);
      })
      .catch((err) => {
        setError(err);
      });
  }, [flagVersion, flagKey, namespace.key, clearError, setError]);

  if (!flag) return <Loading />;

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
          handleDelete={() => deleteFlag(namespace.key, flag.key)}
          onSuccess={() => {
            navigate(`/namespaces/${namespace.key}/flags`);
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
            copyFlag(
              { namespaceKey: namespace.key, key: flag.key },
              { namespaceKey: namespaceKey, key: flag.key }
            )
          }
          onSuccess={() => {
            clearError();
            setShowCopyFlagModal(false);
            setSuccess('Successfully copied flag');
          }}
        />
      </Modal>

      {/* flag header / actions */}
      <div className="flex items-center justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
            {flag.name}
          </h2>
          <div className="mt-1 flex flex-col sm:mt-0 sm:flex-row sm:flex-wrap sm:space-x-6">
            <div
              title={inTimezone(flag.createdAt)}
              className="mt-2 flex items-center text-sm text-gray-500"
            >
              <CalendarIcon
                className="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400"
                aria-hidden="true"
              />
              Created{' '}
              {formatDistanceToNowStrict(parseISO(flag.createdAt), {
                addSuffix: true
              })}
            </div>
          </div>
        </div>
        <div className="flex flex-none">
          <Dropdown
            label="Actions"
            actions={[
              {
                id: 'copy',
                label: 'Copy to Namespace',
                disabled: readOnly || namespaces.length < 2,
                onClick: () => {
                  setShowCopyFlagModal(true);
                },
                icon: DocumentDuplicateIcon
              },
              {
                id: 'delete',
                label: 'Delete',
                disabled: readOnly,
                onClick: () => setShowDeleteFlagModal(true),
                icon: TrashIcon,
                activeClassName: readOnly ? 'text-red-500' : 'text-red-700',
                inActiveClassName: readOnly ? 'text-red-400' : 'text-red-600'
              }
            ]}
          />
        </div>
      </div>
      <TabBar tabs={tabs} />
      <Outlet context={{ flag, onFlagChange: incrementFlagVersion }} />
    </>
  );
}
