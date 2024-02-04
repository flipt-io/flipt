import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { FieldArray, Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useCreateRolloutMutation } from '~/app/flags/rolloutsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import Button from '~/components/forms/buttons/Button';
import Input from '~/components/forms/Input';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import Select from '~/components/forms/Select';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { RolloutType } from '~/types/Rollout';
import {
  FilterableSegment,
  ISegment,
  segmentOperators,
  SegmentOperatorType
} from '~/types/Segment';

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
  flagKey: string;
  segments: ISegment[];
  rank: number;
};

interface RolloutFormValues {
  type: string;
  description?: string;
  segmentKeys?: FilterableSegment[];
  operator?: SegmentOperatorType;
  percentage?: number;
  value: string;
}

export default function RolloutForm(props: RolloutFormProps) {
  const { setOpen, onSuccess, flagKey, segments, rank } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);

  const [rolloutRuleType, setRolloutRuleType] = useState(RolloutType.THRESHOLD);
  const [createRollout] = useCreateRolloutMutation();
  const handleSegmentSubmit = (values: RolloutFormValues) => {
    return createRollout({
      namespaceKey: namespace.key,
      flagKey,
      values: {
        rank,
        type: rolloutRuleType,
        description: values.description,
        segment: {
          segmentKeys: values.segmentKeys?.map((s) => s.key),
          segmentOperator: values.operator,
          value: values.value === 'true'
        }
      }
    }).unwrap();
  };

  const handleThresholdSubmit = (values: RolloutFormValues) => {
    return createRollout({
      namespaceKey: namespace.key,
      flagKey,
      values: {
        rank,
        type: rolloutRuleType,
        description: values.description,
        threshold: {
          percentage: values.percentage || 0,
          value: values.value === 'true'
        }
      }
    }).unwrap();
  };

  return (
    <Formik
      enableReinitialize
      initialValues={{
        type: rolloutRuleType,
        description: '',
        segmentKeys: [],
        operator: SegmentOperatorType.OR,
        percentage: 50, // TODO: make this 0?
        value: 'true'
      }}
      validate={(values) => {
        if (values.type === RolloutType.SEGMENT) {
          if (values.segmentKeys.length <= 0) {
            return {
              segmentKeys: true
            };
          }
        } else if (values.type === RolloutType.THRESHOLD) {
          if (values.percentage < 0 || values.percentage > 100) {
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
            setSuccess('Successfully created rollout');
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
        <Form className="bg-white flex h-full flex-col overflow-y-scroll shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-gray-900 text-lg font-medium">
                    New Rollout
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
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
                              className="text-violet-400 border-gray-300 h-4 w-4 focus:ring-violet-400"
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
                              className="text-gray-700 font-medium"
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
              {rolloutRuleType === RolloutType.THRESHOLD && (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <label
                    htmlFor="percentage"
                    className="text-gray-900 mb-2 block text-sm font-medium"
                  >
                    Percentage
                  </label>
                  <Input
                    id="percentage-slider"
                    name="percentage"
                    type="range"
                    className="bg-gray-200 h-2 w-full cursor-pointer appearance-none self-center rounded-lg align-middle dark:bg-gray-700"
                  />
                  <div className="relative">
                    <div className="text-black pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
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
                      className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <FieldArray
                      name="segmentKeys"
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
                          selectedSegments={formik.values.segmentKeys}
                        />
                      )}
                    />
                  </div>
                  {formik.values.segmentKeys.length > 1 && (
                    <>
                      <div>
                        <label
                          htmlFor="operator"
                          className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                        >
                          Operator
                        </label>
                      </div>
                      <div>
                        <div className="sm:col-span-2">
                          <div className="w-48 space-y-4">
                            {segmentOperators.map((segmentOperator, index) => (
                              <div className="flex space-x-4" key={index}>
                                <div>
                                  <input
                                    id={segmentOperator.id}
                                    name="operator"
                                    type="radio"
                                    className="text-violet-400 border-gray-300 h-4 w-4 focus:ring-violet-400"
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
                                    className="text-gray-700 block text-sm"
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
                    </>
                  )}
                </div>
              )}
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <label
                  htmlFor="value"
                  className="text-gray-900 mb-2 block text-sm font-medium"
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                  >
                    Description
                  </label>
                  <span
                    className="text-gray-400 text-xs"
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
          <div className="border-gray-200 flex-shrink-0 border-t px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button onClick={() => setOpen(false)}>Cancel</Button>
              <Button
                variant="primary"
                type="submit"
                className="min-w-[80px]"
                disabled={!formik.isValid || formik.isSubmitting}
              >
                {formik.isSubmitting ? <Loading isPrimary /> : 'Create'}
              </Button>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
}
