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
    <div>
      <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
        <div>
          <label
            htmlFor="variantKey"
            className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
          >
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
              className="block truncate text-right text-sm text-gray-600 sm:mt-px sm:pr-2 sm:pt-2"
            >
              {dist.variantKey}
            </label>
          </div>
          <div className="relative sm:col-span-1">
            <input
              type="number"
              className="block w-full rounded-md border-gray-300 pl-7 pr-12 shadow-sm focus:border-violet-300 focus:ring-violet-300 sm:text-sm"
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
            <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
              <span
                className="text-gray-500 sm:text-sm"
                id={`percentage-${dist.variantKey}`}
              >
                %
              </span>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
