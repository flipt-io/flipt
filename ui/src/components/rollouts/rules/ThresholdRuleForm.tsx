import { useFormikContext } from 'formik';
import Input from '~/components/forms/Input';
import Select from '~/components/forms/Select';

interface ThresholdRuleFormInputsFields {
  percentage: number;
  value: boolean;
}

export default function InnerThresholdRuleFormInputs() {
  const { values, setFieldValue } =
    useFormikContext<ThresholdRuleFormInputsFields>();
  return (
    <>
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
          value={values.percentage}
          onChange={(e) =>
            setFieldValue('percentage', parseInt(e.target.value))
          }
          className="h-2 w-full cursor-pointer appearance-none self-center rounded-lg align-middle bg-gray-200 dark:bg-gray-700"
        />
        <Input
          type="number"
          id="percentage"
          max={100}
          min={0}
          name="percentage"
          value={values.percentage}
          onChange={(e) =>
            setFieldValue('percentage', parseInt(e.target.value))
          }
          className="text-center"
        />
      </div>
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
          value={values.value.toString()}
          onChange={(e) => setFieldValue('value', e.target.value === 'true')}
          options={[
            { label: 'True', value: 'true' },
            { label: 'False', value: 'false' }
          ]}
          className="w-full cursor-pointer appearance-none self-center rounded-lg align-middle bg-gray-200 dark:bg-gray-700"
        />
      </div>
    </>
  );
}
