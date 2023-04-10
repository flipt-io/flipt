import { Combobox as C } from '@headlessui/react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/24/outline';
import { useField } from 'formik';
import { useState } from 'react';
import { classNames } from '~/utils/helpers';

type ComboboxProps<T extends ISelectable> = {
  id: string;
  name: string;
  placeholder?: string;
  values?: T[];
  selected: T | null;
  setSelected?: (v: T | null) => void;
  disabled?: boolean;
  className?: string;
};

export interface ISelectable {
  key: string;
  status?: 'active' | 'inactive';
  filterValue: string;
  displayValue: string;
}

export default function Combobox<T extends ISelectable>(
  props: ComboboxProps<T>
) {
  const {
    id,
    className,
    values,
    selected,
    setSelected,
    placeholder,
    disabled
  } = props;

  const [query, setQuery] = useState('');
  const [field] = useField(props);

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
      <div className="relative flex w-full flex-row">
        <C.Input
          id={id}
          className="w-full rounded-md border border-gray-300 bg-white py-2 pl-3 pr-10 shadow-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500 sm:text-sm"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            setQuery(e.target.value);
          }}
          displayValue={(v: T) => v?.key}
          placeholder={placeholder}
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
      <C.Options
        className="z-10 mt-1 flex max-h-60 w-full flex-col overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
        id={`${id}-select-options`}
      >
        {filteredValues &&
          filteredValues.map((v) => (
            <C.Option
              key={v?.key}
              value={v}
              className={({ active }) =>
                classNames(
                  'relative w-full cursor-default select-none py-2 pl-3 pr-9 text-gray-900',
                  active ? 'bg-violet-100' : ''
                )
              }
            >
              {({ active, selected }) => (
                <>
                  <div className="flex items-center">
                    {v?.status && (
                      <span
                        className={classNames(
                          'mr-3 inline-block h-2 w-2 flex-shrink-0 rounded-full',
                          v.status === 'active' ? 'bg-green-400' : 'bg-gray-200'
                        )}
                        aria-hidden="true"
                      />
                    )}
                    <span
                      className={classNames(
                        'truncate',
                        selected ? 'font-semibold' : ''
                      )}
                    >
                      {v?.filterValue}
                    </span>
                    <span className="ml-2 truncate text-gray-500">
                      {v?.displayValue}
                    </span>
                  </div>

                  {selected && (
                    <span
                      className={classNames(
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                        active ? 'text-white' : 'text-violet-600'
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
    </C>
  );
}
