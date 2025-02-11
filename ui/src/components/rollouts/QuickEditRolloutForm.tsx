import { FieldArray, useFormikContext } from 'formik';

import { TextButton } from '~/components/Button';
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

  const rolloutSegmentKeys = rollout.segment?.segmentKeys || [];

  const rolloutSegments = rolloutSegmentKeys.map((s) => {
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

  const fieldPrefix = `rollouts.[${rollout.rank}].`;

  const formik = useFormikContext<IFlag>();

  return (
    <div className="flex h-full w-full flex-col">
      <div className="w-full flex-1">
        <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
          {rollout.type === RolloutType.THRESHOLD ? (
            <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
              <label
                htmlFor={fieldPrefix + 'threshold.percentage'}
                className="mb-2 block text-sm font-medium text-gray-900"
              >
                Percentage
              </label>
              <Input
                id={fieldPrefix + 'threshold.percentage'}
                name={fieldPrefix + 'threshold.percentage'}
                type="range"
                className="hidden h-2 w-full cursor-pointer appearance-none self-center rounded-lg bg-gray-200 align-middle dark:bg-gray-700 sm:block"
              />
              <div className="relative">
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-black">
                  %
                </div>
                <Input
                  type="number"
                  id={fieldPrefix + 'threshold.percentage'}
                  max={100}
                  min={0}
                  name={fieldPrefix + 'threshold.percentage'}
                  className="text-center"
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
                  formik.setFieldValue(
                    `${fieldPrefix}threshold.value`,
                    e.target.value == 'true'
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
                    render={(arrayHelpers) => (
                      <SegmentsPicker
                        segments={segments}
                        segmentAdd={(segment: FilterableSegment) =>
                          arrayHelpers.push(segment)
                        }
                        segmentRemove={(index: number) =>
                          arrayHelpers.remove(index)
                        }
                        segmentReplace={(
                          index: number,
                          segment: FilterableSegment
                        ) => arrayHelpers.replace(index, segment)}
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
                            onChange={() => {
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
                  formik.setFieldValue(
                    `${fieldPrefix}segment.value`,
                    e.target.value == 'true'
                  );
                }}
                className="w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle"
              />
            </div>
          )}
        </div>
      </div>
      <div className="flex-shrink-0 py-1">
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
