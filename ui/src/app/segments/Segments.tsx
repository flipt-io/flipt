import { PlusIcon } from '@heroicons/react/24/outline';
import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import SegmentTable from '~/components/segments/SegmentTable';
import { useError } from '~/data/hooks/error';

export default function Segments() {
  const namespace = useSelector(selectCurrentNamespace);

  const path = `/namespaces/${namespace.key}/segments`;

  const { data, error } = useListSegmentsQuery(namespace.key);
  const segments = data?.segments || [];
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
            Segments
          </h1>
          <p className="text-gray-500 mt-2 text-sm">
            Segments enable request targeting based on defined criteria
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
              <span>New Segment</span>
            </Button>
          </Link>
        </div>
      </div>
      <div className="mt-4 flex flex-col">
        {segments && segments.length > 0 ? (
          <SegmentTable segments={segments} />
        ) : (
          <EmptyState
            text="Create Segment"
            disabled={readOnly}
            onClick={() => navigate(`${path}/new`)}
          />
        )}
      </div>
    </>
  );
}
