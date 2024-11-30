import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import EmptyState from '~/components/EmptyState';
import { ButtonWithPlus } from '~/components/forms/buttons/Button';
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
          <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl">
            Segments
          </h1>
          <p className="mt-2 text-sm text-gray-500">
            Segments enable request targeting based on defined criteria
          </p>
        </div>
        <div className="mt-4">
          <Link to={`${path}/new`}>
            <ButtonWithPlus
              variant="primary"
              disabled={readOnly}
              title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
            >
              New Segment
            </ButtonWithPlus>
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
