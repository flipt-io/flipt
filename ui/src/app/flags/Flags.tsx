import { PlusIcon } from '@heroicons/react/24/outline';
import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import EmptyState from '~/components/EmptyState';
import FlagTable from '~/components/flags/FlagTable';
import Button from '~/components/forms/buttons/Button';
import { useError } from '~/data/hooks/error';
import { useListFlagsQuery } from './flagsApi';

export default function Flags() {
  const namespace = useSelector(selectCurrentNamespace);
  const path = `/namespaces/${namespace.key}/flags`;

  const { data, error } = useListFlagsQuery(namespace.key);
  const flags = data?.flags || [];

  const navigate = useNavigate();
  const { setError, clearError } = useError();

  const readOnly = useSelector(selectReadonly);

  useEffect(() => {
    if (error) {
      setError(error);
      return;
    }
    clearError();
  }, [clearError, error, setError]);

  return (
    <>
      <div className="flex-row justify-between pb-5 sm:flex sm:items-center">
        <div className="flex flex-col">
          <h1 className="text-gray-900 text-2xl font-bold leading-7 sm:truncate sm:text-3xl">
            Flags
          </h1>
          <p className="text-gray-500 mt-2 text-sm">
            Flags represent features that you can easily enable or disable
          </p>
        </div>
        <div className="mt-4">
          <Link to={`${path}/new`}>
            <Button
              variant="primary"
              disabled={readOnly}
              title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
            >
              <PlusIcon
                className="text-white -ml-1.5 mr-1 h-5 w-5"
                aria-hidden="true"
              />
              <span>New Flag</span>
            </Button>
          </Link>
        </div>
      </div>
      <div className="mt-4 flex flex-col">
        {flags && flags.length > 0 ? (
          <FlagTable flags={flags} />
        ) : (
          <EmptyState
            text="Create Flag"
            disabled={readOnly}
            onClick={() => navigate(`${path}/new`)}
          />
        )}
      </div>
    </>
  );
}
