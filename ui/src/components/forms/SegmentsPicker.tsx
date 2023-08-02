import { MinusIcon } from '@heroicons/react/24/outline';
import { useRef } from 'react';
import Button from '~/components/forms/buttons/Button';
import Combobox from '~/components/forms/Combobox';
import { FilterableSegment, ISegment } from '~/types/Segment';
import { truncateKey } from '~/utils/helpers';

type SegmentPickerProps = {
  segments: ISegment[];
  selectedSegments: FilterableSegment[];
  segmentAdd: (segment: FilterableSegment) => void;
  segmentReplace: (index: number, segment: FilterableSegment) => void;
  segmentRemove: (index: number) => void;
};

export default function SegmentsPicker(props: SegmentPickerProps) {
  const {
    segmentAdd,
    selectedSegments: parentSegments,
    segments,
    segmentRemove,
    segmentReplace
  } = props;

  const segmentsSet = useRef<Set<string>>(new Set<string>());

  const handleSegmentRemove = (index: number) => {
    const filterableSegment = parentSegments[index];

    // Remove references to the segment that is being deleted.
    segmentsSet.current!.delete(filterableSegment.key);
    segmentRemove(index);
  };

  const handleSegmentSelected = (
    index: number,
    segment: FilterableSegment | null
  ) => {
    // Do the replace logic here.
    const selectedSegmentList = [...parentSegments];
    const segmentSetCurrent = segmentsSet.current!;

    if (index <= parentSegments.length - 1) {
      const previousSegment = selectedSegmentList[index];
      if (segmentSetCurrent.has(previousSegment.key)) {
        segmentSetCurrent.delete(previousSegment.key);
      }

      segmentSetCurrent.add(segment?.key!);
      segmentReplace(index, segment!);
    } else {
      segmentSetCurrent.add(segment?.key!);
      segmentAdd(segment!);
    }
  };

  return (
    <div className="space-y-2">
      {parentSegments.map((selectedSegment, index) => (
        <div className="flex space-x-4" key={index}>
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id={`segmentKey-${index}`}
              name={`segmentKey-${index}`}
              placeholder="Select or search for a segment"
              values={segments
                .filter((s) => !segmentsSet.current!.has(s.key))
                .map((s) => {
                  return {
                    ...s,
                    filterValue: truncateKey(s.key),
                    displayValue: s.name
                  };
                })}
              selected={selectedSegment}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(index, filterableSegment);
              }}
            />
          </div>
          <div className="w-1/6">
            <Button
              primary
              type="button"
              onClick={() => handleSegmentRemove(index)}
            >
              <MinusIcon className="text-white h-5" aria-hidden="true" />
            </Button>
          </div>
        </div>
      ))}
      <div className="flex space-x-4">
        <div className="w-5/6">
          <Combobox<FilterableSegment>
            id={`segmentKey-${parentSegments.length}`}
            name={`segmentKey-${parentSegments.length}`}
            placeholder="Select or search for a segment"
            values={segments
              .filter((s) => !segmentsSet.current!.has(s.key))
              .map((s) => {
                return {
                  ...s,
                  filterValue: truncateKey(s.key),
                  displayValue: s.name
                };
              })}
            selected={null}
            setSelected={(filterableSegment) => {
              handleSegmentSelected(parentSegments.length, filterableSegment);
            }}
          />
        </div>
      </div>
    </div>
  );
}
