import { IRolloutRuleThreshold } from '~/types/Rollout';

type ThresholdRuleFormInputProps = {
  rule: IRolloutRuleThreshold;
};

export default function ThresholdRuleFormInputs(
  props: ThresholdRuleFormInputProps
) {
  const { rule } = props;

  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <label
        htmlFor="percentage"
        className="mb-2 block text-sm font-medium text-gray-900"
      >
        Percentage
      </label>
      <input
        id="percentage"
        type="range"
        defaultValue="50"
        value={rule?.percentage}
        className="h-2 w-full cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700"
      />
    </div>
  );
}
