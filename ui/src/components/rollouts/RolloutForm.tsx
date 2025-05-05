import { XMarkIcon } from '@heroicons/react/24/outline';
import * as Dialog from '@radix-ui/react-dialog';
import { FieldArray, Form, Formik } from 'formik';
import { useCallback, useState } from 'react';

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

type RolloutFormProps = {
  setOpen: (open: boolean) => void;
  onSuccess: () => void;
  createRollout: (rollout: IRollout) => void;
  segments: ISegment[];
};

interface RolloutFormValues {
  type: string;
  description?: string;
  segments: FilterableSegment[];
  operator?: SegmentOperatorType;
  percentage?: number;
  value: string;
}

export default function RolloutForm(props: RolloutFormProps) {
  const { setOpen, onSuccess, segments, createRollout } = props;

  const { setError, clearError } = useError();

  const [rolloutRuleType, setRolloutRuleType] = useState(RolloutType.THRESHOLD);

  const handleSegmentSubmit = useCallback(
    (values: RolloutFormValues) => {
      createRollout({
        type: rolloutRuleType,
        description: values.description,
        segment: {
          segments: values.segments?.map((s) => s.key) || [],
          segmentOperator: values.operator,
          value: values.value === 'true'
        }
      });
      return Promise.resolve();
    },
    [createRollout, rolloutRuleType]
  );

  const handleThresholdSubmit = useCallback(
    (values: RolloutFormValues) => {
      createRollout({
        type: rolloutRuleType,
        description: values.description,
        threshold: {
          percentage: values.percentage || 0,
          value: values.value === 'true'
        }
      });
      return Promise.resolve();
    },
    [createRollout, rolloutRuleType]
  );

  return (
    <Formik<RolloutFormValues>
      enableReinitialize
      initialValues={{
        type: rolloutRuleType,
        description: '',
        segments: [],
        operator: SegmentOperatorType.OR,
        percentage: 50, // TODO: make this 0?
        value: 'true'
      }}
      validate={(values) => {
        if (values.type === RolloutType.SEGMENT) {
          if (values.segments.length <= 0) {
            return {
              segments: true
            };
          }
        } else if (values.type === RolloutType.THRESHOLD) {
          const percentage = values.percentage ?? 0;
          if (percentage < 0 || percentage > 100) {
            return {
              percentage: true
            };
          }
        }
      }}
      onSubmit={(values, { setSubmitting }) => {
        let handleSubmit = async (_values: RolloutFormValues) => {};

        if (rolloutRuleType === RolloutType.SEGMENT) {
          handleSubmit = handleSegmentSubmit;
        } else if (rolloutRuleType === RolloutType.THRESHOLD) {
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
      {(formik) => {
        // Helper function to handle segments with proper object representations
        const updateSegmentsValue = (segments: FilterableSegment[]) => {
          formik.setFieldValue('segments', segments);
        };

        // Custom segment handlers that work with the FormValues type
        const addSegment = (segment: FilterableSegment) => {
          const updatedSegments = [...(formik.values.segments || []), segment];
          updateSegmentsValue(updatedSegments);
        };

        const removeSegment = (index: number) => {
          const updatedSegments = [...(formik.values.segments || [])];
          updatedSegments.splice(index, 1);
          updateSegmentsValue(updatedSegments);
        };

        const replaceSegment = (index: number, segment: FilterableSegment) => {
          const updatedSegments = [...(formik.values.segments || [])];
          updatedSegments[index] = segment;
          updateSegmentsValue(updatedSegments);
        };

        return (
          <Form className="flex h-full flex-col overflow-y-scroll bg-background dark:bg-gray-900 shadow-xl">
            <div className="flex-1">
              <div className="bg-gray-50 dark:bg-gray-800 px-4 py-6 sm:px-6">
                <div className="flex items-start justify-between space-x-3">
                  <div className="space-y-1">
                    <Dialog.Title className="text-lg font-medium text-gray-900 dark:text-gray-100">
                      New Rollout
                    </Dialog.Title>
                    <MoreInfo href="https://www.flipt.io/docs/concepts#rollouts">
                      Learn more about rollouts
                    </MoreInfo>
                  </div>
                  <div className="flex h-7 items-center">
                    <button
                      type="button"
                      className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
                      onClick={() => setOpen(false)}
                    >
                      <span className="sr-only">Close panel</span>
                      <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                    </button>
                  </div>
                </div>
              </div>
              <div className="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 dark:sm:divide-gray-700 sm:py-0">
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <div>
                    <label
                      htmlFor="type"
                      className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
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
                                onChange={() => {
                                  setRolloutRuleType(rolloutRule.id);
                                  formik.setFieldValue('type', rolloutRule.id);
                                }}
                                checked={rolloutRule.id === rolloutRuleType}
                                value={rolloutRule.id}
                              />
                            </div>
                            <div className="ml-3 text-sm">
                              <label
                                htmlFor={rolloutRule.id}
                                className="font-medium text-gray-700 dark:text-gray-300"
                              >
                                {rolloutRule.name}
                              </label>
                              <p
                                id={`${rolloutRule.id}-description`}
                                className="text-gray-500 dark:text-gray-400"
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
                {rolloutRuleType === RolloutType.THRESHOLD && (
                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                    <label
                      htmlFor="percentage"
                      className="mb-2 block text-sm font-medium text-gray-900 dark:text-gray-100"
                    >
                      Percentage
                    </label>
                    <Input
                      id="percentage-slider"
                      name="percentage"
                      type="range"
                      className="h-2 w-full cursor-pointer appearance-none self-center rounded-lg bg-gray-200 align-middle dark:bg-gray-600"
                    />
                    <div className="relative">
                      <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-black dark:text-white">
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
                {rolloutRuleType === RolloutType.SEGMENT && (
                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                    <div>
                      <label
                        htmlFor="segmentKey"
                        className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                      >
                        Segment
                      </label>
                    </div>
                    <div className="sm:col-span-2">
                      <FieldArray
                        name="segments"
                        render={() => (
                          <SegmentsPicker
                            segments={segments}
                            segmentAdd={addSegment}
                            segmentRemove={removeSegment}
                            segmentReplace={replaceSegment}
                            selectedSegments={formik.values.segments || []}
                          />
                        )}
                      />
                    </div>
                    {formik.values.segments &&
                      formik.values.segments.length > 1 && (
                        <>
                          <div>
                            <label
                              htmlFor="operator"
                              className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                            >
                              Operator
                            </label>
                          </div>
                          <div>
                            <div className="sm:col-span-2">
                              <div className="w-48 space-y-4">
                                {segmentOperators.map(
                                  (segmentOperator, index) => (
                                    <div className="flex space-x-4" key={index}>
                                      <div>
                                        <input
                                          id={segmentOperator.id}
                                          name="operator"
                                          type="radio"
                                          className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                                          onChange={() => {
                                            formik.setFieldValue(
                                              'operator',
                                              segmentOperator.id
                                            );
                                          }}
                                          checked={
                                            segmentOperator.id ===
                                            formik.values.operator
                                          }
                                          value={segmentOperator.id}
                                        />
                                      </div>
                                      <div className="mt-1">
                                        <label
                                          htmlFor={segmentOperator.id}
                                          className="block text-sm text-gray-700 dark:text-gray-300"
                                        >
                                          {segmentOperator.name}{' '}
                                          <span className="font-light dark:text-gray-400">
                                            {segmentOperator.meta}
                                          </span>
                                        </label>
                                      </div>
                                    </div>
                                  )
                                )}
                              </div>
                            </div>
                          </div>
                        </>
                      )}
                  </div>
                )}
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <label
                    htmlFor="value"
                    className="mb-2 block text-sm font-medium text-gray-900 dark:text-gray-100"
                  >
                    Value
                  </label>
                  <Select
                    id="value"
                    name="value"
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
                      className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                    >
                      Description
                    </label>
                    <span
                      className="text-xs text-gray-400 dark:text-gray-500"
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
            <div className="shrink-0 border-t border-gray-200 dark:border-gray-700 px-4 py-5 sm:px-6">
              <div className="flex justify-end space-x-3">
                <Button variant="secondary" onClick={() => setOpen(false)}>Cancel</Button>
                <Button
                  variant="primary"
                  type="submit"
                  className="min-w-[80px]"
                  disabled={!formik.isValid || formik.isSubmitting}
                >
                  {formik.isSubmitting ? <Loading isPrimary /> : 'Add'}
                </Button>
              </div>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
}
