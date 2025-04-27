import { FilesIcon, Trash2Icon } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';
import {
  useCopySegmentMutation,
  useDeleteSegmentMutation,
  useGetSegmentQuery
} from '~/app/segments/segmentsApi';

import Dropdown from '~/components/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import { PageHeader } from '~/components/Page';
import CopyToNamespacePanel from '~/components/panels/CopyToNamespacePanel';
import DeletePanel from '~/components/panels/DeletePanel';
import SegmentForm from '~/components/segments/SegmentForm';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { getRevision } from '~/utils/helpers';

export default function Segment() {
  let { segmentKey } = useParams();

  const [showDeleteSegmentModal, setShowDeleteSegmentModal] =
    useState<boolean>(false);
  const [showCopySegmentModal, setShowCopySegmentModal] =
    useState<boolean>(false);
  const skipRefetch = useRef<boolean>(false);

  const { setError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const environment = useSelector(selectCurrentEnvironment);
  const namespaces = useSelector(selectNamespaces);
  const namespace = useSelector(selectCurrentNamespace);

  const revision = getRevision();

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
      <Modal open={showDeleteSegmentModal} setOpen={setShowDeleteSegmentModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the segment{' '}
              <span className="font-medium text-violet-500">{segment.key}</span>
              ? This action cannot be undone.
            </>
          }
          panelType="Segment"
          setOpen={setShowDeleteSegmentModal}
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
      </Modal>

      {/* segment copy modal */}
      <Modal open={showCopySegmentModal} setOpen={setShowCopySegmentModal}>
        <CopyToNamespacePanel
          panelMessage={
            <>
              Copy the segment{' '}
              <span className="font-medium text-violet-500">{segment.key}</span>{' '}
              to the namespace:
            </>
          }
          panelType="Segment"
          setOpen={setShowCopySegmentModal}
          handleCopy={(namespaceKey: string) =>
            copySegment({
              environmentKey: environment.key,
              from: { namespaceKey: namespace.key, segmentKey: segment.key },
              to: { namespaceKey: namespaceKey, segmentKey: segment.key }
            }).unwrap()
          }
          onSuccess={() => {
            setShowCopySegmentModal(false);
            setSuccess('Successfully copied segment');
          }}
        />
      </Modal>

      {/* segment header / actions */}
      <PageHeader
        title={<div className="flex items-center">{segment.name}</div>}
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
      <div className="mb-8">
        {segment.key && (
          <div className="my-2 inline-flex items-center rounded-md bg-secondary/30 px-3 py-1.5">
            <code className="text-sm font-mono text-muted-foreground">
              {segment.key}
            </code>
          </div>
        )}
      </div>

      {/* Form Section - Full Width */}
      <div className="mt-5">
        <SegmentForm segment={segment} />
      </div>
    </>
  );
}
