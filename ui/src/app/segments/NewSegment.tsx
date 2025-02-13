import MoreInfo from '~/components/MoreInfo';
import { PageHeader } from '~/components/Page';
import SegmentForm from '~/components/segments/SegmentForm';

export default function NewSegment() {
  return (
    <>
      <PageHeader title="New Segment" />
      <div className="mb-8 space-y-4">
        <MoreInfo href="https://www.flipt.io/docs/concepts#segments">
          Learn more about segments
        </MoreInfo>
      </div>

      <div className="mb-8">
        <SegmentForm />
      </div>
    </>
  );
}
