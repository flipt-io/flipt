import { Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import TextButton from '~/components/forms/buttons/TextButton';
import Combobox from '~/components/forms/Combobox';
import Input from '~/components/forms/Input';
import Select from '~/components/forms/Select';
import Loading from '~/components/Loading';
import { updateRollout } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IFlag } from '~/types/Flag';
import { IRollout, RolloutType } from '~/types/Rollout';
import { FilterableSegment, ISegment } from '~/types/Segment';
import { truncateKey } from '~/utils/helpers';

type QuickEditRolloutFormProps = {
  flag: IFlag;
  rollout: IRollout;
  segments: ISegment[];
  onSuccess?: () => void;
};

interface RolloutFormValues {
  segmentKey?: string;
  percentage?: number;
  value: string;
}

export default function QuickEditRolloutForm(props: QuickEditRolloutFormProps) {
  const { onSuccess, flag, rollout, segments } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);

  const [selectedSegment, setSelectedSegment] =
    useState<FilterableSegment | null>(() => {
      let selected =
        segments.find((s) => s.key === rollout.segment?.segmentKey) || null;
      if (selected) {
        return {
          ...selected,
          displayValue: selected.name,
          filterValue: selected.key
        };
      }
      return null;
    });

  const handleSegmentSubmit = (values: RolloutFormValues) => {
    let rolloutSegment = rollout;
    rolloutSegment.threshold = undefined;

    return updateRollout(namespace.key, flag.key, rollout.id, {
      ...rolloutSegment,
      segment: {
        segmentKey: values.segmentKey || '',
        value: values.value === 'true'
      }
    });
  };

  const handleThresholdSubmit = (values: RolloutFormValues) => {
    let rolloutThreshold = rollout;
    rolloutThreshold.segment = undefined;

    return updateRollout(namespace.key, flag.key, rollout.id, {
      ...rolloutThreshold,
      threshold: {
        percentage: values.percentage || 0,
        value: values.value === 'true'
      }
    });
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
        segmentKey: rollout.segment?.segmentKey,
        percentage: rollout.threshold?.percentage,
        value: initialValue
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
        <Form className="flex h-full w-full flex-col overflow-y-scroll bg-white">
          <div className="w-full flex-1">
            <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
              {rollout.type === RolloutType.THRESHOLD ? (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
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
                    className="hidden h-2 w-full cursor-pointer appearance-none self-center rounded-lg align-middle bg-gray-200 dark:bg-gray-700 sm:block"
                  />
                  <div className="relative">
                    <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                      %
                    </div>
                    <Input
                      type="number"
                      id="percentage"
                      max={100}
                      min={0}
                      name="percentage"
                      className="pl-10 text-center"
                    />
                  </div>
                </div>
              ) : (
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
                  <div>
                    <label
                      htmlFor="segmentKey"
                      className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <Combobox<FilterableSegment>
                      id="segmentKey"
                      name="segmentKey"
                      placeholder="Select or search for a segment"
                      values={segments.map((s) => ({
                        ...s,
                        filterValue: truncateKey(s.key),
                        displayValue: s.name
                      }))}
                      selected={selectedSegment}
                      setSelected={setSelectedSegment}
                    />
                  </div>
                </div>
              )}
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2">
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
                  className="w-full cursor-pointer appearance-none self-center rounded-lg py-1 align-middle bg-gray-50"
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
