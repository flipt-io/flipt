import { Plus } from 'lucide-react';
import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import SegmentTable from '~/components/segments/SegmentTable';
import { Button } from '~/components/ui/button';
import Guide from '~/components/ui/guide';
import { PageHeader } from '~/components/ui/page';
import { useError } from '~/data/hooks/error';

export default function Segments() {
  const namespace = useSelector(selectCurrentNamespace);

  const path = `/namespaces/${namespace.key}/segments`;

  const { data, error } = useListSegmentsQuery(namespace.key);
  const segments = data?.segments || [];
  const navigate = useNavigate();
  const { setError } = useError();

  const readOnly = useSelector(selectReadonly);

  useEffect(() => {
    if (error) {
      setError(error);
      return;
    }
  }, [error, setError]);

  return (
    <>
      <PageHeader title="Segments">
        <Button onClick={() => navigate(`${path}/new`)} disabled={readOnly}>
          <Plus />
          New Segment
        </Button>
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2 py-2">
        {segments && segments.length > 0 ? (
          <SegmentTable segments={segments} />
        ) : (
          <Guide className="mt-6">
            Segments enable request targeting based on defined criteria.
          </Guide>
        )}
      </div>
    </>
  );
}
