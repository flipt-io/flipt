import { useCallback, useEffect } from 'react';

import Combobox from '~/components/Combobox';

import { FilterableVariant, IVariant } from '~/types/Variant';

type SingleDistributionFormInputProps = {
  variants: IVariant[];
  selectedVariant: FilterableVariant | null;
  setSelectedVariant: (variant: FilterableVariant | null) => void;
  id?: string;
};

export default function SingleDistributionFormInput(
  props: SingleDistributionFormInputProps
) {
  const { variants, selectedVariant, setSelectedVariant } = props;
  const id = props.id || 'variant';

  // Create a formatted variants array for the Combobox
  const formattedVariants = variants.map((v) => ({
    ...v,
    filterValue: v.key,
    displayValue: v.name
  }));

  // Synchronize with parent's selected variant when it changes externally
  useEffect(() => {
    // This ensures the component stays in sync with parent state
    if (selectedVariant) {
      // Check if selected variant exists in available variants
      const found = formattedVariants.find(
        (v) => v.key === selectedVariant.key
      );
      // If variant no longer exists in available variants, clear selection
      if (!found && setSelectedVariant) {
        setSelectedVariant(null);
      }
    }
  }, [formattedVariants, selectedVariant, setSelectedVariant]);

  // Handler for variant selection
  const handleSelect = useCallback(
    (variant: FilterableVariant | null) => {
      // Immediately update parent component
      setSelectedVariant(variant);
    },
    [setSelectedVariant]
  );

  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <label
          htmlFor="variant"
          className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
        >
          Variant
        </label>
      </div>
      <div className="sm:col-span-2">
        <Combobox<FilterableVariant>
          id={id}
          name="variant"
          placeholder="Select or search for a variant"
          values={formattedVariants}
          selected={selectedVariant}
          setSelected={handleSelect}
        />
      </div>
    </div>
  );
}
