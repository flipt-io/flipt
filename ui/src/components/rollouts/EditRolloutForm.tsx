import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { FieldArray, Form, Formik } from 'formik';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import Input from '~/components/forms/Input';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import Select from '~/components/forms/Select';

import { IRollout, RolloutType } from '~/types/Rollout';
import {
  FilterableSegment,
  ISegment,
  SegmentOperatorType,
  segmentOperators
} from '~/types/Segment';

import { useError } from '~/data/hooks/error';
import { cls } from '~/utils/helpers';

const rolloutRuleTypes = [
  {
    id: RolloutType.SEGMENT,
    name: 'Segment',
    description: 'Rollout to a specific segment'
  },
  {
    id: RolloutType.THRESHOLD,
    name: 'Threshold',
    description: 'Rollout to a percentage of entities'
  }
];

type EditRolloutFormProps = {
  setOpen: (open: boolean) => void;
  onSuccess: () => void;
  updateRollout: (rollout: IRollout) => void;
  rollout: IRollout;
  segments: ISegment[];
};

interface RolloutFormValues {
  description: string;
  operator?: SegmentOperatorType;
  percentage?: number;
  segments?: FilterableSegment[];
  value: string;
}

export default function EditRolloutForm(props: EditRolloutFormProps) {
  const { setOpen, onSuccess, rollout, segments, updateRollout } = props;

  const { setError, clearError } = useError();

  const segmentOperator =
    rollout.segment && rollout.segment.segmentOperator
      ? rollout.segment.segmentOperator
      : SegmentOperatorType.OR;

  const handleSegmentSubmit = (values: RolloutFormValues) => {
    let rolloutSegment = rollout;
    rolloutSegment.threshold = undefined;

    updateRollout({
      ...rolloutSegment,
      description: values.description,
      segment: {
        segments: values.segments?.map((s) => s.key),
        segmentOperator: values.operator,
        value: values.value === 'true'
      }
    });

    return Promise.resolve();
  };

  const handleThresholdSubmit = (values: RolloutFormValues) => {
    let rolloutThreshold = rollout;
    rolloutThreshold.segment = undefined;

    updateRollout({
      ...rolloutThreshold,
      description: values.description,
      threshold: {
        percentage: values.percentage || 0,
        value: values.value === 'true'
      }
    });

    return Promise.resolve();
  };

  const initialValue =
    rollout.type === RolloutType.THRESHOLD
      ? rollout.threshold?.value
        ? 'true'
        : 'false'
      : rollout.segment?.value
        ? 'true'
        : 'false';

  return (
    <Formik
      enableReinitialize
      initialValues={{
        type: rollout.type,
        description: rollout.description || '',
        segments: segments.flatMap((s) =>
          rollout.segment?.segments?.includes(s.key)
            ? {
                ...s,
                displayValue: s.name,
                filterValue: s.key
              }
            : []
        ),
        operator: segmentOperator,
        percentage: rollout.threshold?.percentage,
        value: initialValue
      }}
      validate={(values) => {
        if (values.type === RolloutType.SEGMENT) {
          if (values.segments.length <= 0) {
            return {
              segments: true
            };
          }
        } else if (values.type === RolloutType.THRESHOLD) {
          if (
            values.percentage &&
            (values.percentage < 0 || values.percentage > 100)
          ) {
            return {
              percentage: true
            };
          }
        }
      }}
      onSubmit={(values, { setSubmitting }) => {
        let handleSubmit = async (_values: RolloutFormValues) => {};

        if (rollout.type === RolloutType.SEGMENT) {
          handleSubmit = handleSegmentSubmit;
        } else if (rollout.type === RolloutType.THRESHOLD) {
          handleSubmit = handleThresholdSubmit;
        }

        handleSubmit(values)
          .then(() => {
            onSuccess();
            clearError();
            setOpen(false);
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
    >
      {(formik) => (
        <Form className="flex h-full flex-col overflow-y-scroll bg-background shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-lg font-medium text-gray-900">
                    Edit Rollout
                  </Dialog.Title>
                  <MoreInfo href="https://www.flipt.io/docs/concepts#rollouts">
                    Learn more about rollouts
                  </MoreInfo>
                </div>
                <div className="flex h-7 items-center">
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500"
                    onClick={() => setOpen(false)}
                  >
                    <span className="sr-only">Close panel</span>
                    <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>
            <div className="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="type"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Type
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <fieldset>
                    <legend className="sr-only">Type</legend>
                    <div className="space-y-5">
                      {rolloutRuleTypes.map((rolloutRule) => (
                        <div
                          key={rolloutRule.id}
                          className="relative flex items-start"
                        >
                          <div className="flex h-5 items-center">
                            <input
                              id={rolloutRule.id}
                              aria-describedby={`${rolloutRule.id}-description`}
                              name="type"
                              type="radio"
                              className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                              disabled={true}
                              checked={rolloutRule.id === rollout.type}
                              value={rolloutRule.id}
                            />
                          </div>
                          <div className="ml-3 text-sm">
                            <label
                              htmlFor={rolloutRule.id}
                              className="font-medium text-gray-700"
                            >
                              {rolloutRule.name}
                            </label>
                            <p
                              id={`${rolloutRule.id}-description`}
                              className="text-gray-500"
                            >
                              {rolloutRule.description}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </fieldset>
                </div>
              </div>
              {rollout.type === RolloutType.THRESHOLD && (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <label
                    htmlFor="percentage"
                    className="mb-2 block text-sm font-medium text-gray-900"
                  >
                    Percentage
                  </label>
                  <Input
                    id="percentage-slider"
                    name="percentage"
                    type="range"
                    className="h-2 w-full cursor-pointer appearance-none self-center rounded-lg bg-gray-200 align-middle dark:bg-gray-700"
                  />
                  <div className="relative">
                    <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-black">
                      %
                    </div>
                    <Input
                      type="number"
                      id="percentage"
                      max={100}
                      min={0}
                      name="percentage"
                      className="text-center"
                    />
                  </div>
                </div>
              )}
              {rollout.type === RolloutType.SEGMENT && (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <div>
                    <label
                      htmlFor="segments"
                      className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <div>
                      <FieldArray
                        name="segments"
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
                            selectedSegments={formik.values.segments}
                          />
                        )}
                      />
                    </div>
                    <div className="mt-6 flex space-x-8">
                      {formik.values.segments.length > 1 &&
                        segmentOperators.map((segmentOperator, index) => (
                          <div className="flex space-x-2" key={index}>
                            <div>
                              <input
                                id={segmentOperator.id}
                                name="operator"
                                type="radio"
                                className={cls(
                                  'h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400'
                                )}
                                onChange={() => {
                                  formik.setFieldValue(
                                    'operator',
                                    segmentOperator.id
                                  );
                                }}
                                checked={
                                  segmentOperator.id === formik.values.operator
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
                </div>
              )}
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <label
                  htmlFor="value"
                  className="mb-2 block text-sm font-medium text-gray-900"
                >
                  Value
                </label>
                <Select
                  id="value"
                  name="value"
                  value={formik.values.value}
                  options={[
                    { label: 'True', value: 'true' },
                    { label: 'False', value: 'false' }
                  ]}
                  className="w-full cursor-pointer appearance-none self-center rounded-lg align-middle"
                />
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="description"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Description
                  </label>
                  <span
                    className="text-xs text-gray-400"
                    id="description-optional"
                  >
                    Optional
                  </span>
                </div>
                <div className="sm:col-span-2">
                  <Input name="description" id="description" />
                </div>
              </div>
            </div>
          </div>
          <div className="flex-shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button onClick={() => setOpen(false)}>Cancel</Button>
              <Button
                variant="primary"
                type="submit"
                className="min-w-[80px]"
                disabled={!formik.isValid || formik.isSubmitting}
              >
                {formik.isSubmitting ? <Loading isPrimary /> : 'Done'}
              </Button>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
}
