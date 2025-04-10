import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { ButtonWithPlus } from '~/components/Button';
import { PageHeader } from '~/components/Page';
import FlagTable from '~/components/flags/FlagTable';

export default function Flags() {
  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const path = `/namespaces/${namespace.key}/flags`;

  const navigate = useNavigate();

  return (
    <div className="space-y-6">
      <PageHeader title="Flags">
        <ButtonWithPlus
          variant="primary"
          onClick={() => navigate(`${path}/new`)}
        >
          New Flag
        </ButtonWithPlus>
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2">
        <FlagTable environment={environment} namespace={namespace} />
      </div>
    </div>
  );
}
