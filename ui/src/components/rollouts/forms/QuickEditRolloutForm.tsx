import { FieldArray, Form, Formik } from 'formik';
import { useSelector } from 'react-redux';
import { useUpdateRolloutMutation } from '~/app/flags/rolloutsApi';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import TextButton from '~/components/forms/buttons/TextButton';
import Input from '~/components/forms/Input';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import Select from '~/components/forms/Select';
import Loading from '~/components/Loading';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IFlag } from '~/types/Flag';
import { IRollout, RolloutType } from '~/types/Rollout';
import {
  FilterableSegment,
  ISegment,
  segmentOperators,
  SegmentOperatorType
} from '~/types/Segment';
import { cls } from '~/utils/helpers';

type QuickEditRolloutFormProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
};

interface RolloutFormValues {
  operator?: SegmentOperatorType;
  percentage?: number;
  segmentKeys?: FilterableSegment[];
  value: string;
}

export default function QuickEditRolloutForm(props: QuickEditRolloutFormProps) {
  const { onSuccess, flag, rollout, segments } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);

  const segmentOperator =
    rollout.segment && rollout.segment.segmentOperator
      ? rollout.segment.segmentOperator
      : SegmentOperatorType.OR;

  const readOnly = useSelector(selectReadonly);
  const [updateRollout] = useUpdateRolloutMutation();
  const handleSegmentSubmit = (values: RolloutFormValues) => {
    let rolloutSegment = rollout;
    rolloutSegment.threshold = undefined;

    return updateRollout({
      namespaceKey: namespace.key,
      flagKey: flag.key,
      rolloutId: rollout.id,
      values: {
        ...rolloutSegment,
        segment: {
          segmentKeys: values.segmentKeys?.map((s) => s.key),
          segmentOperator: values.operator,
          value: values.value === 'true'
        }
      }
    }).unwrap();
  };

  const handleThresholdSubmit = (values: RolloutFormValues) => {
    let rolloutThreshold = rollout;
    rolloutThreshold.segment = undefined;

    return updateRollout({
      namespaceKey: namespace.key,
      flagKey: flag.key,
      rolloutId: rollout.id,
      values: {
        ...rolloutThreshold,
        threshold: {
          percentage: values.percentage || 0,
          value: values.value === 'true'
        }
      }
    }).unwrap();
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
        segmentKeys: segments.flatMap((s) =>
          rollout.segment?.segmentKeys?.includes(s.key)
            ? {
                ...s,
                displayValue: s.name,
                filterValue: s.key
              }
            : []
        ),
        percentage: rollout.threshold?.percentage,
        value: initialValue,
        operator: segmentOperator
      }}
      validate={(values) => {
        if (values.type === RolloutType.SEGMENT) {
          if (values.segmentKeys.length <= 0) {
            return {
              segmentKeys: true
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

        if (rollout.type === RolloutType.THRESHOLD) {
          handleSubmit = handleThresholdSubmit;
        } else {
          handleSubmit = handleSegmentSubmit;
        }

        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess('Successfully updated rollout');
            onSuccess && onSuccess();
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
        <Form className="bg-white flex h-full w-full flex-col overflow-y-scroll">
          <div className="w-full flex-1">
            <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
              {rollout.type === RolloutType.THRESHOLD ? (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
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
                    className="bg-gray-200 hidden h-2 w-full cursor-pointer appearance-none self-center rounded-lg align-middle dark:bg-gray-700 sm:block"
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
              ) : (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
                  <div>
                    <label
                      htmlFor="segmentKeys"
                      className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <div>
                      <FieldArray
                        name="segmentKeys"
                        render={(arrayHelpers) => (
                          <SegmentsPicker
                            readonly={readOnly}
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
                    <div className="mt-6 flex space-x-8">
                      {formik.values.segmentKeys.length > 1 &&
                        segmentOperators.map((segmentOperator, index) => (
                          <div className="flex space-x-2" key={index}>
                            <div>
                              <input
                                id={segmentOperator.id}
                                name="operator"
                                type="radio"
                                className={cls(
                                  'text-violet-400 border-gray-300 h-4 w-4 focus:ring-violet-400',
                                  { 'cursor-not-allowed': readOnly }
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
                                disabled={readOnly}
                                title={
                                  readOnly
                                    ? 'Not allowed in Read-Only mode'
                                    : undefined
                                }
                              />
                            </div>
                            <div>
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
              )}
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
                <label
                  htmlFor="value"
                  className="text-gray-900 mb-2 block text-sm font-medium"
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
                  className={cls(
                    'w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle',
                    { 'text-gray-500 bg-gray-100 cursor-not-allowed': readOnly }
                  )}
                  disabled={readOnly}
                />
              </div>
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
              <TextButton
                type="submit"
                className="min-w-[80px]"
                disabled={
                  !formik.isValid || formik.isSubmitting || !formik.dirty
                }
              >
                {formik.isSubmitting ? <Loading isPrimary /> : 'Update'}
              </TextButton>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
}
