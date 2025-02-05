import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { ButtonWithPlus } from '~/components/Button';
import SegmentTable from '~/components/segments/SegmentTable';
import { PageHeader } from '~/components/ui/page';

export default function Segments() {
  const namespace = useSelector(selectCurrentNamespace);

  const path = `/namespaces/${namespace.key}/segments`;

  const navigate = useNavigate();

  return (
    <>
      <PageHeader title="Segments">
        <ButtonWithPlus
          variant="primary"
          onClick={() => navigate(`${path}/new`)}
        >
          New Segment
        </ButtonWithPlus>
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2 py-2">
        <SegmentTable namespace={namespace} />
      </div>
    </>
  );
}
