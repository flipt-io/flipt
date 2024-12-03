import MoreInfo from '~/components/MoreInfo';
import SegmentForm from '~/components/segments/SegmentForm';
import { PageHeader } from '~/components/ui/page';

export default function NewSegment() {
  return (
    <>
      <PageHeader title="New Segment" />
      <div className="my-6">
        <div className="md:grid md:grid-cols-3 md:gap-6">
          <div className="md:col-span-1">
            <p className="mt-2 text-sm text-gray-500">
              Basic information about the segment.
            </p>
            <MoreInfo
              className="mt-5"
              href="https://www.flipt.io/docs/concepts#segments"
            >
              Learn more about segmentation
            </MoreInfo>
          </div>
          <div className="mt-5 md:col-span-2 md:mt-0">
            <SegmentForm />
          </div>
        </div>
      </div>
    </>
  );
}
