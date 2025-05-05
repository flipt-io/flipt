import { Combobox as C } from '@headlessui/react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/24/outline';
import { useField } from 'formik';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { IFilterable } from '~/types/Selectable';

import { cls } from '~/utils/helpers';

type ComboboxProps<T extends IFilterable> = {
  id: string;
  name: string;
  placeholder?: string;
  values?: T[];
  selected: T | null;
  setSelected?: (v: T | null) => void;
  disabled?: boolean;
  className?: string;
  inputClassName?: string;
};

export default function Combobox<T extends IFilterable>(
  props: ComboboxProps<T>
) {
  const {
    id,
    name,
    className,
    inputClassName,
    values,
    selected,
    setSelected,
    placeholder,
    disabled
  } = props;

  const [query, setQuery] = useState('');
  const [field] = useField(props);
  const [openOptions, setOpenOptions] = useState(false);

  // Reset query when selected value changes to prevent incorrect display
  useEffect(() => {
    if (selected) {
      setQuery('');
      setOpenOptions(false);
    }
  }, [selected]);

  const filteredValues = useMemo(() => {
    // If there's a query, filter the values
    if (query.trim() !== '') {
      return values?.filter((v) =>
        v.filterValue.toLowerCase().includes(query.toLowerCase())
      );
    }
    // Otherwise, return all values
    return values;
  }, [values, query]);

  const handleChange = useCallback(
    (v: T | null) => {
      // Close dropdown when a value is selected
      setOpenOptions(false);

      // Update parent component state
      setSelected && setSelected(v);

      // Handle both cases: when v is a full object or just a string key
      const value = v ? (typeof v === 'string' ? v : v.key) : null;
      field.onChange({ target: { value, id } });

      // Reset the query when an item is selected
      setQuery('');
    },
    [setSelected, field, id]
  );

  const handleFocus = useCallback(() => {
    // Show options on focus
    setOpenOptions(true);
  }, []);

  const handleBlur = useCallback(() => {
    // Use timeout to allow click events to complete before closing
    setTimeout(() => {
      setOpenOptions(false);
      // Reset query on blur if we have a selected value
      // This ensures the display shows the selected value again
      if (selected) {
        setQuery('');
      }
    }, 100);
  }, [selected]);

  const handleQueryChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setQuery(e.target.value);
    },
    []
  );

  // Enhanced displayValue function that falls back to filterValue and key
  const displayValue = useCallback((v: T | null): string => {
    if (!v) return '';
    return v.displayValue || v.filterValue || v.key || '';
  }, []);

  return (
    <div className="relative w-full">
      <C
        as="div"
        className={className}
        value={selected}
        onChange={handleChange}
        disabled={disabled}
        nullable
      >
        {({ open }) => (
          <div onFocus={handleFocus} onBlur={handleBlur}>
            <div className="relative flex w-full flex-row">
              <C.Input
                className={cls(
                  'w-full rounded-md border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-900 py-2 pl-3 pr-10 text-gray-900 dark:text-gray-100 shadow-xs sm:text-sm',
                  inputClassName
                )}
                onChange={handleQueryChange}
                displayValue={displayValue}
                placeholder={placeholder}
                name={name}
                id={`${id}-select-input`}
                autoComplete="off"
              />
              <C.Button
                className="absolute -inset-y-0 right-0 items-center rounded-r-md px-2"
                id={`${id}-select-button`}
                data-testid={`${id}-select-button`}
              >
                <ChevronUpDownIcon
                  className="h-5 w-5 text-gray-400 dark:text-gray-300"
                  aria-hidden="true"
                />
              </C.Button>
            </div>
            {open && (
              <div className="relative">
                <C.Options
                  className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white dark:bg-gray-800 py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 dark:ring-gray-700 dark:ring-opacity-20 sm:text-sm"
                  id={`${id}-select-options`}
                  static={openOptions}
                >
                  {filteredValues &&
                    filteredValues.map((v) => (
                      <C.Option
                        key={v?.key}
                        value={v}
                        className={({ active }) =>
                          cls(
                            'relative w-full cursor-default select-none py-2 pl-3 pr-9',
                            {
                              'bg-violet-300 dark:bg-violet-600 text-gray-900 dark:text-white':
                                active
                            }
                          )
                        }
                      >
                        {({ active, selected }) => (
                          <>
                            <div className="flex items-center">
                              {v?.status && (
                                <span
                                  className={cls(
                                    'mr-3 inline-block h-2 w-2 shrink-0 rounded-full bg-gray-200 dark:bg-gray-600',
                                    { 'bg-green-400': v.status === 'active' },
                                    {
                                      'bg-green-600':
                                        v.status === 'active' && active
                                    }
                                  )}
                                  aria-hidden="true"
                                />
                              )}
                              <span
                                className={cls(
                                  'truncate text-gray-700 dark:text-gray-200',
                                  {
                                    'font-semibold': selected,
                                    'text-gray-900 dark:text-white': active
                                  }
                                )}
                              >
                                {v?.filterValue}
                              </span>
                              <span
                                className={cls(
                                  'ml-2 truncate text-gray-500 dark:text-gray-400',
                                  {
                                    'text-gray-900 dark:text-white': active
                                  }
                                )}
                              >
                                {v?.displayValue}
                              </span>
                            </div>
                            {selected && (
                              <span
                                className={cls(
                                  'absolute inset-y-0 right-0 flex items-center pr-4 text-violet-600 dark:text-violet-400',
                                  { 'text-white': active }
                                )}
                              >
                                <CheckIcon
                                  className="h-5 w-5"
                                  aria-hidden="true"
                                />
                              </span>
                            )}
                          </>
                        )}
                      </C.Option>
                    ))}
                  {!filteredValues?.length && (
                    <div className="w-full py-2 text-center text-gray-500 dark:text-gray-400">
                      No results found
                    </div>
                  )}
                </C.Options>
              </div>
            )}
          </div>
        )}
      </C>
    </div>
  );
}
