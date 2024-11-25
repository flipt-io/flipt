import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import EmptyState from '~/components/EmptyState';
import FlagTable from '~/components/flags/FlagTable';
import { ButtonWithPlus } from '~/components/forms/buttons/Button';
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
          <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl">
            Flags
          </h1>
          <p className="mt-2 text-sm text-gray-500">
            Flags represent features that you can easily enable or disable
          </p>
        </div>
        <div className="mt-4">
          <Link to={`${path}/new`}>
            <ButtonWithPlus
              variant="primary"
              disabled={readOnly}
              title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
            >
              New Flag
            </ButtonWithPlus>
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
