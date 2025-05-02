import { faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { useEffect, useRef, useState } from 'react';

import { ButtonIcon } from '~/components/Button';
import Combobox from '~/components/Combobox';

import { FilterableSegment, ISegment } from '~/types/Segment';

import { truncateKey } from '~/utils/helpers';

type SegmentPickerProps = {
  segments: ISegment[];
  selectedSegments: FilterableSegment[];
  segmentAdd: (segment: FilterableSegment) => void;
  segmentReplace: (index: number, segment: FilterableSegment) => void;
  segmentRemove: (index: number) => void;
};

export default function SegmentsPicker({
  segments,
  selectedSegments: parentSegments,
  segmentAdd,
  segmentReplace,
  segmentRemove
}: SegmentPickerProps) {
  // Track selected segment keys to avoid duplicates
  const segmentsSet = useRef<Set<string>>(
    new Set<string>(
      parentSegments.map((s) => (typeof s === 'string' ? s : s.key))
    )
  );

  const [editing, setEditing] = useState<boolean>(true);

  // Store previous parent segments length to prevent unnecessary updates
  const prevParentSegmentsLength = useRef(parentSegments.length);

  // Update editing state when parent segments change
  useEffect(() => {
    // Set editing to true whenever segments change
    setEditing(true);

    // Update the set of selected segment keys
    segmentsSet.current = new Set<string>(
      parentSegments.map((s) => (typeof s === 'string' ? s : s.key))
    );

    // Update the previous length
    prevParentSegmentsLength.current = parentSegments.length;
  }, [parentSegments]); // Depend on the entire parentSegments array

  const handleSegmentRemove = (index: number) => {
    const filterableSegment = parentSegments[index];
    if (!filterableSegment) return;

    // Remove references to the segment that is being deleted.
    segmentsSet.current.delete(
      typeof filterableSegment === 'string'
        ? filterableSegment
        : filterableSegment.key
    );
    segmentRemove(index);

    if (editing && parentSegments.length === 1) {
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
    const segmentSetCurrent = segmentsSet.current;

    if (index <= parentSegments.length - 1) {
      const previousSegment = selectedSegmentList[index];
      if (previousSegment) {
        const prevKey =
          typeof previousSegment === 'string'
            ? previousSegment
            : previousSegment.key;
        if (segmentSetCurrent.has(prevKey)) {
          segmentSetCurrent.delete(prevKey);
        }
      }

      segmentSetCurrent.add(segment.key);
      segmentReplace(index, segment);
    } else {
      segmentSetCurrent.add(segment.key);
      segmentAdd(segment);
    }
  };

  // Create filtered segment options with proper display values
  const getSegmentOptions = () => {
    return segments
      .filter((s) => !segmentsSet.current.has(s.key))
      .map((s) => ({
        ...s,
        filterValue: truncateKey(s.key),
        displayValue: s.name
      }));
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
              values={getSegmentOptions()}
              selected={selectedSegment}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(index, filterableSegment);
              }}
            />
          </div>
          {editing && parentSegments.length - 1 === index ? (
            <ButtonIcon icon={faPlus} onClick={() => setEditing(false)} />
          ) : (
            <ButtonIcon
              icon={faMinus}
              onClick={() => handleSegmentRemove(index)}
            />
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
              values={getSegmentOptions()}
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
