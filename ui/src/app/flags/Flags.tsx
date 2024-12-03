import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { Plus } from 'lucide-react';
import { Button } from '~/components/ui/button';
import FlagTable from '~/components/flags/FlagTable';
import { PageHeader } from '~/components/ui/page';

export default function Flags() {
  const namespace = useSelector(selectCurrentNamespace);
  const path = `/namespaces/${namespace.key}/flags`;

  const navigate = useNavigate();

  const readOnly = useSelector(selectReadonly);

  return (
    <>
      <PageHeader title="Flags">
        <Button onClick={() => navigate(`${path}/new`)} disabled={readOnly}>
          <Plus />
          New Flag
        </Button>
      </PageHeader>
      <div className="flex flex-col gap-1 space-y-2 py-2">
        <FlagTable namespace={namespace} />
      </div>
    </>
  );
}
