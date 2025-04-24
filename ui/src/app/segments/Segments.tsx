import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { ButtonWithPlus } from '~/components/Button';
import { PageHeader } from '~/components/Page';
import SegmentTable from '~/components/segments/SegmentTable';

import { useListSegmentsQuery } from './segmentsApi';

export default function Segments() {
  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const path = `/namespaces/${namespace.key}/segments`;

  const navigate = useNavigate();

  const { data } = useListSegmentsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });

  const hasSegments = data?.segments && data.segments.length > 0;

  return (
    <div className="space-y-6">
      <PageHeader title="Segments">
        {hasSegments && (
          <ButtonWithPlus
            variant="primary"
            onClick={() => navigate(`${path}/new`)}
          >
            New Segment
          </ButtonWithPlus>
        )}
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2">
        <SegmentTable environment={environment} namespace={namespace} />
      </div>
    </div>
  );
}
