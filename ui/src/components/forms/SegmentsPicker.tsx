import { MinusSmallIcon, PlusSmallIcon } from '@heroicons/react/24/outline';
import { useRef, useState } from 'react';
import Combobox from '~/components/forms/Combobox';
import { FilterableSegment, ISegment } from '~/types/Segment';
import { truncateKey } from '~/utils/helpers';

type SegmentPickerProps = {
  editMode?: boolean;
  segments: ISegment[];
  selectedSegments: FilterableSegment[];
  segmentAdd: (segment: FilterableSegment) => void;
  segmentReplace: (index: number, segment: FilterableSegment) => void;
  segmentRemove: (index: number) => void;
};

export default function SegmentsPicker({
  editMode = false,
  segments,
  selectedSegments: parentSegments,
  segmentAdd,
  segmentReplace,
  segmentRemove
}: SegmentPickerProps) {
  const segmentsSet = useRef<Set<string>>(
    new Set<string>(parentSegments.map((s) => s.key))
  );

  const handleSegmentRemove = (index: number) => {
    const filterableSegment = parentSegments[index];

    // Remove references to the segment that is being deleted.
    segmentsSet.current!.delete(filterableSegment.key);
    segmentRemove(index);

    if (editMode && parentSegments.length == 1) {
      setEditing(true);
    }
  };

  const [editing, setEditing] = useState<boolean>(editMode);

  const handleSegmentSelected = (
    index: number,
    segment: FilterableSegment | null
  ) => {
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
        <div className="flex w-full space-x-1" key={index}>
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id={`segmentKey-${index}`}
              name={`segmentKey-${index}`}
              placeholder="Select or search for a segment"
              values={segments
                .filter((s) => !segmentsSet.current!.has(s.key))
                .map((s) => ({
                  ...s,
                  filterValue: truncateKey(s.key),
                  displayValue: s.name
                }))}
              selected={selectedSegment}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(index, filterableSegment);
              }}
            />
          </div>
          {editing ? (
            <div>
              <button
                type="button"
                className="text-gray-400 mt-2 hover:text-gray-500"
                onClick={() => setEditing(false)}
              >
                <PlusSmallIcon className="h-6 w-6" aria-hidden="true" />
              </button>
            </div>
          ) : (
            <div>
              <button
                type="button"
                className="text-gray-400 mt-2 hover:text-gray-500"
                onClick={() => handleSegmentRemove(index)}
              >
                <MinusSmallIcon className="h-6 w-6" aria-hidden="true" />
              </button>
            </div>
          )}
        </div>
      ))}
      {!editing && (
        <div className="w-full">
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id={`segmentKey-${parentSegments.length}`}
              name={`segmentKey-${parentSegments.length}`}
              placeholder="Select or search for a segment"
              values={segments
                .filter((s) => !segmentsSet.current!.has(s.key))
                .map((s) => ({
                  ...s,
                  filterValue: truncateKey(s.key),
                  displayValue: s.name
                }))}
              selected={null}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(parentSegments.length, filterableSegment);
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
