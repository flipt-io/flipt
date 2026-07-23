import { MinusIcon, PlusIcon } from 'lucide-react';
import { useMemo, useState } from 'react';

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
  // State to control whether to show the new segment selection field
  const [showNewSegmentField, setShowNewSegmentField] =
    useState<boolean>(false);

  const segmentOptions = useMemo(() => {
    const selectedKeys = new Set<string>(
      parentSegments.map((s) => (typeof s === 'string' ? s : s.key))
    );
    return segments
      .filter((s) => !selectedKeys.has(s.key))
      .map((s) => ({
        ...s,
        displayValue: s.name
      }));
  }, [segments, parentSegments]);

  const handleSegmentRemove = (index: number) => {
    segmentRemove(index);
  };

  const handleSegmentSelected = (
    index: number,
    segment: FilterableSegment | null
  ) => {
    if (!segment) return;

    const selectedKeys = new Set<string>(
      parentSegments.map((s) => (typeof s === 'string' ? s : s.key))
    );

    if (selectedKeys.has(segment.key)) return;

    if (index < parentSegments.length) {
      segmentReplace(index, segment);
    } else {
      segmentAdd(segment);
      setShowNewSegmentField(false);
    }
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
              values={segmentOptions}
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
              segmentOptions.length > 0 &&
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
              segmentOptions.length > 0 && (
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
              values={segmentOptions}
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

      {parentSegments.length === 0 && segmentOptions.length > 0 && (
        <div className="flex w-full space-x-1">
          <div className="w-5/6">
            <Combobox<FilterableSegment>
              id="segmentKey-0"
              name="segmentKey-0"
              placeholder="Select or search for a segment"
              values={segmentOptions}
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
