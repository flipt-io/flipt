import { Plus } from 'lucide-react';
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import SegmentTable from '~/components/segments/SegmentTable';
import { Button } from '~/components/ui/button';
import { PageHeader } from '~/components/ui/page';

export default function Segments() {
  const namespace = useSelector(selectCurrentNamespace);

  const path = `/namespaces/${namespace.key}/segments`;

  const navigate = useNavigate();

  const readOnly = useSelector(selectReadonly);

  return (
    <>
      <PageHeader title="Segments">
        <Button onClick={() => navigate(`${path}/new`)} disabled={readOnly}>
          <Plus />
          New Segment
        </Button>
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2 py-2">
        <SegmentTable namespace={namespace} />
      </div>
    </>
  );
}
