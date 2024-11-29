import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useError } from '~/data/hooks/error';
import { useListFlagsQuery } from './flagsApi';
import { Plus } from 'lucide-react';
import { Button } from '~/components/ui/button';
import FlagTable from '~/components/flags/FlagTable';
import Guide from '~/components/ui/guide';
import { PageHeader } from '~/components/ui/page';

export default function Flags() {
  const namespace = useSelector(selectCurrentNamespace);
  const path = `/namespaces/${namespace.key}/flags`;

  const { data, error } = useListFlagsQuery(namespace.key);
  const flags = data?.flags || [];

  const navigate = useNavigate();
  const { setError } = useError();

  const readOnly = useSelector(selectReadonly);

  useEffect(() => {
    if (error) {
      setError(error);
    }
  }, [error, setError]);

  return (
    <>
      <PageHeader title="Flags">
        <Button onClick={() => navigate(`${path}/new`)} disabled={readOnly}>
          <Plus />
          New Flag
        </Button>
      </PageHeader>
      <div className="flex flex-col gap-1 py-2">
        {flags && flags.length > 0 ? (
          <FlagTable flags={flags} />
        ) : (
          <Guide className="mt-6">
            Flags enable you to control and roll out new functionality
            dynamically. Create a new flag to get started.
          </Guide>
        )}
      </div>
    </>
  );
}
