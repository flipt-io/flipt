import Percent from '~/components/Percent';
import { Input } from '~/components/ui/input';

import { IDistribution } from '~/types/Distribution';

type MultiDistributionFormInputProps = {
  distributions: IDistribution[];
  setDistributions: (distributions: IDistribution[]) => void;
};

export default function MultiDistributionFormInputs(
  props: MultiDistributionFormInputProps
) {
  const { distributions, setDistributions } = props;

  return (
    <div className="sm:pb-4">
      <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
        <div>
          <label
            htmlFor="variantKey"
            className="block text-sm font-medium sm:mt-px sm:pt-2"
          >
            Variants
          </label>
        </div>
      </div>
      {distributions.map((dist: IDistribution, index: number) => (
        <div
          className="flex space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-1"
          key={dist.variant}
        >
          <label
            htmlFor={dist.variant}
            className="block truncate sm:text-right text-sm text-secondary-foreground sm:mt-px sm:pr-2 sm:pt-2"
          >
            {dist.variant}
          </label>
          <div className="relative sm:col-span-1">
            <Percent />
            <Input
              type="number"
              className="pl-10 text-center"
              value={dist.rollout}
              name={dist.variant}
              data-testid="distribution-input"
              // eslint-disable-next-line react/no-unknown-property
              typeof="number"
              step=".01"
              min="0"
              max="100"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                const newDistributions = [...distributions];
                newDistributions[index].rollout = parseFloat(e.target.value);
                setDistributions(newDistributions);
              }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
