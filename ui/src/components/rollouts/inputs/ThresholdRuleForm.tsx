import { useFormikContext } from 'formik';
import Input from '~/components/forms/Input';

interface ThresholdRuleFormInputsFields {
  percentage: number;
}

export default function ThresholdRuleFormInputs() {
  const { values, setFieldValue } =
    useFormikContext<ThresholdRuleFormInputsFields>();
  return (
    <>
      <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:p-2 ">
        <label
          htmlFor="percentage"
          className="mb-2 block text-sm font-normal text-gray-900"
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
    </>
  );
}
