import { IDistributionVariant } from '~/types/Distribution';

type MultiDistributionFormInputProps = {
  distributions?: IDistributionVariant[];
  setDistributions: (distributions: IDistributionVariant[] | undefined) => void;
};

export default function MultiDistributionFormInputs(
  props: MultiDistributionFormInputProps
) {
  const { distributions, setDistributions } = props;

  return (
    <div className="sm:pb-4">
      <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
        <div>
          <label htmlFor="variantKey" className="sm:mt-px sm:pt-2">
            Variants
          </label>
        </div>
      </div>
      {distributions?.map((dist, index) => (
        <div
          className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-1"
          key={dist.variantId}
        >
          <div>
            <label
              htmlFor={dist.variantKey}
              className="text-secondary-foreground block truncate text-right text-sm sm:mt-px sm:pt-2 sm:pr-2"
            >
              {dist.variantKey}
            </label>
          </div>
          <div className="relative sm:col-span-1">
            <div className="text-muted-foreground pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
              %
            </div>
            <input
              type="number"
              className="text-secondary-foreground bg-secondary block w-full rounded-md border-gray-300 pl-10 shadow-xs focus:border-violet-300 focus:ring-violet-300 sm:text-sm"
              value={dist.rollout}
              name={dist.variantKey}
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
