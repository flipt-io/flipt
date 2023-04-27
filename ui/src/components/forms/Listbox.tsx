import { Listbox as L, Transition } from '@headlessui/react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid';
import { Fragment } from 'react';
import { classNames } from '~/utils/helpers';

type ListBoxProps<T extends ISelectable> = {
  id: string;
  name: string;
  placeholder?: string;
  values?: T[];
  selected: T;
  setSelected?: (v: T) => void;
  disabled?: boolean;
  className?: string;
};

export interface ISelectable {
  key: string;
  displayValue: string;
}

export default function Listbox<T extends ISelectable>(props: ListBoxProps<T>) {
  const { id, name, className, values, selected, setSelected, disabled } =
    props;

  return (
    <L
      as="div"
      name={name}
      className={className}
      value={selected}
      by="key"
      onChange={(v: T) => {
        setSelected && setSelected(v);
      }}
      disabled={disabled}
    >
      {({ open }) => (
        <>
          <div className="relative mt-2">
            <L.Button
              className={classNames(
                disabled
                  ? 'bg-gray-100'
                  : 'bg-gray-50 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-1 focus:ring-violet-600',
                'relative w-full cursor-default rounded-md px-2 py-2 pl-3 pr-10 text-left text-gray-900 focus:outline-none sm:text-sm sm:leading-6'
              )}
              id={`${id}-select-button`}
            >
              <div className="flex items-center">
                <span className="block truncate font-medium text-gray-600">
                  {selected?.displayValue}
                </span>
                {!disabled && (
                  <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                    <ChevronUpDownIcon
                      className="h-5 w-5 text-gray-400"
                      aria-hidden="true"
                    />
                  </span>
                )}
              </div>
            </L.Button>

            <Transition
              show={open}
              as={Fragment}
              leave="transition ease-in duration-100"
              leaveFrom="opacity-100"
              leaveTo="opacity-0"
            >
              <L.Options
                className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-gray-50 py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                id={`${id}-select-options`}
              >
                {values?.map((v) => (
                  <L.Option
                    key={v.key}
                    className={({ active }) =>
                      classNames(
                        active ? 'bg-violet-300 text-white' : 'text-gray-900',
                        'relative cursor-default select-none py-2 pl-3 pr-9'
                      )
                    }
                    value={v}
                  >
                    {({ selected, active }) => (
                      <>
                        <span
                          className={classNames(
                            selected ? 'font-semibold' : 'font-normal',
                            'block truncate'
                          )}
                        >
                          {v.displayValue}
                        </span>

                        {selected ? (
                          <span
                            className={classNames(
                              active ? 'text-white' : 'text-violet-600',
                              'absolute inset-y-0 right-0 flex items-center pr-4'
                            )}
                          >
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        ) : null}
                      </>
                    )}
                  </L.Option>
                ))}
              </L.Options>
            </Transition>
          </div>
        </>
      )}
    </L>
  );
}
