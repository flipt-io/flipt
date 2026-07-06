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

import { IFlag, flagTypeToLabel } from '~/types/Flag';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { flagsReferencingSegment } from '~/utils/segmentReferences';

function ReferencedFlags({
  flags,
  isLoading,
  namespaceKey,
  segmentKey
}: {
  flags: IFlag[];
  isLoading: boolean;
  namespaceKey: string;
  segmentKey: string;
}) {
  const referencingFlags = useMemo(
    () => flagsReferencingSegment(flags, segmentKey),
    [flags, segmentKey]
  );

  return (
    <section className="mt-8 space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold">Referenced Flags</h2>
        <Badge variant="outlinemuted">
          {isLoading ? 'Loading' : referencingFlags.length}
        </Badge>
      </div>

      <div className="overflow-hidden rounded-lg border dark:border-gray-700">
        {isLoading && (
          <div className="flex items-center gap-2 p-4 text-sm text-muted-foreground">
            <FlagIcon className="h-4 w-4" />
            Loading flags
          </div>
        )}

        {!isLoading && referencingFlags.length === 0 && (
          <div className="flex items-center gap-2 p-4 text-sm text-muted-foreground">
            <FlagIcon className="h-4 w-4" />
            No flags reference this segment
          </div>
        )}

        {!isLoading && referencingFlags.length > 0 && (
          <div className="divide-y dark:divide-gray-700">
            {referencingFlags.map((flag) => (
              <Link
                key={flag.key}
                to={`/namespaces/${namespaceKey}/flags/${flag.key}`}
                className="flex items-center justify-between gap-4 p-4 text-sm transition-colors hover:bg-accent"
              >
                <span className="min-w-0">
                  <span className="block truncate font-medium">
                    {flag.name}
                  </span>
                  <span className="block truncate text-muted-foreground">
                    {flag.key}
                  </span>
                </span>
                <Badge variant="outlinemuted">
                  {flagTypeToLabel(flag.type)}
                </Badge>
              </Link>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

export default function Segment() {
  let { segmentKey } = useParams();

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

  const {
    data: flagsData,
    error: flagsError,
    isLoading: isLoadingFlags
  } = useListFlagsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });

  const [deleteSegment] = useDeleteSegmentMutation();
  const [copySegment] = useCopySegmentMutation();

  useEffect(() => {
    if (isError) {
      setError(error);
    }
  }, [error, isError, setError]);

  useEffect(() => {
    if (flagsError) {
      setError(flagsError);
    }
  }, [flagsError, setError]);

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
        flags={flagsData?.flags || []}
        isLoading={isLoadingFlags}
        namespaceKey={namespace.key}
        segmentKey={segment.key}
      />
    </>
  );
}
