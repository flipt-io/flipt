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
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import DeletePanel from '~/components/DeletePanel';
import Dropdown from '~/components/forms/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import TabBar from '~/components/TabBar';
import { copyFlag, deleteFlag, getFlag } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useTimezone } from '~/data/hooks/timezone';
import { IFlag } from '~/types/Flag';

export default function Flag() {
  let { flagKey } = useParams();
  const { inTimezone } = useTimezone();

  const [flag, setFlag] = useState<IFlag | null>(null);
  const [flagVersion, setFlagVersion] = useState(0);

  const { setError, clearError } = useError();
  const navigate = useNavigate();

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const [showDeleteFlagModal, setShowDeleteFlagModal] =
    useState<boolean>(false);

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
        clearError();
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
                disabled: readOnly,
                onClick: () =>
                  copyFlag(
                    { namespaceKey: namespace.key, key: flag.key },
                    { namespaceKey: 'adsfdsafdsaf' }
                  ),
                icon: DocumentDuplicateIcon
              },
              {
                id: 'delete',
                label: 'Delete',
                disabled: readOnly,
                onClick: () => setShowDeleteFlagModal(true),
                icon: TrashIcon,
                activeClassName: 'text-red-700',
                inActiveClassName: 'text-red-600'
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
