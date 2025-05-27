import { FieldArray, useFormikContext } from 'formik';
import { useCallback, useEffect, useMemo, useRef } from 'react';

import Input from '~/components/forms/Input';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import Select from '~/components/forms/Select';

import { IFlag } from '~/types/Flag';
import { IRollout, RolloutType } from '~/types/Rollout';
import { FilterableSegment, ISegment, segmentOperators } from '~/types/Segment';

type QuickEditRolloutFormProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
};

export default function QuickEditRolloutForm(props: QuickEditRolloutFormProps) {
  const { rollout, segments } = props;

  // Use useMemo to ensure this is only calculated when dependencies change
  const rolloutSegmentKeys = useMemo(
    () => rollout.segment?.segments || [],
    [rollout.segment?.segments]
  );

  // Create refs for the input fields
  const percentageInputRef = useRef<HTMLInputElement>(null);
  const percentageRangeRef = useRef<HTMLInputElement>(null);

  // Convert segment keys to filterable segments with useMemo to optimize performance
  const rolloutSegments = useMemo(() => {
    return rolloutSegmentKeys.map((s) => {
      const segment = segments.find((seg) => seg.key === s);
      if (!segment) {
        throw new Error(`Segment ${s} not found in segments`);
      }
      return {
        ...segment,
        displayValue: segment.name
      };
    });
  }, [rolloutSegmentKeys, segments]);

  const formik = useFormikContext<IFlag>();

  // Calculate the actual index in the rollouts array by finding this rollout's position
  const rolloutIndex = useMemo(() => {
    const rollouts = formik.values.rollouts || [];
    // Try to find by ID first if available
    if (rollout.id) {
      const index = rollouts.findIndex((r) => r.id === rollout.id);
      if (index !== -1) return index;
    }
    // Fallback to position in array
    return rollouts.indexOf(rollout);
  }, [formik.values.rollouts, rollout]);

  // Use the calculated index for the field path
  const fieldPrefix = `rollouts.[${rolloutIndex !== -1 ? rolloutIndex : 0}].`;

  // Initialize the input fields with the current percentage value
  useEffect(() => {
    if (rollout.type === RolloutType.THRESHOLD && rollout.threshold) {
      if (percentageInputRef.current) {
        percentageInputRef.current.value = String(
          rollout.threshold.percentage || 50
        );
      }
      if (percentageRangeRef.current) {
        percentageRangeRef.current.value = String(
          rollout.threshold.percentage || 50
        );
      }
    }
  }, [rollout]);

  // Manual controlled update function that only updates the form when the blur event happens
  const handlePercentageChange = useCallback(
    (value: string) => {
      const numValue = parseInt(value, 10);
      if (!isNaN(numValue) && numValue >= 0 && numValue <= 100) {
        // Only update Formik when we're sure we have a valid value
        formik.setFieldValue(`${fieldPrefix}threshold.percentage`, numValue);

        // Keep both inputs in sync
        if (percentageInputRef.current) {
          percentageInputRef.current.value = String(numValue);
        }
        if (percentageRangeRef.current) {
          percentageRangeRef.current.value = String(numValue);
        }
      }
    },
    [formik, fieldPrefix]
  );

  // Custom segment management functions
  const handleSegmentAdd = useCallback(
    (segment: FilterableSegment) => {
      // Double-check that we're working with the correct rollout in formik values
      const rollouts = formik.values.rollouts || [];
      if (rolloutIndex === -1 || rolloutIndex >= rollouts.length) {
        return;
      }

      // Get current segments or initialize empty array
      const currentSegments = [...(rollout.segment?.segments || [])];

      // Add the new segment key if it doesn't already exist
      if (!currentSegments.includes(segment.key)) {
        currentSegments.push(segment.key);

        // Update only the segments array within the existing rollout structure
        formik.setFieldValue(`${fieldPrefix}segment.segments`, currentSegments);

        // Ensure the segment object exists and has a value property
        if (!rollout.segment) {
          formik.setFieldValue(`${fieldPrefix}segment`, {
            segments: currentSegments,
            value: false
          });
        }
      }
    },
    [formik, fieldPrefix, rollout, rolloutIndex]
  );

  const handleSegmentRemove = useCallback(
    (index: number) => {
      // Double-check that we're working with the correct rollout in formik values
      const rollouts = formik.values.rollouts || [];
      if (rolloutIndex === -1 || rolloutIndex >= rollouts.length) {
        return;
      }

      // Get current segments
      const currentSegments = [...(rollout.segment?.segments || [])];

      // Remove the segment at the specified index
      if (index >= 0 && index < currentSegments.length) {
        currentSegments.splice(index, 1);

        // Update only the segments array
        formik.setFieldValue(`${fieldPrefix}segment.segments`, currentSegments);
      }
    },
    [formik, fieldPrefix, rollout, rolloutIndex]
  );

  const handleSegmentReplace = useCallback(
    (index: number, segment: FilterableSegment) => {
      // Double-check that we're working with the correct rollout in formik values
      const rollouts = formik.values.rollouts || [];
      if (rolloutIndex === -1 || rolloutIndex >= rollouts.length) {
        return;
      }

      // Get current segments
      const currentSegments = [...(rollout.segment?.segments || [])];

      // Replace the segment at the specified index
      if (index >= 0 && index < currentSegments.length) {
        currentSegments[index] = segment.key;

        // Update only the segments array
        formik.setFieldValue(`${fieldPrefix}segment.segments`, currentSegments);
      }
    },
    [formik, fieldPrefix, rollout, rolloutIndex]
  );

  return (
    <div className="flex h-full w-full flex-col">
      <div className="w-full flex-1">
        <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
          {rollout.type === RolloutType.THRESHOLD ? (
            <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
              <label
                htmlFor="threshold-percentage-range"
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-gray-100"
              >
                Percentage
              </label>
              <Input
                forwardRef={percentageRangeRef}
                id="threshold-percentage-range"
                name="threshold-percentage-range"
                type="range"
                min="0"
                max="100"
                className="hidden h-2 w-full cursor-pointer appearance-none self-center rounded-lg bg-secondary align-middle sm:block"
                defaultValue={String(rollout.threshold?.percentage || 50)}
                onChange={(e) => {
                  // Update the number input when slider changes
                  if (percentageInputRef.current) {
                    percentageInputRef.current.value = e.target.value;
                  }
                }}
                onMouseUp={(e) => handlePercentageChange(e.currentTarget.value)}
                onKeyUp={(e) => handlePercentageChange(e.currentTarget.value)}
              />
              <div className="relative">
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-foreground">
                  %
                </div>
                <Input
                  forwardRef={percentageInputRef}
                  type="number"
                  id="threshold-percentage-input"
                  name="threshold-percentage-input"
                  min="0"
                  max="100"
                  className="text-center pl-7"
                  defaultValue={String(rollout.threshold?.percentage || 50)}
                  onChange={(e) => {
                    // Update the range slider when number input changes
                    if (percentageRangeRef.current) {
                      percentageRangeRef.current.value = e.target.value;
                    }
                  }}
                  onBlur={(e) => handlePercentageChange(e.target.value)}
                />
              </div>
              <label
                htmlFor={fieldPrefix + 'threshold.value'}
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-gray-100"
              >
                Value
              </label>
              <Select
                id={fieldPrefix + 'threshold.value'}
                name={fieldPrefix + 'threshold.value'}
                value={rollout.threshold?.value ? 'true' : 'false'}
                options={[
                  { label: 'True', value: 'true' },
                  { label: 'False', value: 'false' }
                ]}
                onChange={(e) => {
                  e.preventDefault();
                  formik.setFieldValue(
                    `${fieldPrefix}threshold.value`,
                    e.target.value === 'true'
                  );
                }}
                className="w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle"
              />
            </div>
          ) : (
            <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
              <div>
                <label
                  htmlFor={fieldPrefix + 'segment.segments'}
                  className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                >
                  Segment
                </label>
              </div>
              <div className="sm:col-span-2">
                <div>
                  <FieldArray
                    name={fieldPrefix + 'segment.segments'}
                    render={() => (
                      <SegmentsPicker
                        segments={segments}
                        segmentAdd={handleSegmentAdd}
                        segmentRemove={handleSegmentRemove}
                        segmentReplace={handleSegmentReplace}
                        selectedSegments={rolloutSegments}
                      />
                    )}
                  />
                </div>
                {rolloutSegments.length > 1 && (
                  <div className="mt-4 flex space-x-8">
                    {segmentOperators.map((segmentOperator, index) => (
                      <div
                        className="flex items-center space-x-2 cursor-pointer"
                        key={index}
                        onClick={() => {
                          formik.setFieldValue(
                            `${fieldPrefix}segment.segmentOperator`,
                            segmentOperator.id
                          );
                        }}
                      >
                        <div className="flex items-center">
                          <input
                            id={segmentOperator.id}
                            name={fieldPrefix + 'segment.segmentOperator'}
                            type="radio"
                            className="h-4 w-4 border text-ring focus:ring-ring cursor-pointer"
                            checked={
                              segmentOperator.id ===
                              rollout.segment?.segmentOperator
                            }
                            value={segmentOperator.id}
                            readOnly
                          />
                        </div>
                        <div className="flex items-center">
                          <label
                            htmlFor={segmentOperator.id}
                            className="block text-sm text-gray-700 dark:text-gray-200 cursor-pointer"
                          >
                            {segmentOperator.name}{' '}
                            <span className="font-light">
                              {segmentOperator.meta}
                            </span>
                          </label>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
              <label
                htmlFor={fieldPrefix + 'segment.value'}
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-gray-100"
              >
                Value
              </label>
              <Select
                id={fieldPrefix + 'segment.value'}
                name={fieldPrefix + 'segment.value'}
                value={rollout.segment?.value ? 'true' : 'false'}
                options={[
                  { label: 'True', value: 'true' },
                  { label: 'False', value: 'false' }
                ]}
                onChange={(e) => {
                  e.preventDefault();
                  formik.setFieldValue(
                    `${fieldPrefix}segment.value`,
                    e.target.value === 'true'
                  );
                }}
                className="w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle"
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
