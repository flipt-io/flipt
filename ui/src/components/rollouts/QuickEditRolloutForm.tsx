import { FieldArray, useFormikContext } from 'formik';
import { useCallback, useEffect, useMemo, useRef } from 'react';

import { TextButton } from '~/components/Button';
import Input from '~/components/forms/Input';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import Select from '~/components/forms/Select';

import { IFlag } from '~/types/Flag';
import { IRollout, RolloutType } from '~/types/Rollout';
import { ISegment, segmentOperators } from '~/types/Segment';

import { createSegmentHandlers } from '~/utils/formik-helpers';

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
        displayValue: segment.name,
        filterValue: segment.key
      };
    });
  }, [rolloutSegmentKeys, segments]);

  const formik = useFormikContext<IFlag>();
  const fieldPrefix = `rollouts.[${rollout.rank}].`;

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

  // Use the shared segment handlers with rollout-specific update function
  const { handleSegmentAdd, handleSegmentRemove, handleSegmentReplace } =
    createSegmentHandlers<IRollout, IFlag>(
      formik,
      rollout,
      'rollouts',
      (original, segments) => ({
        ...original,
        segment: {
          ...(original.segment || {}),
          segments,
          // Ensure value is always a boolean to satisfy the type system
          value:
            original.segment?.value !== undefined
              ? original.segment.value
              : false
        }
      })
    );

  return (
    <div className="flex h-full w-full flex-col">
      <div className="w-full flex-1">
        <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
          {rollout.type === RolloutType.THRESHOLD ? (
            <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
              <label
                htmlFor="threshold-percentage-range"
                className="mb-2 block text-sm font-medium text-gray-900"
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
                className="hidden h-2 w-full cursor-pointer appearance-none self-center rounded-lg bg-gray-200 align-middle dark:bg-gray-700 sm:block"
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
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-black">
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
                className="mb-2 block text-sm font-medium text-gray-900"
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
                  className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
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
                <div className="mt-6 flex space-x-8">
                  {rolloutSegments.length > 1 &&
                    segmentOperators.map((segmentOperator, index) => (
                      <div className="flex space-x-2" key={index}>
                        <div>
                          <input
                            id={segmentOperator.id}
                            name={fieldPrefix + 'segment.segmentOperator'}
                            type="radio"
                            className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                            onChange={(e) => {
                              e.preventDefault();
                              formik.setFieldValue(
                                fieldPrefix + 'segment.segmentOperator',
                                segmentOperator.id
                              );
                            }}
                            checked={
                              segmentOperator.id ===
                              rollout.segment?.segmentOperator
                            }
                            value={segmentOperator.id}
                          />
                        </div>
                        <div>
                          <label
                            htmlFor={segmentOperator.id}
                            className="block text-sm text-gray-700"
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
              </div>
              <label
                htmlFor={fieldPrefix + 'segment.value'}
                className="mb-2 block text-sm font-medium text-gray-900"
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
      <div className="shrink-0 py-1">
        <div className="flex justify-end space-x-3">
          <TextButton
            disabled={formik.isSubmitting || !formik.dirty}
            onClick={() => formik.resetForm()}
          >
            Reset
          </TextButton>
        </div>
      </div>
    </div>
  );
}
