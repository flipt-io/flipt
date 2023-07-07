import Combobox, { IFilterable } from '~/components/forms/Combobox';
import { ISegment } from '~/types/Segment';
import { truncateKey } from '~/utils/helpers';

interface SegmentRuleFormInputsFields {
  segmentKey: string;
}

type SegmentRuleFormInputsProps = {
  segments: ISegment[];
  selectedSegment: (ISegment & IFilterable) | null;
  setSelectedSegment: (v: (ISegment & IFilterable) | null) => void;
};

export default function SegmentRuleFormInputs(
  props: SegmentRuleFormInputsProps
) {
  const { segments, selectedSegment, setSelectedSegment } = props;

  return (
    <>
      <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
        <div>
          <label
            htmlFor="segmentKey"
            className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
          >
            Segment
          </label>
        </div>
        <div className="sm:col-span-2">
          <Combobox<ISegment & IFilterable>
            id="segmentKey"
            name="segmentKey"
            placeholder="Select or search for a segment"
            values={segments.map((s) => ({
              ...s,
              filterValue: truncateKey(s.key),
              displayValue: s.name
            }))}
            selected={selectedSegment}
            setSelected={setSelectedSegment}
          />
        </div>
      </div>
    </>
  );
}
