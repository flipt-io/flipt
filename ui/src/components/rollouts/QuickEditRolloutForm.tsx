import { Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import Select from '~/components/forms/Select';
import { createRollout } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IRollout, RolloutType, rolloutTypeToLabel } from '~/types/Rollout';
import { FilterableSegment, ISegment } from '~/types/Segment';
import TextButton from '../forms/buttons/TextButton';
import Loading from '../Loading';
import SegmentRuleFormInputs from './inputs/SegmentRuleForm';
import ThresholdRuleFormInputs from './inputs/ThresholdRuleForm';

const rolloutRuleTypeSegment = 'SEGMENT_ROLLOUT_TYPE';
const rolloutRuleTypeThreshold = 'THRESHOLD_ROLLOUT_TYPE';

type QuickEditRolloutFormProps = {
  onSuccess: () => void;
  flagKey: string;
  rollout: IRollout;
  segments: ISegment[];
};

interface RolloutFormValues {
  segmentKey?: string;
  percentage?: number;
  value: string;
}

export default function QuickEditRolloutForm(props: QuickEditRolloutFormProps) {
  const { onSuccess, flagKey, rollout, segments } = props;

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
    return createRollout(namespace.key, flagKey, {
      ...rollout,
      segment: {
        segmentKey: values.segmentKey || '',
        value: values.value === 'true'
      }
    });
  };

  const handleThresholdSubmit = (values: RolloutFormValues) => {
    return createRollout(namespace.key, flagKey, {
      ...rollout,
      threshold: {
        percentage: values.percentage || 0,
        value: values.value === 'true'
      }
    });
  };

  return (
    <Formik
      enableReinitialize
      initialValues={{
        type: rollout.type,
        segmentKey: rollout.segment?.segmentKey,
        percentage: rollout.threshold?.percentage,
        value: rollout.segment?.value || rollout.threshold?.value ? 'true' : ''
      }}
      onSubmit={(values, { setSubmitting }) => {
        let handleSubmit = async (_values: RolloutFormValues) => {};

        if (rollout.type === RolloutType.THRESHOLD_ROLLOUT_TYPE) {
          handleSubmit = handleThresholdSubmit;
        } else {
          handleSubmit = handleSegmentSubmit;
        }

        handleSubmit(values)
          .then(() => {
            onSuccess();
            clearError();
            setSuccess('Successfully updated rollout');
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
              {rolloutTypeToLabel(rollout.type) ===
              RolloutType.THRESHOLD_ROLLOUT_TYPE ? (
                <ThresholdRuleFormInputs />
              ) : (
                <SegmentRuleFormInputs
                  segments={segments}
                  selectedSegment={selectedSegment}
                  setSelectedSegment={setSelectedSegment}
                />
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
              <TextButton onClick={() => {}}>Reset</TextButton>
              <TextButton
                type="submit"
                className="min-w-[80px]"
                disabled={!formik.isValid || formik.isSubmitting}
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
