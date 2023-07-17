import Combobox from '~/components/forms/Combobox';
import { FilterableVariant, IVariant } from '~/types/Variant';

type SingleDistributionFormInputProps = {
  variants: IVariant[];
  selectedVariant: FilterableVariant | null;
  setSelectedVariant: (variant: FilterableVariant | null) => void;
};

export default function SingleDistributionFormInput(
  props: SingleDistributionFormInputProps
) {
  const { variants, selectedVariant, setSelectedVariant } = props;
  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <label
          htmlFor="variantKey"
          className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
        >
          Variant
        </label>
      </div>
      <div className="sm:col-span-2">
        <Combobox<FilterableVariant>
          id="variant"
          name="variant"
          placeholder="Select or search for a variant"
          values={
            variants?.map((v) => ({
              ...v,
              filterValue: v.key,
              displayValue: v.name
            })) || []
          }
          selected={selectedVariant}
          setSelected={setSelectedVariant}
        />
      </div>
    </div>
  );
}
