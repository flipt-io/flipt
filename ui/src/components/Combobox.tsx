import { Combobox as C } from '@headlessui/react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/24/outline';
import { useField } from 'formik';
import { useState } from 'react';
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

  const filteredValues = values?.filter((v) =>
    v.filterValue.toLowerCase().includes(query.toLowerCase())
  );

  return (
    <C
      as="div"
      className={className}
      value={selected}
      onChange={(v: T | null) => {
        setSelected && setSelected(v);
        field.onChange({ target: { value: v?.key, id } });
      }}
      disabled={disabled}
      nullable
    >
      {({ open }) => (
        <div
          onFocus={() => setTimeout(() => setOpenOptions(true), 100)}
          onBlur={() => setTimeout(() => setOpenOptions(false), 100)}
        >
          <div className="relative flex w-full flex-row">
            <C.Input
              //id={id}
              className={cls(
                'w-full rounded-md border border-gray-300 bg-gray-50 py-2 pl-3 pr-10 text-gray-900 shadow-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500 sm:text-sm',
                inputClassName
              )}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setQuery(e.target.value);
              }}
              displayValue={(v: T) => v?.key}
              placeholder={placeholder}
              name={name}
              id={`${id}-select-input`}
              autoComplete="off"
            />
            <C.Button
              className="absolute -inset-y-0 right-0 items-center rounded-r-md px-2 focus:outline-none"
              id={`${id}-select-button`}
            >
              <ChevronUpDownIcon
                className="h-5 w-5 text-gray-400"
                aria-hidden="true"
              />
            </C.Button>
          </div>
          {open && (
            <C.Options
              className="z-10 mt-1 flex max-h-60 w-full flex-col overflow-auto bg-background py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
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
                          'bg-violet-300': active
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
                                'mr-3 inline-block h-2 w-2 flex-shrink-0 rounded-full bg-gray-200',
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
                            className={cls('truncate text-gray-700', {
                              'font-semibold': selected,
                              'text-gray-100': active
                            })}
                          >
                            {v?.filterValue}
                          </span>
                          <span
                            className={cls('ml-2 truncate text-gray-500', {
                              'text-gray-100': active
                            })}
                          >
                            {v?.displayValue}
                          </span>
                        </div>
                        {selected && (
                          <span
                            className={cls(
                              'absolute inset-y-0 right-0 flex items-center pr-4 text-violet-600',
                              { 'text-white': active }
                            )}
                          >
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        )}
                      </>
                    )}
                  </C.Option>
                ))}
              {!filteredValues?.length && (
                <div className="w-full py-2 text-center text-gray-500">
                  No results found
                </div>
              )}
            </C.Options>
          )}
        </div>
      )}
    </C>
  );
}
