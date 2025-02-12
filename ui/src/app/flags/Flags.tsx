import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { ButtonWithPlus } from '~/components/Button';
import FlagTable from '~/components/flags/FlagTable';
import { PageHeader } from '~/components/Page';

export default function Flags() {
  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const path = `/namespaces/${namespace.key}/flags`;

  const navigate = useNavigate();

  return (
    <>
      <PageHeader title="Flags">
        <ButtonWithPlus
          variant="primary"
          onClick={() => navigate(`${path}/new`)}
        >
          New Flag
        </ButtonWithPlus>
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2 py-2">
        <FlagTable environment={environment} namespace={namespace} />
      </div>
    </>
  );
}
