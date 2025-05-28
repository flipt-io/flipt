import { MinusIcon, PlusIcon } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

import { IconButton } from '~/components/Button';
import Combobox from '~/components/Combobox';

import { FilterableSegment, ISegment } from '~/types/Segment';

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

  // State to control whether to show the new segment selection field
  const [showNewSegmentField, setShowNewSegmentField] =
    useState<boolean>(false);

  // Store previous parent segments length to prevent unnecessary updates
  const prevParentSegmentsLength = useRef(parentSegments.length);

  // Update segment set when parent segments change
  useEffect(() => {
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
  };

  const handleSegmentSelected = (
    index: number,
    segment: FilterableSegment | null
  ) => {
    if (!segment) {
      return;
    }

    // Get the current set of segment keys to prevent duplicates
    const segmentSetCurrent = segmentsSet.current;

    // Check if this segment is already selected
    if (segmentSetCurrent.has(segment.key)) {
      return;
    }

    if (index < parentSegments.length) {
      // REPLACE: We're updating an existing segment
      const previousSegment = parentSegments[index];
      if (previousSegment) {
        const prevKey =
          typeof previousSegment === 'string'
            ? previousSegment
            : previousSegment.key;

        // Remove the old segment key from the set
        if (segmentSetCurrent.has(prevKey)) {
          segmentSetCurrent.delete(prevKey);
        }
      }

      // Add the new segment key to the set
      segmentSetCurrent.add(segment.key);

      // Update the segment at the specified index
      segmentReplace(index, segment);
    } else {
      // ADD: We're adding a new segment
      // Add the segment key to the set to prevent duplicates
      segmentSetCurrent.add(segment.key);

      // Add the segment to the parent component's state
      segmentAdd(segment);

      // Hide the new segment field after adding
      setShowNewSegmentField(false);
    }
  };

  // Create filtered segment options with proper display values
  const getSegmentOptions = () => {
    return segments
      .filter((s) => !segmentsSet.current.has(s.key))
      .map((s) => ({
        ...s,
        displayValue: s.name
      }));
  };

  if (segments.length == 0) {
    return (
      <div className="space-y-2 text-muted-foreground text-sm mt-1">
        No segments found.
      </div>
    );
  }

  return (
    <div className="space-y-2" data-testid="segments">
      {parentSegments.map((selectedSegment, index) => (
        <div className="flex w-full space-x-1" key={index}>
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id={`segmentKey-${index}`}
              name={`segmentKey-${index}`}
              placeholder="Select or search for a segment"
              values={getSegmentOptions()}
              selected={selectedSegment}
              className="w-full"
              setSelected={(filterableSegment) => {
                handleSegmentSelected(index, filterableSegment);
              }}
            />
          </div>
          <div className="flex space-x-1">
            {parentSegments.length > 1 ? (
              <IconButton
                icon={MinusIcon}
                onClick={() => handleSegmentRemove(index)}
                data-testid={`remove-segment-button-${index}`}
                className="dark:hover:text-violet-300"
              />
            ) : (
              // When there's only one segment and more available to add,
              // show the plus button in place of the minus button
              getSegmentOptions().length > 0 &&
              !showNewSegmentField && (
                <IconButton
                  icon={PlusIcon}
                  onClick={() => setShowNewSegmentField(true)}
                  data-testid={`add-segment-button-${index}`}
                  className="dark:hover:text-violet-300"
                />
              )
            )}
            {/* Only show plus button next to minus when there are multiple segments */}
            {index === parentSegments.length - 1 &&
              parentSegments.length > 1 &&
              !showNewSegmentField &&
              getSegmentOptions().length > 0 && (
                <IconButton
                  icon={PlusIcon}
                  onClick={() => setShowNewSegmentField(true)}
                  data-testid={`add-segment-button-${index}`}
                  className="dark:hover:text-violet-300"
                />
              )}
          </div>
        </div>
      ))}

      {showNewSegmentField && (
        <div className="flex w-full space-x-1">
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id={`segmentKey-${parentSegments.length}`}
              name={`segmentKey-${parentSegments.length}`}
              placeholder="Select or search for a segment"
              values={getSegmentOptions()}
              className="w-full"
              selected={null}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(parentSegments.length, filterableSegment);
              }}
            />
          </div>
          <IconButton
            icon={MinusIcon}
            onClick={() => setShowNewSegmentField(false)}
            data-testid={`remove-segment-button-${parentSegments.length}`}
            className="dark:hover:text-violet-300"
          />
        </div>
      )}

      {parentSegments.length === 0 && getSegmentOptions().length > 0 && (
        <div className="flex w-full space-x-1">
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id="segmentKey-0"
              name="segmentKey-0"
              placeholder="Select or search for a segment"
              values={getSegmentOptions()}
              selected={null}
              setSelected={(filterableSegment) => {
                handleSegmentSelected(0, filterableSegment);
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
