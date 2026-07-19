import { FilesIcon, FlagIcon, Trash2Icon } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate, useParams } from 'react-router';

import {
  selectCurrentEnvironment,
  selectRevision
} from '~/app/environments/environmentsApi';
import { useListFlagsQuery } from '~/app/flags/flagsApi';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';
import {
  useCopySegmentMutation,
  useDeleteSegmentMutation,
  useGetSegmentQuery
} from '~/app/segments/segmentsApi';

import { Badge } from '~/components/Badge';
import Dropdown from '~/components/Dropdown';
import Loading from '~/components/Loading';
import { PageHeader } from '~/components/Page';
import CopyToNamespacePanel from '~/components/panels/CopyToNamespacePanel';
import DeletePanel from '~/components/panels/DeletePanel';
import SegmentForm from '~/components/segments/SegmentForm';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { flagsReferencingSegment } from '~/utils/segmentReferences';

function ReferencedFlags({
  environmentKey,
  namespaceKey,
  segmentKey
}: {
  environmentKey: string;
  namespaceKey: string;
  segmentKey: string;
}) {
  const { data, isLoading, isError, error } = useListFlagsQuery({
    environmentKey: environmentKey,
    namespaceKey: namespaceKey
  });

  const { setError } = useError();
  useEffect(() => {
    if (error) {
      setError(error);
    }
  }, [error, setError]);

  const referencingFlags = useMemo(
    () => flagsReferencingSegment(data?.flags || [], segmentKey),
    [data?.flags, segmentKey]
  );

  return (
    <section className="mt-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="sm:flex-auto">
          <h3 className="font-medium leading-6 text-gray-900 dark:text-gray-100">
            Referenced Flags
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Flags that reference this segment (in rules or rollouts). Changing
            constraints may affect how these flags evaluate.
          </p>
        </div>

        <Badge variant="outlinemuted">
          {isLoading ? 'Loading' : referencingFlags.length}
        </Badge>
      </div>

      <div className="overflow-hidden">
        {isLoading && (
          <div className="text-sm text-muted-foreground">Loading flags</div>
        )}

        {!isLoading && referencingFlags.length === 0 && !isError && (
          <div className="text-sm text-muted-foreground p-3 border rounded-lg">
            No flags reference this segment
          </div>
        )}
        {!isLoading && isError && (
          <div className="text-sm text-muted-foreground p-3 border rounded-lg">
            Unable to load flag references
          </div>
        )}

        {!isLoading && referencingFlags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {referencingFlags.map((flag) => (
              <Badge
                variant="secondary"
                className="font-normal"
                title={flag.name + ' | ' + flag.key}
                key={flag.key}
              >
                <Link
                  to={`/namespaces/${namespaceKey}/flags/${flag.key}`}
                  className="flex items-center justify-between gap-2 text-sm transition-colors hover:bg-accent"
                >
                  <FlagIcon className="h-4 w-4" />{' '}
                  <span className="max-w-24 truncate">{flag.name}</span>
                </Link>
              </Badge>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

export default function Segment() {
  const { segmentKey } = useParams();

  const [showDeleteSegmentModal, setShowDeleteSegmentModal] =
    useState<boolean>(false);
  const [showCopySegmentModal, setShowCopySegmentModal] =
    useState<boolean>(false);
  const skipRefetch = useRef<boolean>(false);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const environment = useSelector(selectCurrentEnvironment);
  const namespaces = useSelector(selectNamespaces);
  const namespace = useSelector(selectCurrentNamespace);
  const revision = useSelector(selectRevision);

  const {
    data: segment,
    error,
    isLoading,
    isError
  } = useGetSegmentQuery(
    {
      environmentKey: environment.key,
      namespaceKey: namespace.key,
      segmentKey: segmentKey || ''
    },
    { skip: skipRefetch.current }
  );

  const [deleteSegment] = useDeleteSegmentMutation();
  const [copySegment] = useCopySegmentMutation();

  useEffect(() => {
    if (isError) {
      setError(error);
    }
  }, [error, isError, setError]);

  if (isLoading || !segment) {
    return <Loading />;
  }

  return (
    <>
      {/* segment delete modal */}
      <DeletePanel
        open={showDeleteSegmentModal}
        setOpen={setShowDeleteSegmentModal}
        panelMessage={
          <>
            Are you sure you want to delete the segment{' '}
            <span className="font-medium text-brand">{segment.key}</span>?
          </>
        }
        panelType="Segment"
        handleDelete={() => {
          skipRefetch.current = true;
          return deleteSegment({
            environmentKey: environment.key,
            namespaceKey: namespace.key,
            segmentKey: segment.key,
            revision
          }).unwrap();
        }}
        onSuccess={() => {
          navigate(`/namespaces/${namespace.key}/segments`, {
            replace: true
          });
          setSuccess('Successfully deleted segment');
        }}
        onError={() => {
          skipRefetch.current = false;
        }}
      />

      {/* segment copy modal */}
      <CopyToNamespacePanel
        open={showCopySegmentModal}
        setOpen={setShowCopySegmentModal}
        panelMessage={
          <>
            Copy the segment{' '}
            <span className="font-medium text-brand">{segment.key}</span> to the
            namespace:
          </>
        }
        panelType="Segment"
        handleCopy={(namespaceKey: string) =>
          copySegment({
            environmentKey: environment.key,
            from: { namespaceKey: namespace.key, segmentKey: segment.key },
            to: { namespaceKey: namespaceKey, segmentKey: segment.key }
          }).unwrap()
        }
        onSuccess={() => {
          clearError();
          setShowCopySegmentModal(false);
          setSuccess('Successfully copied segment');
        }}
      />

      {/* segment header / actions */}
      <PageHeader
        title={
          <div className="flex items-center gap-2">
            {segment.name}
            <Badge variant="outlinemuted" className="hidden sm:block">
              {segment.key}
            </Badge>
          </div>
        }
      >
        <Dropdown
          label="Actions"
          actions={[
            {
              id: 'segment-copy',
              label: 'Copy to Namespace',
              disabled: namespaces.length < 2,
              onClick: () => {
                setShowCopySegmentModal(true);
              },
              icon: FilesIcon
            },
            {
              id: 'segment-delete',
              label: 'Delete',
              onClick: () => setShowDeleteSegmentModal(true),
              icon: Trash2Icon,
              variant: 'destructive'
            }
          ]}
        />
      </PageHeader>

      {/* Key Section */}
      <div className="flex mb-8 gap-3 mt-2">
        <Badge variant="outlinemuted" className="sm:hidden">
          {segment.key}
        </Badge>
      </div>

      {/* Form Section - Full Width */}
      <div className="mt-5">
        <SegmentForm segment={segment} />
      </div>

      <ReferencedFlags
        environmentKey={environment.key}
        namespaceKey={namespace.key}
        segmentKey={segment.key}
      />
    </>
  );
}
