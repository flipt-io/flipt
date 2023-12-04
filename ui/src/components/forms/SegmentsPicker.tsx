import { MinusSmallIcon, PlusSmallIcon } from '@heroicons/react/24/outline';
import { useEffect, useRef, useState } from 'react';
import Combobox from '~/components/forms/Combobox';
import { FilterableSegment, ISegment } from '~/types/Segment';
import { cls, truncateKey } from '~/utils/helpers';

type SegmentPickerProps = {
  readonly?: boolean;
  segments: ISegment[];
  selectedSegments: FilterableSegment[];
  segmentAdd: (segment: FilterableSegment) => void;
  segmentReplace: (index: number, segment: FilterableSegment) => void;
  segmentRemove: (index: number) => void;
};

export default function SegmentsPicker({
  readonly = false,
  segments,
  selectedSegments: parentSegments,
  segmentAdd,
  segmentReplace,
  segmentRemove
}: SegmentPickerProps) {
  const segmentsSet = useRef<Set<string>>(
    new Set<string>(parentSegments.map((s) => s.key))
  );

  const [editing, setEditing] = useState<boolean>(true);

  useEffect(() => {
    setEditing(true);
  }, [parentSegments]);

  const handleSegmentRemove = (index: number) => {
    const filterableSegment = parentSegments[index];

    // Remove references to the segment that is being deleted.
    segmentsSet.current!.delete(filterableSegment.key);
    segmentRemove(index);

    if (editing && parentSegments.length == 1) {
      setEditing(true);
    }
  };

  const handleSegmentSelected = (
    index: number,
    segment: FilterableSegment | null
  ) => {
    if (!segment) {
      return;
    }

    const selectedSegmentList = [...parentSegments];
    const segmentSetCurrent = segmentsSet.current!;

    if (index <= parentSegments.length - 1) {
      const previousSegment = selectedSegmentList[index];
      if (segmentSetCurrent.has(previousSegment.key)) {
        segmentSetCurrent.delete(previousSegment.key);
      }

      segmentSetCurrent.add(segment.key);
      segmentReplace(index, segment);
    } else {
      segmentSetCurrent.add(segment.key);
      segmentAdd(segment);
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
              disabled={readonly}
              selected={selectedSegment}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(index, filterableSegment);
              }}
              inputClassName={
                readonly
                  ? 'cursor-not-allowed bg-gray-100 text-gray-500'
                  : undefined
              }
            />
          </div>
          {editing && parentSegments.length - 1 === index ? (
            <div>
              <button
                type="button"
                className={cls('text-gray-400 mt-2 hover:text-gray-500', {
                  'hover:text-gray-400': readonly
                })}
                onClick={() => setEditing(false)}
                title={readonly ? 'Not allowed in Read-Only mode' : undefined}
                disabled={readonly}
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
                title={readonly ? 'Not allowed in Read-Only mode' : undefined}
                disabled={readonly}
              >
                <MinusSmallIcon className="h-6 w-6" aria-hidden="true" />
              </button>
            </div>
          )}
        </div>
      ))}
      {(!editing || parentSegments.length === 0) && (
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
