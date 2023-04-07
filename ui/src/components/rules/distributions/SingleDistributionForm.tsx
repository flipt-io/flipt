import Combobox from '~/components/forms/Combobox';
import { IVariant, SelectableVariant } from '~/types/Variant';

type SingleDistributionFormInputProps = {
  variants: IVariant[];
  selectedVariant: SelectableVariant | null;
  setSelectedVariant: (variant: SelectableVariant | null) => void;
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
          className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
        >
          Variant
        </label>
      </div>
      <div className="sm:col-span-2">
        <Combobox<SelectableVariant>
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
